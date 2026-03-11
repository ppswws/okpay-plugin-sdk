package host

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ppswws/okpay-plugin-sdk/contract"
	"github.com/ppswws/okpay-plugin-sdk/proto"

	"github.com/hashicorp/go-plugin"
)

func (m *Manager) Info(ctx context.Context, id string) (*proto.PluginInfoResponse, error) {
	var out *proto.PluginInfoResponse
	err := m.Invoke(ctx, id, "info", func(ctx context.Context, ch contract.PluginService) error {
		resp, err := ch.Info(ctx, &proto.PluginInfoRequest{})
		if err != nil {
			return err
		}
		out = resp
		return nil
	})
	return out, err
}

func (m *Manager) Create(ctx context.Context, id string, req *proto.CreateRequest) (*proto.CreateResponse, error) {
	var out *proto.CreateResponse
	err := m.Invoke(ctx, id, "create", func(ctx context.Context, ch contract.PluginService) error {
		cleanup, err := m.attachKernelBroker(ch, req.GetCtx())
		if err != nil {
			return err
		}
		defer cleanup()
		resp, err := ch.Create(ctx, req)
		if err != nil {
			return err
		}
		out = resp
		return nil
	})
	return out, err
}

func (m *Manager) Query(ctx context.Context, id string, req *proto.QueryRequest) (*proto.QueryResponse, error) {
	var out *proto.QueryResponse
	err := m.Invoke(ctx, id, "query", func(ctx context.Context, ch contract.PluginService) error {
		cleanup, err := m.attachKernelBroker(ch, req.GetCtx())
		if err != nil {
			return err
		}
		defer cleanup()
		resp, err := ch.Query(ctx, req)
		if err != nil {
			return err
		}
		out = resp
		return nil
	})
	return out, err
}

func (m *Manager) Refund(ctx context.Context, id string, req *proto.RefundRequest) (*proto.RefundResponse, error) {
	var out *proto.RefundResponse
	err := m.Invoke(ctx, id, "refund", func(ctx context.Context, ch contract.PluginService) error {
		cleanup, err := m.attachKernelBroker(ch, req.GetCtx())
		if err != nil {
			return err
		}
		defer cleanup()
		resp, err := ch.Refund(ctx, req)
		if err != nil {
			return err
		}
		out = resp
		return nil
	})
	return out, err
}

func (m *Manager) Transfer(ctx context.Context, id string, req *proto.TransferRequest) (*proto.TransferResponse, error) {
	var out *proto.TransferResponse
	err := m.Invoke(ctx, id, "transfer", func(ctx context.Context, ch contract.PluginService) error {
		cleanup, err := m.attachKernelBroker(ch, req.GetCtx())
		if err != nil {
			return err
		}
		defer cleanup()
		resp, err := ch.Transfer(ctx, req)
		if err != nil {
			return err
		}
		out = resp
		return nil
	})
	return out, err
}

func (m *Manager) Balance(ctx context.Context, id string, req *proto.BalanceRequest) (*proto.BalanceResponse, error) {
	var out *proto.BalanceResponse
	err := m.Invoke(ctx, id, "balance", func(ctx context.Context, ch contract.PluginService) error {
		cleanup, err := m.attachKernelBroker(ch, req.GetCtx())
		if err != nil {
			return err
		}
		defer cleanup()
		resp, err := ch.Balance(ctx, req)
		if err != nil {
			return err
		}
		out = resp
		return nil
	})
	return out, err
}

func (m *Manager) InvokeFunc(ctx context.Context, id string, req *proto.InvokeFuncRequest) (*proto.InvokeFuncResponse, error) {
	var out *proto.InvokeFuncResponse
	err := m.Invoke(ctx, id, "invoke_func", func(ctx context.Context, ch contract.PluginService) error {
		cleanup, err := m.attachKernelBroker(ch, req.GetCtx())
		if err != nil {
			return err
		}
		defer cleanup()
		resp, err := ch.InvokeFunc(ctx, req)
		if err != nil {
			return err
		}
		out = resp
		return nil
	})
	return out, err
}

// InspectRuntime reads plugin info via protobuf/grpc contract.
func (m *Manager) InspectRuntime(ctx context.Context, path string) (*contract.PluginInfo, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if path == "" {
		return nil, fmt.Errorf("插件路径为空")
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
	raw, err := rpcClient.Dispense(contract.PluginName)
	if err != nil {
		return nil, fmt.Errorf("获取插件实例失败: %w", err)
	}
	ch, ok := raw.(contract.PluginService)
	if !ok {
		return nil, fmt.Errorf("插件实例类型不匹配")
	}
	return m.invokeInfo(ctx, ch)
}

// Invoke executes a typed grpc call against plugin.
func (m *Manager) Invoke(ctx context.Context, id, funcName string, call func(context.Context, contract.PluginService) error) error {
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
		release(true)
		return fmt.Errorf("启动插件失败: %w", err)
	}
	raw, err := rpcClient.Dispense(contract.PluginName)
	if err != nil {
		release(true)
		return fmt.Errorf("获取插件实例失败: %w", err)
	}
	ch, ok := raw.(contract.PluginService)
	if !ok {
		release(true)
		return fmt.Errorf("插件实例类型不匹配")
	}
	start := time.Now()
	err = wrapCall(funcName, func() error {
		return call(ctx, ch)
	})
	duration := time.Since(start)
	if m.observer != nil {
		m.observer(id, funcName, duration, err)
	}
	m.recordMetrics(id, duration)
	if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
		release(true)
		return ctx.Err()
	}
	return err
}

func (m *Manager) newClient(path string) (*plugin.Client, error) {
	if err := ensurePluginPath(path, m.dir); err != nil {
		return nil, err
	}
	cfg := &plugin.ClientConfig{
		HandshakeConfig:     contract.HandshakeConfig,
		Plugins:             map[string]plugin.Plugin{contract.PluginName: &contract.GRPCPlugin{}},
		Cmd:                 exec.Command(path),
		Managed:             true,
		AllowedProtocols:    []plugin.Protocol{plugin.ProtocolGRPC},
		GRPCBrokerMultiplex: true,
		SyncStdout:          m.stdout,
		SyncStderr:          m.stderr,
		Stderr:              m.stderr,
		Logger:              m.logger,
		StartTimeout:        2 * time.Second,
	}
	client := plugin.NewClient(cfg)
	if _, err := client.Client(); err != nil {
		client.Kill()
		return nil, fmt.Errorf("启动插件失败: %w", err)
	}
	return client, nil
}

func (m *Manager) getClient(id, path string) (*plugin.Client, func(bool), error) {
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

func (m *Manager) invokeInfo(ctx context.Context, ch contract.PluginService) (*contract.PluginInfo, error) {
	resp, err := ch.Info(ctx, &proto.PluginInfoRequest{})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("info 返回为空")
	}
	info := &contract.PluginInfo{
		ID:         resp.Id,
		Name:       resp.Name,
		Link:       resp.Link,
		Paytypes:   append([]string(nil), resp.Paytypes...),
		Transtypes: append([]string(nil), resp.Transtypes...),
		Note:       resp.Note,
		Inputs:     map[string]contract.InputField{},
	}
	for key, val := range resp.Inputs {
		info.Inputs[key] = contract.InputField{
			Name:     val.Name,
			Type:     val.Type,
			Note:     val.Note,
			Required: val.Required,
			Default:  val.DefaultValue,
			Options:  val.Options,
		}
	}
	if info.ID == "" || info.Name == "" {
		return nil, fmt.Errorf("info 缺少 id/name")
	}
	if len(info.Paytypes) == 0 {
		return nil, fmt.Errorf("info 缺少 paytypes")
	}
	return info, nil
}

func (m *Manager) attachKernelBroker(ch contract.PluginService, invokeCtx *proto.InvokeContext) (func(), error) {
	if m == nil || invokeCtx == nil {
		return func() {}, nil
	}
	m.mu.Lock()
	kernel := m.kernel
	m.mu.Unlock()
	if kernel == nil {
		return func() {}, nil
	}
	withBroker, ok := ch.(interface{ KernelBroker() *plugin.GRPCBroker })
	if !ok || withBroker.KernelBroker() == nil {
		return nil, fmt.Errorf("插件未提供 grpc broker")
	}
	// go-plugin grpc broker multiplex requires sequential Accept/Dial setup.
	m.brokerMu.Lock()
	broker := withBroker.KernelBroker()
	brokerID := broker.NextId()
	cleanup, err := contract.ServeKernelService(broker, brokerID, kernel)
	if err != nil {
		m.brokerMu.Unlock()
		return nil, err
	}
	invokeCtx.KernelBrokerId = brokerID
	return func() {
		cleanup()
		m.brokerMu.Unlock()
	}, nil
}
