package host

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"okpay/payment/plugin/contract"
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
	kernel      contract.KernelService
	stdout      io.Writer
	stderr      io.Writer
	logger      hclog.Logger

	mu       sync.Mutex
	brokerMu sync.Mutex
	clients  map[string]*clientHolder
	metrics  map[string]*pluginMetrics
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

// WithKernelService sets plugin -> kernel callback implementation for grpc version calls.
func WithKernelService(kernel contract.KernelService) Option {
	return func(m *Manager) {
		m.kernel = kernel
	}
}

// NewManager 创建插件管理器，默认使用 ./plugins 作为存储目录。
func NewManager(opts ...Option) (*Manager, error) {
	mgr := &Manager{
		dir:         pluginsDir,
		callTimeout: CallTimeout,
		stdout:      io.Discard,
		stderr:      io.Discard,
		logger:      hclog.NewNullLogger(),
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

// SetKernelService updates plugin -> kernel callback implementation at runtime.
func (m *Manager) SetKernelService(kernel contract.KernelService) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.kernel = kernel
}

// SaveAndInspect 将上传的二进制写入磁盘（按 info.ID 命名），并返回插件信息，上传相同 ID 时直接覆盖原文件。
func (m *Manager) SaveAndInspect(ctx context.Context, filename string, content []byte) (*contract.PluginInfo, string, error) {
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

// InstallFromPath 从指定路径安装插件（仅临时校验，成功后落盘到插件目录）。
func (m *Manager) InstallFromPath(ctx context.Context, path string) (*contract.PluginInfo, string, error) {
	if strings.TrimSpace(path) == "" {
		return nil, "", fmt.Errorf("插件路径为空")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ensureRegularFile(path); err != nil {
		return nil, "", err
	}
	if strings.TrimSpace(m.dir) == "" {
		return nil, "", fmt.Errorf("插件目录未配置")
	}

	tempFile, err := os.CreateTemp(m.dir, "upload-*")
	if err != nil {
		return nil, "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() { _ = os.Remove(tempPath) }()

	src, err := os.Open(path)
	if err != nil {
		_ = tempFile.Close()
		return nil, "", fmt.Errorf("读取插件文件失败: %w", err)
	}
	if _, err := io.Copy(tempFile, src); err != nil {
		_ = src.Close()
		_ = tempFile.Close()
		return nil, "", fmt.Errorf("写入临时文件失败: %w", err)
	}
	_ = src.Close()
	if err := tempFile.Close(); err != nil {
		return nil, "", fmt.Errorf("写入临时文件失败: %w", err)
	}

	if err := ensurePluginPath(tempPath, m.dir); err != nil {
		return nil, "", err
	}
	info, err := m.Inspect(ctx, tempPath)
	if err != nil {
		return nil, "", fmt.Errorf("检测插件失败: %w", err)
	}
	if err := validatePluginID(info.ID); err != nil {
		return nil, "", fmt.Errorf("插件 ID 非法: %w", err)
	}
	destPath := filepath.Join(m.dir, info.ID)
	if err := ensureInsideDir(destPath, m.dir); err != nil {
		return nil, "", fmt.Errorf("插件路径非法: %w", err)
	}
	if err := os.Rename(tempPath, destPath); err != nil {
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
	if err := os.Remove(destPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("删除插件文件失败: %w", err)
	}
	m.invalidate(id)
	return nil
}

// Inspect 读取指定路径的插件信息（不落盘、不缓存）。
func (m *Manager) Inspect(ctx context.Context, path string) (*contract.PluginInfo, error) {
	return m.InspectRuntime(ctx, path)
}

func wrapCall(funcName string, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("调用 %s 发生 panic: %v", funcName, r)
		}
	}()
	return fn()
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
	client    *plugin.Client
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
