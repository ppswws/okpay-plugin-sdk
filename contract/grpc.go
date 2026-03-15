package contract

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/hashicorp/go-plugin"
	"github.com/ppswws/okpay-plugin-sdk/proto"
	"google.golang.org/grpc"
)

// PluginService defines the strongly-typed protobuf contract between kernel and plugin.
type PluginService interface {
	Info(context.Context, *proto.PluginInfoRequest) (*proto.PluginInfoResponse, error)
	Handle(context.Context, *proto.HandleRequest) (*proto.HandleResponse, error)
	Submit(context.Context, *proto.BizRequest) (*proto.BizResult, error)
	Query(context.Context, *proto.BizRequest) (*proto.BizResult, error)
}

// KernelService defines plugin -> kernel callbacks over GRPCBroker.
type KernelService interface {
	CompleteBiz(context.Context, *proto.BizDoneReq) (*proto.Ack, error)
	LockOrderExt(context.Context, *proto.LockExtReq) (*proto.LockExtResp, error)
}

// GRPCPlugin is the go-plugin gRPC implementation for strongly typed plugin API.
type GRPCPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl PluginService
}

func (p *GRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	if p == nil || p.Impl == nil {
		return fmt.Errorf("plugin implementation is nil")
	}
	proto.RegisterPluginServiceServer(s, &pluginServiceServer{impl: p.Impl, broker: broker})
	return nil
}

func (p *GRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &pluginServiceClient{
		client: proto.NewPluginServiceClient(c),
		broker: broker,
	}, nil
}

var _ plugin.GRPCPlugin = (*GRPCPlugin)(nil)

type pluginServiceServer struct {
	proto.UnimplementedPluginServiceServer
	impl   PluginService
	broker *plugin.GRPCBroker
}

func (s *pluginServiceServer) Info(ctx context.Context, in *proto.PluginInfoRequest) (*proto.PluginInfoResponse, error) {
	return s.impl.Info(ctx, in)
}

func (s *pluginServiceServer) Handle(ctx context.Context, in *proto.HandleRequest) (*proto.HandleResponse, error) {
	return s.impl.Handle(withKernelDialContext(ctx, s.broker, in.GetCtx()), in)
}

func (s *pluginServiceServer) Submit(ctx context.Context, in *proto.BizRequest) (*proto.BizResult, error) {
	return s.impl.Submit(withKernelDialContext(ctx, s.broker, in.GetCtx()), in)
}

func (s *pluginServiceServer) Query(ctx context.Context, in *proto.BizRequest) (*proto.BizResult, error) {
	return s.impl.Query(withKernelDialContext(ctx, s.broker, in.GetCtx()), in)
}

type pluginServiceClient struct {
	client proto.PluginServiceClient
	broker *plugin.GRPCBroker
}

func (c *pluginServiceClient) KernelBroker() *plugin.GRPCBroker {
	if c == nil {
		return nil
	}
	return c.broker
}

func (c *pluginServiceClient) Info(ctx context.Context, in *proto.PluginInfoRequest) (*proto.PluginInfoResponse, error) {
	return c.client.Info(ctx, in)
}

func (c *pluginServiceClient) Handle(ctx context.Context, in *proto.HandleRequest) (*proto.HandleResponse, error) {
	return c.client.Handle(ctx, in)
}

func (c *pluginServiceClient) Submit(ctx context.Context, in *proto.BizRequest) (*proto.BizResult, error) {
	return c.client.Submit(ctx, in)
}

func (c *pluginServiceClient) Query(ctx context.Context, in *proto.BizRequest) (*proto.BizResult, error) {
	return c.client.Query(ctx, in)
}

// ServeKernelService serves kernel callbacks for plugin side usage via GRPCBroker.
// The returned cleanup function must be called to release broker resources.
func ServeKernelService(broker *plugin.GRPCBroker, brokerID uint32, impl KernelService) (func(), error) {
	if broker == nil || impl == nil || brokerID == 0 {
		return func() {}, nil
	}
	lis, err := broker.Accept(brokerID)
	if err != nil {
		return nil, err
	}
	server := grpc.NewServer()
	proto.RegisterKernelServiceServer(server, &kernelServiceServer{impl: impl})
	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			server.Stop()
			_ = lis.Close()
		})
	}
	go func(l net.Listener) {
		defer cleanup()
		_ = server.Serve(l)
	}(lis)
	return cleanup, nil
}

// DialKernelService dials kernel callback service from plugin side.
func DialKernelService(broker *plugin.GRPCBroker, brokerID uint32) (KernelService, *grpc.ClientConn, error) {
	if broker == nil || brokerID == 0 {
		return nil, nil, fmt.Errorf("invalid kernel broker")
	}
	conn, err := broker.Dial(brokerID)
	if err != nil {
		return nil, nil, err
	}
	return &kernelServiceClient{client: proto.NewKernelServiceClient(conn)}, conn, nil
}

type kernelServiceServer struct {
	proto.UnimplementedKernelServiceServer
	impl KernelService
}

func (s *kernelServiceServer) CompleteBiz(ctx context.Context, in *proto.BizDoneReq) (*proto.Ack, error) {
	return s.impl.CompleteBiz(ctx, in)
}

func (s *kernelServiceServer) LockOrderExt(ctx context.Context, in *proto.LockExtReq) (*proto.LockExtResp, error) {
	return s.impl.LockOrderExt(ctx, in)
}

type kernelServiceClient struct {
	client proto.KernelServiceClient
}

func (c *kernelServiceClient) CompleteBiz(ctx context.Context, in *proto.BizDoneReq) (*proto.Ack, error) {
	return c.client.CompleteBiz(ctx, in)
}

func (c *kernelServiceClient) LockOrderExt(ctx context.Context, in *proto.LockExtReq) (*proto.LockExtResp, error) {
	return c.client.LockOrderExt(ctx, in)
}

type ctxKernelBrokerKey struct{}
type ctxKernelBrokerIDKey struct{}

func withKernelDialContext(ctx context.Context, broker *plugin.GRPCBroker, invokeCtx *proto.InvokeContext) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if broker == nil || invokeCtx == nil || invokeCtx.GetBrokerId() == 0 {
		return ctx
	}
	ctx = context.WithValue(ctx, ctxKernelBrokerKey{}, broker)
	ctx = context.WithValue(ctx, ctxKernelBrokerIDKey{}, invokeCtx.GetBrokerId())
	return ctx
}

// DialKernelServiceFromContext resolves KernelService from grpc broker metadata injected by host.
func DialKernelServiceFromContext(ctx context.Context) (KernelService, *grpc.ClientConn, error) {
	if ctx == nil {
		return nil, nil, fmt.Errorf("context is nil")
	}
	broker, _ := ctx.Value(ctxKernelBrokerKey{}).(*plugin.GRPCBroker)
	if broker == nil {
		return nil, nil, fmt.Errorf("kernel broker is unavailable")
	}
	bid, _ := ctx.Value(ctxKernelBrokerIDKey{}).(uint32)
	if bid == 0 {
		return nil, nil, fmt.Errorf("kernel broker id is unavailable")
	}
	return DialKernelService(broker, bid)
}
