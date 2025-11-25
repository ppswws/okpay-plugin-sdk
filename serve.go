package plugin

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"
	hplugin "github.com/hashicorp/go-plugin"
)

const defaultServeCallTimeout = 15 * time.Second

// HandlerFunc 定义插件侧的业务处理函数。
type HandlerFunc func(context.Context, *CallRequest) (map[string]any, error)

// ServeOption 配置 Serve 行为。
type ServeOption func(*serveConfig)

// WithServeCallTimeout 为每次业务调用设置超时（<=0 表示不设置）。
func WithServeCallTimeout(timeout time.Duration) ServeOption {
	return func(cfg *serveConfig) {
		if timeout > 0 {
			cfg.timeout = timeout
		}
	}
}

type serveConfig struct {
	timeout time.Duration
}

// Serve 注册函数映射并启动 go-plugin 服务，业务侧无需关心 RPC/握手细节。
func Serve(funcs map[string]HandlerFunc, opts ...ServeOption) {
	cfg := &serveConfig{timeout: defaultServeCallTimeout}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}
	dispatcher := &handlerDispatcher{
		funcs:   funcs,
		timeout: cfg.timeout,
	}
	hplugin.Serve(&hplugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]hplugin.Plugin{
			PluginName: &RPCPlugin{Impl: dispatcher},
		},
		Logger: hclog.NewNullLogger(), // 静默 go-plugin 内部日志
	})
}

// handlerDispatcher 负责路由和调用注册的业务处理函数。
type handlerDispatcher struct {
	funcs   map[string]HandlerFunc
	timeout time.Duration
}

func (h *handlerDispatcher) Call(ctx context.Context, funcName string, req *CallRequest) (map[string]any, error) {
	if h == nil {
		return nil, fmt.Errorf("处理器未初始化")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	handler, ok := h.funcs[funcName]
	if !ok || handler == nil {
		return nil, fmt.Errorf("未知的函数 %s", funcName)
	}
	if h.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, h.timeout)
		defer cancel()
	}
	return safeCall(ctx, funcName, handler, req)
}

func safeCall(ctx context.Context, funcName string, fn HandlerFunc, req *CallRequest) (resp map[string]any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("调用 %s 时发生 panic: %v", funcName, r)
		}
	}()
	return fn(ctx, req)
}
