package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/rpc"

	hplugin "github.com/hashicorp/go-plugin"
)

const (
	// PluginName 用于 go-plugin Dispense 的插件键名。
	PluginName = "payment_channel"
)

// HandshakeConfig 约束宿主与插件的握手参数，避免错误进程被当作插件启动。
var HandshakeConfig = hplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "OKPAY_PLUGIN",
	MagicCookieValue: "okpay-payment",
}

var (
	// ErrNoImplementation 当 RPC 服务未准备好时返回。
	ErrNoImplementation = errors.New("插件实现不可用")
)

// RPCPlugin 将 PaymentChannel 封装成 go-plugin 可识别的插件类型。
type RPCPlugin struct {
	Impl PaymentChannel
}

// Server 启动 RPC 服务端。
func (p *RPCPlugin) Server(b *hplugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Impl: p.Impl}, nil
}

// Client 构造 RPC 客户端代理。
func (p *RPCPlugin) Client(b *hplugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCClient{client: c}, nil
}

// RPCServer 将插件实现暴露给 net/rpc。
type RPCServer struct {
	Impl PaymentChannel
}

func (s *RPCServer) Invoke(args *InvokeArgs, resp *[]byte) error {
	if s == nil || s.Impl == nil {
		return ErrNoImplementation
	}
	if args == nil {
		return fmt.Errorf("调用参数为空")
	}
	req, err := decodeCallRequest(args.Payload)
	if err != nil {
		return err
	}
	ctx := context.Background()
	result, err := s.Impl.Call(ctx, args.Func, req)
	if err != nil {
		return err
	}
	if resp == nil {
		return nil
	}
	if result == nil {
		*resp = []byte("{}")
		return nil
	}
	out, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("序列化响应失败: %w", err)
	}
	*resp = out
	return nil
}

// RPCClient 将 RPC 调用包装为 PaymentChannel。
type RPCClient struct {
	client *rpc.Client
}

func (c *RPCClient) Call(ctx context.Context, funcName string, req *CallRequest) (map[string]any, error) {
	payload, err := encodeCallRequest(req)
	if err != nil {
		return nil, err
	}
	args := &InvokeArgs{Func: funcName, Payload: payload}
	var resp []byte
	if err := callRPC(ctx, c.client, "Plugin.Invoke", args, &resp); err != nil {
		return nil, err
	}
	return decodeCallResponse(resp)
}

// InvokeArgs 是 RPC 调用参数，需要导出以满足 net/rpc 要求。
type InvokeArgs struct {
	Func    string `json:"func"`
	Payload []byte `json:"payload"`
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

func decodeCallResponse(payload []byte) (map[string]any, error) {
	if len(payload) == 0 {
		return map[string]any{}, nil
	}
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.UseNumber()
	var out map[string]any
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

func encodeCallRequest(req *CallRequest) ([]byte, error) {
	if req == nil {
		return []byte("null"), nil
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}
	return payload, nil
}

func decodeCallRequest(payload []byte) (*CallRequest, error) {
	if len(payload) == 0 || string(payload) == "null" {
		return &CallRequest{}, nil
	}
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.UseNumber()
	var req CallRequest
	if err := dec.Decode(&req); err != nil {
		return nil, fmt.Errorf("解析请求失败: %w", err)
	}
	return &req, nil
}
