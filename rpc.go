package plugin

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"net/rpc"

	hplugin "github.com/hashicorp/go-plugin"
)

const (
	// PluginName 用于 go-plugin Dispense 的插件键名。
	PluginName = "payment_channel"
)

func init() {
	// 注册 gob 需要透传的具体类型，否则 map[string]any 中的嵌套结构会报未注册错误。
	gob.Register(InputField{})
	gob.Register(map[string]InputField{})
	gob.Register(map[string]*InputField{})
	gob.Register([]InputField{})
	gob.Register([]*InputField{})
}

// HandshakeConfig 约束宿主与插件的握手参数，避免错误进程被当作插件启动。
var HandshakeConfig = hplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "OKPAY_PLUGIN",
	MagicCookieValue: "okpay-payment",
}

var (
	// ErrNoImplementation 当 RPC 服务未准备好时返回。
	ErrNoImplementation = errors.New("plugin: implementation not available")
)

// RPCPlugin 将 PaymentChannel 封装成 go-plugin 可识别的插件类型。
type RPCPlugin struct {
	Impl PaymentChannel
}

// Server 启动 RPC 服务端。
func (p *RPCPlugin) Server(*hplugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Impl: p.Impl}, nil
}

// Client 构造 RPC 客户端代理。
func (p *RPCPlugin) Client(_ *hplugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCClient{client: c}, nil
}

// RPCServer 将插件实现暴露给 net/rpc。
type RPCServer struct {
	Impl PaymentChannel
}

func (s *RPCServer) Invoke(args *InvokeArgs, resp *map[string]any) error {
	if s == nil || s.Impl == nil {
		return ErrNoImplementation
	}
	if args == nil {
		return fmt.Errorf("调用参数为空")
	}
	result, err := s.Impl.Call(context.Background(), args.Func, args.Payload)
	if err != nil {
		return err
	}
	if result != nil && resp != nil {
		*resp = result
	}
	return nil
}

// RPCClient 将 RPC 调用包装为 PaymentChannel。
type RPCClient struct {
	client *rpc.Client
}

func (c *RPCClient) Call(ctx context.Context, funcName string, req *CallRequest) (map[string]any, error) {
	args := &InvokeArgs{Func: funcName, Payload: req}
	var resp map[string]any
	if err := callRPC(ctx, c.client, "Plugin.Invoke", args, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// InvokeArgs 是 RPC 调用参数，需要导出以满足 net/rpc 要求。
type InvokeArgs struct {
	Func    string       `json:"func"`
	Payload *CallRequest `json:"payload"`
}

func callRPC(ctx context.Context, client *rpc.Client, method string, req any, resp any) error {
	if ctx == nil {
		ctx = context.Background()
	}
	done := make(chan error, 1)
	go func() {
		done <- client.Call(method, req, resp)
	}()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		_ = client.Close()
		return ctx.Err()
	}
}
