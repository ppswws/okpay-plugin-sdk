package plugin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	hplugin "github.com/hashicorp/go-plugin"
)

const (
	CallTimeout = 20 * time.Second
	pluginsDir  = "plugins"
)

// Manager: 单插件单进程常驻，按需重启。
type Manager struct {
	dir         string
	callTimeout time.Duration
	observer    CallObserver
	stdout      io.Writer
	stderr      io.Writer
	logger      hclog.Logger

	mu      sync.Mutex
	clients map[string]*clientHolder
	metrics map[string]*pluginMetrics
}

// Option 配置 Manager。
type Option func(*Manager)

// CallObserver 用于记录每次插件调用的耗时/结果，便于指标与日志落地。
type CallObserver func(pluginID, funcName string, duration time.Duration, err error)

// WithDir 覆盖插件存放目录。
func WithDir(dir string) Option {
	return func(m *Manager) {
		if strings.TrimSpace(dir) != "" {
			m.dir = dir
		}
	}
}

// WithCallTimeout 配置 RPC 调用超时时间。
func WithCallTimeout(timeout time.Duration) Option {
	return func(m *Manager) {
		if timeout > 0 {
			m.callTimeout = timeout
		}
	}
}

// WithPluginLogWriters 重定向插件进程的 stdout/stderr。
func WithPluginLogWriters(stdout, stderr io.Writer) Option {
	return func(m *Manager) {
		if stdout != nil {
			m.stdout = stdout
		}
		if stderr != nil {
			m.stderr = stderr
		}
	}
}

// WithCallObserver 注入调用观测回调，便于埋点/日志。
func WithCallObserver(observer CallObserver) Option {
	return func(m *Manager) {
		if observer != nil {
			m.observer = observer
		}
	}
}

// WithPluginLogger 配置 go-plugin 的内部日志（默认静默）。
func WithPluginLogger(logger hclog.Logger) Option {
	return func(m *Manager) {
		if logger != nil {
			m.logger = logger
		}
	}
}

// NewManager 创建插件管理器，默认使用 ./plugins 作为存储目录。
func NewManager(opts ...Option) (*Manager, error) {
	mgr := &Manager{
		dir:         pluginsDir,
		callTimeout: CallTimeout,
		stdout:      io.Discard,
		stderr:      io.Discard,
		clients:     make(map[string]*clientHolder),
		metrics:     make(map[string]*pluginMetrics),
	}
	for _, opt := range opts {
		opt(mgr)
	}
	absDir, err := filepath.Abs(mgr.dir)
	if err != nil {
		return nil, fmt.Errorf("无法解析插件目录: %w", err)
	}
	mgr.dir = absDir
	if err := os.MkdirAll(mgr.dir, 0o755); err != nil {
		return nil, fmt.Errorf("创建插件目录失败: %w", err)
	}
	return mgr, nil
}

// SaveAndInspect 将上传的二进制写入磁盘（按 info.ID 命名），并返回插件信息，上传相同 ID 时直接覆盖原文件。
func (m *Manager) SaveAndInspect(ctx context.Context, filename string, content []byte) (*PluginInfo, string, error) {
	if len(content) == 0 {
		return nil, "", fmt.Errorf("上传内容为空: %w", errors.New("content empty"))
	}
	if strings.Contains(filename, string(os.PathSeparator)) {
		return nil, "", fmt.Errorf("文件名非法")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	tempFile, err := os.CreateTemp(m.dir, "upload-*")
	if err != nil {
		return nil, "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	tempPath := tempFile.Name()
	_ = tempFile.Close()

	if err := os.WriteFile(tempPath, content, 0o750); err != nil {
		_ = os.Remove(tempPath)
		return nil, "", fmt.Errorf("写入临时文件失败: %w", err)
	}
	info, err := m.Inspect(ctx, tempPath)
	if err != nil {
		_ = os.Remove(tempPath)
		return nil, "", fmt.Errorf("检测插件失败: %w", err)
	}
	if err := validatePluginID(info.ID); err != nil {
		_ = os.Remove(tempPath)
		return nil, "", fmt.Errorf("插件 ID 非法: %w", err)
	}
	destPath := filepath.Join(m.dir, info.ID)
	if err := ensureInsideDir(destPath, m.dir); err != nil {
		_ = os.Remove(tempPath)
		return nil, "", fmt.Errorf("插件路径非法: %w", err)
	}
	if err := os.Rename(tempPath, destPath); err != nil {
		_ = os.Remove(tempPath)
		return nil, "", fmt.Errorf("覆盖插件文件失败: %w", err)
	}
	m.invalidate(info.ID)
	return info, destPath, nil
}

// Remove 删除已注册的插件文件并关闭对应的进程复用客户端。
func (m *Manager) Remove(id string) error {
	if err := validatePluginID(id); err != nil {
		return err
	}
	if strings.TrimSpace(m.dir) == "" {
		return fmt.Errorf("插件目录未配置")
	}
	destPath := filepath.Join(m.dir, id)
	if err := ensureInsideDir(destPath, m.dir); err != nil {
		return err
	}
	// 删除文件，即便不存在也不视为致命错误
	if err := os.Remove(destPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("删除插件文件失败: %w", err)
	}
	m.invalidate(id)
	return nil
}

// Inspect 读取指定路径的插件信息（不落盘、不缓存）。
func (m *Manager) Inspect(ctx context.Context, path string) (*PluginInfo, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("插件路径为空")
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("无法访问插件文件: %w", err)
	}
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("路径 %s 是目录", path)
	}
	ctx, cancel := context.WithTimeout(ctx, m.callTimeout)
	defer cancel()

	if err := ensurePluginPath(path, m.dir); err != nil {
		return nil, err
	}
	client, err := m.newClient(path)
	if err != nil {
		return nil, err
	}
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("启动插件失败: %w", err)
	}

	raw, err := rpcClient.Dispense(PluginName)
	if err != nil {
		return nil, fmt.Errorf("获取插件实例失败: %w", err)
	}
	channel, ok := raw.(PaymentChannel)
	if !ok {
		return nil, fmt.Errorf("插件实例类型不匹配")
	}
	info, err := m.invokeInfo(ctx, channel)
	if err != nil {
		return nil, fmt.Errorf("获取info失败: %w", err)
	}
	return info, nil
}

// Call 直接调用指定函数名并返回结果，免去手写 Invoke 包装。
func (m *Manager) Call(ctx context.Context, id, funcName string, req *CallRequest) (map[string]any, error) {
	// 对 InvokeFunc 的薄包装，处理返回值收集。
	var resp map[string]any
	err := m.InvokeFunc(ctx, id, funcName, func(ctx context.Context, ch PaymentChannel) error {
		out, err := ch.Call(ctx, funcName, req)
		if err != nil {
			return err
		}
		resp = out
		return nil
	})
	return resp, err
}

// InvokeFunc 执行带函数名的插件操作。
func (m *Manager) InvokeFunc(ctx context.Context, id, funcName string, call func(context.Context, PaymentChannel) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := validatePluginID(id); err != nil {
		return err
	}
	pluginPath := filepath.Join(m.dir, id)
	ctx, cancel := m.applyTimeout(ctx)
	defer cancel()

	client, release, err := m.getClient(id, pluginPath)
	if err != nil {
		return err
	}
	defer release(false)

	rpcClient, err := client.Client()
	if err != nil {
		return fmt.Errorf("启动插件失败: %w", err)
	}

	raw, err := rpcClient.Dispense(PluginName)
	if err != nil {
		return fmt.Errorf("获取插件实例失败: %w", err)
	}
	channel, ok := raw.(PaymentChannel)
	if !ok {
		return fmt.Errorf("插件实例类型不匹配")
	}
	start := time.Now()
	err = wrapCall(funcName, func() error {
		return call(ctx, channel)
	})
	duration := time.Since(start)
	if m.observer != nil {
		m.observer(id, funcName, duration, err)
	}
	m.recordMetrics(id, duration)
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		release(true)
		return err
	}
	return err
}

func wrapCall(funcName string, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("调用 %s 发生 panic: %v", funcName, r)
		}
	}()
	return fn()
}

func (m *Manager) newClient(path string) (*hplugin.Client, error) {
	if err := ensurePluginPath(path, m.dir); err != nil {
		return nil, err
	}
	cfg := &hplugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig,
		Plugins:          map[string]hplugin.Plugin{PluginName: &RPCPlugin{}},
		Cmd:              exec.Command(path),
		AllowedProtocols: []hplugin.Protocol{hplugin.ProtocolNetRPC},
		Stderr:           m.stderr,
		Logger:           m.logger,
	}
	if cfg.Logger == nil {
		cfg.Logger = hclog.NewNullLogger()
	}
	client := hplugin.NewClient(cfg)
	if _, err := client.Client(); err != nil {
		client.Kill()
		return nil, fmt.Errorf("启动插件失败: %w", err)
	}
	return client, nil
}

func (m *Manager) getClient(id, path string) (*hplugin.Client, func(bool), error) {
	if err := ensurePluginPath(path, m.dir); err != nil {
		return nil, nil, err
	}

	m.mu.Lock()
	holder := m.clients[id]
	if holder != nil {
		if holder.client == nil || holder.client.Exited() {
			if holder.client != nil {
				holder.client.Kill()
			}
			holder = nil
		} else {
			holder.lastUsed = time.Now()
			client := holder.client
			m.mu.Unlock()
			return client, func(bool) {}, nil
		}
	}
	m.mu.Unlock()

	client, err := m.newClient(path)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now()
	newHolder := &clientHolder{client: client, createdAt: now, lastUsed: now}
	m.mu.Lock()
	// 若并发下已有存活的，则复用已有的，关闭新建的。
	if exist := m.clients[id]; exist != nil && exist.client != nil && !exist.client.Exited() {
		client.Kill()
		client = exist.client
		exist.lastUsed = time.Now()
	} else {
		m.clients[id] = newHolder
	}
	m.mu.Unlock()

	release := func(drop bool) {
		if !drop {
			return
		}
		m.mu.Lock()
		defer m.mu.Unlock()
		if cur := m.clients[id]; cur != nil && cur.client == client {
			client.Kill()
			delete(m.clients, id)
		}
	}
	return client, release, nil
}

func (m *Manager) recordMetrics(id string, d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	mt := m.metrics[id]
	if mt == nil {
		mt = &pluginMetrics{}
		m.metrics[id] = mt
	}
	mt.calls++
	mt.total += d
}

func (m *Manager) invokeInfo(ctx context.Context, channel PaymentChannel) (*PluginInfo, error) {
	if channel == nil {
		return nil, fmt.Errorf("实例为空")
	}
	result, err := channel.Call(ctx, "info", &CallRequest{})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("info 返回为空")
	}
	info := &PluginInfo{}
	if id, ok := result["id"].(string); ok {
		info.ID = id
	}
	if name, ok := result["name"].(string); ok {
		info.Name = name
	}
	if link, ok := result["link"].(string); ok {
		info.Link = link
	}
	if types, ok := result["paytypes"].([]string); ok {
		info.Paytypes = types
	}
	if typesAny, ok := result["paytypes"].([]any); ok {
		info.Paytypes = toStringSlice(typesAny)
	}
	if transtypes, ok := result["transtypes"].([]string); ok {
		info.Transtypes = transtypes
	}
	if transtypesAny, ok := result["transtypes"].([]any); ok {
		info.Transtypes = toStringSlice(transtypesAny)
	}
	if inputs, ok := result["inputs"].(map[string]InputField); ok {
		info.Inputs = inputs
	} else if rawInputs, ok := result["inputs"].(map[string]any); ok {
		info.Inputs = make(map[string]InputField, len(rawInputs))
		for key, val := range rawInputs {
			if m := toInputField(val); m != nil {
				info.Inputs[key] = *m
			}
		}
	}
	if note, ok := result["note"].(string); ok {
		info.Note = note
	}
	if info.ID == "" || info.Name == "" {
		return nil, fmt.Errorf("info 缺少 id/name")
	}
	return info, nil
}

// Close 关闭所有复用的插件进程。
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, holder := range m.clients {
		if holder != nil && holder.client != nil {
			holder.client.Kill()
		}
		delete(m.clients, id)
	}
}

// invalidate 清理指定插件的复用客户端（如覆盖更新时）。
func (m *Manager) invalidate(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if holder, ok := m.clients[id]; ok && holder != nil {
		if holder.client != nil {
			holder.client.Kill()
		}
		delete(m.clients, id)
	}
}

// Stats 返回当前所有插件的运行状态与调用指标（懒加载，未运行的也列出）。
func (m *Manager) Stats() []PluginStat {
	out := make([]PluginStat, 0)
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, holder := range m.clients {
		alive := holder != nil && holder.client != nil && !holder.client.Exited()
		if !alive {
			continue
		}
		out = append(out, buildStat(id, holder, m.metrics[id]))
	}
	return out
}

func (m *Manager) applyTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if m.callTimeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, m.callTimeout)
}

type clientHolder struct {
	client    *hplugin.Client
	createdAt time.Time
	lastUsed  time.Time
}

// PluginStat 描述单个插件的当前状态和简要指标。
type PluginStat struct {
	ID      string        `json:"id"`
	Run     bool          `json:"run"`
	Created time.Time     `json:"created,omitempty"`
	Last    time.Time     `json:"last,omitempty"`
	Calls   int64         `json:"calls,omitempty"`
	Avg     time.Duration `json:"avg,omitempty"`
}

type pluginMetrics struct {
	calls int64
	total time.Duration
}

func buildStat(id string, holder *clientHolder, mt *pluginMetrics) PluginStat {
	stat := PluginStat{ID: id}
	stat.Run = holder != nil && holder.client != nil && !holder.client.Exited()
	if stat.Run {
		stat.Created = holder.createdAt
		stat.Last = holder.lastUsed
	}
	if mt != nil && mt.calls > 0 {
		stat.Calls = mt.calls
		stat.Avg = mt.total / time.Duration(mt.calls)
	}
	return stat
}
