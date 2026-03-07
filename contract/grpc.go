package contract

import (
	"context"
	"fmt"
	"net"
	"sync"

	hplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
	"okpay/payment/plugin/proto"
)

// PluginService defines the strongly-typed protobuf contract between kernel and plugin.
type PluginService interface {
	Info(context.Context, *proto.PluginInfoRequest) (*proto.PluginInfoResponse, error)
	Create(context.Context, *proto.CreateRequest) (*proto.CreateResponse, error)
	Query(context.Context, *proto.QueryRequest) (*proto.QueryResponse, error)
	Refund(context.Context, *proto.RefundRequest) (*proto.RefundResponse, error)
	Transfer(context.Context, *proto.TransferRequest) (*proto.TransferResponse, error)
	Balance(context.Context, *proto.BalanceRequest) (*proto.BalanceResponse, error)
	InvokeFunc(context.Context, *proto.InvokeFuncRequest) (*proto.InvokeFuncResponse, error)
}

// KernelService defines plugin -> kernel callbacks over GRPCBroker.
type KernelService interface {
	CompleteOrder(context.Context, *proto.CompleteOrderRequest) (*proto.Ack, error)
	CompleteRefund(context.Context, *proto.CompleteRefundRequest) (*proto.Ack, error)
	CompleteTransfer(context.Context, *proto.CompleteTransferRequest) (*proto.Ack, error)
	RecordCNotify(context.Context, *proto.RecordCNotifyRequest) (*proto.Ack, error)
	LockOrderExt(context.Context, *proto.LockOrderExtRequest) (*proto.LockOrderExtResponse, error)
}

// GRPCPlugin is the go-plugin gRPC implementation for strongly typed plugin API.
type GRPCPlugin struct {
	hplugin.NetRPCUnsupportedPlugin
	Impl PluginService
}

func (p *GRPCPlugin) GRPCServer(broker *hplugin.GRPCBroker, s *grpc.Server) error {
	if p == nil || p.Impl == nil {
		return fmt.Errorf("plugin implementation is nil")
	}
	proto.RegisterPluginServiceServer(s, &pluginServiceServer{impl: p.Impl, broker: broker})
	return nil
}

func (p *GRPCPlugin) GRPCClient(ctx context.Context, broker *hplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &pluginServiceClient{
		client: proto.NewPluginServiceClient(c),
		broker: broker,
	}, nil
}

var _ hplugin.GRPCPlugin = (*GRPCPlugin)(nil)

type pluginServiceServer struct {
	proto.UnimplementedPluginServiceServer
	impl   PluginService
	broker *hplugin.GRPCBroker
}

func (s *pluginServiceServer) Info(ctx context.Context, in *proto.PluginInfoRequest) (*proto.PluginInfoResponse, error) {
	return s.impl.Info(ctx, in)
}

func (s *pluginServiceServer) Create(ctx context.Context, in *proto.CreateRequest) (*proto.CreateResponse, error) {
	return s.impl.Create(withKernelDialContext(ctx, s.broker, in.GetCtx()), in)
}

func (s *pluginServiceServer) Query(ctx context.Context, in *proto.QueryRequest) (*proto.QueryResponse, error) {
	return s.impl.Query(withKernelDialContext(ctx, s.broker, in.GetCtx()), in)
}

func (s *pluginServiceServer) Refund(ctx context.Context, in *proto.RefundRequest) (*proto.RefundResponse, error) {
	return s.impl.Refund(withKernelDialContext(ctx, s.broker, in.GetCtx()), in)
}

func (s *pluginServiceServer) Transfer(ctx context.Context, in *proto.TransferRequest) (*proto.TransferResponse, error) {
	return s.impl.Transfer(withKernelDialContext(ctx, s.broker, in.GetCtx()), in)
}

func (s *pluginServiceServer) Balance(ctx context.Context, in *proto.BalanceRequest) (*proto.BalanceResponse, error) {
	return s.impl.Balance(withKernelDialContext(ctx, s.broker, in.GetCtx()), in)
}

func (s *pluginServiceServer) InvokeFunc(ctx context.Context, in *proto.InvokeFuncRequest) (*proto.InvokeFuncResponse, error) {
	return s.impl.InvokeFunc(withKernelDialContext(ctx, s.broker, in.GetCtx()), in)
}

type pluginServiceClient struct {
	client proto.PluginServiceClient
	broker *hplugin.GRPCBroker
}

func (c *pluginServiceClient) KernelBroker() *hplugin.GRPCBroker {
	if c == nil {
		return nil
	}
	return c.broker
}

func (c *pluginServiceClient) Info(ctx context.Context, in *proto.PluginInfoRequest) (*proto.PluginInfoResponse, error) {
	return c.client.Info(ctx, in)
}

func (c *pluginServiceClient) Create(ctx context.Context, in *proto.CreateRequest) (*proto.CreateResponse, error) {
	return c.client.Create(ctx, in)
}

func (c *pluginServiceClient) Query(ctx context.Context, in *proto.QueryRequest) (*proto.QueryResponse, error) {
	return c.client.Query(ctx, in)
}

func (c *pluginServiceClient) Refund(ctx context.Context, in *proto.RefundRequest) (*proto.RefundResponse, error) {
	return c.client.Refund(ctx, in)
}

func (c *pluginServiceClient) Transfer(ctx context.Context, in *proto.TransferRequest) (*proto.TransferResponse, error) {
	return c.client.Transfer(ctx, in)
}

func (c *pluginServiceClient) Balance(ctx context.Context, in *proto.BalanceRequest) (*proto.BalanceResponse, error) {
	return c.client.Balance(ctx, in)
}

func (c *pluginServiceClient) InvokeFunc(ctx context.Context, in *proto.InvokeFuncRequest) (*proto.InvokeFuncResponse, error) {
	return c.client.InvokeFunc(ctx, in)
}

// ServeKernelService serves kernel callbacks for plugin side usage via GRPCBroker.
// The returned cleanup function must be called to release broker resources.
func ServeKernelService(broker *hplugin.GRPCBroker, brokerID uint32, impl KernelService) (func(), error) {
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
func DialKernelService(broker *hplugin.GRPCBroker, brokerID uint32) (KernelService, *grpc.ClientConn, error) {
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

func (s *kernelServiceServer) CompleteOrder(ctx context.Context, in *proto.CompleteOrderRequest) (*proto.Ack, error) {
	return s.impl.CompleteOrder(ctx, in)
}

func (s *kernelServiceServer) CompleteRefund(ctx context.Context, in *proto.CompleteRefundRequest) (*proto.Ack, error) {
	return s.impl.CompleteRefund(ctx, in)
}

func (s *kernelServiceServer) CompleteTransfer(ctx context.Context, in *proto.CompleteTransferRequest) (*proto.Ack, error) {
	return s.impl.CompleteTransfer(ctx, in)
}

func (s *kernelServiceServer) RecordCNotify(ctx context.Context, in *proto.RecordCNotifyRequest) (*proto.Ack, error) {
	return s.impl.RecordCNotify(ctx, in)
}

func (s *kernelServiceServer) LockOrderExt(ctx context.Context, in *proto.LockOrderExtRequest) (*proto.LockOrderExtResponse, error) {
	return s.impl.LockOrderExt(ctx, in)
}

type kernelServiceClient struct {
	client proto.KernelServiceClient
}

func (c *kernelServiceClient) CompleteOrder(ctx context.Context, in *proto.CompleteOrderRequest) (*proto.Ack, error) {
	return c.client.CompleteOrder(ctx, in)
}

func (c *kernelServiceClient) CompleteRefund(ctx context.Context, in *proto.CompleteRefundRequest) (*proto.Ack, error) {
	return c.client.CompleteRefund(ctx, in)
}

func (c *kernelServiceClient) CompleteTransfer(ctx context.Context, in *proto.CompleteTransferRequest) (*proto.Ack, error) {
	return c.client.CompleteTransfer(ctx, in)
}

func (c *kernelServiceClient) RecordCNotify(ctx context.Context, in *proto.RecordCNotifyRequest) (*proto.Ack, error) {
	return c.client.RecordCNotify(ctx, in)
}

func (c *kernelServiceClient) LockOrderExt(ctx context.Context, in *proto.LockOrderExtRequest) (*proto.LockOrderExtResponse, error) {
	return c.client.LockOrderExt(ctx, in)
}

type ctxKernelBrokerKey struct{}
type ctxKernelBrokerIDKey struct{}

func withKernelDialContext(ctx context.Context, broker *hplugin.GRPCBroker, invokeCtx *proto.InvokeContext) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if broker == nil || invokeCtx == nil || invokeCtx.GetKernelBrokerId() == 0 {
		return ctx
	}
	ctx = context.WithValue(ctx, ctxKernelBrokerKey{}, broker)
	ctx = context.WithValue(ctx, ctxKernelBrokerIDKey{}, invokeCtx.GetKernelBrokerId())
	return ctx
}

// DialKernelServiceFromContext resolves KernelService from grpc broker metadata injected by host.
func DialKernelServiceFromContext(ctx context.Context) (KernelService, *grpc.ClientConn, error) {
	if ctx == nil {
		return nil, nil, fmt.Errorf("context is nil")
	}
	broker, _ := ctx.Value(ctxKernelBrokerKey{}).(*hplugin.GRPCBroker)
	if broker == nil {
		return nil, nil, fmt.Errorf("kernel broker is unavailable")
	}
	bid, _ := ctx.Value(ctxKernelBrokerIDKey{}).(uint32)
	if bid == 0 {
		return nil, nil, fmt.Errorf("kernel broker id is unavailable")
	}
	return DialKernelService(broker, bid)
}
