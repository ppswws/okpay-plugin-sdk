package sdk

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	hplugin "github.com/hashicorp/go-plugin"
	"okpay/payment/plugin/contract"
)

// Serve starts plugin process with strongly-typed grpc/protobuf contract.
func Serve(impl contract.PluginService) error {
	if impl == nil {
		return fmt.Errorf("version plugin implementation is nil")
	}
	hplugin.Serve(&hplugin.ServeConfig{
		HandshakeConfig: contract.HandshakeConfig,
		Plugins: map[string]hplugin.Plugin{
			contract.PluginName: &contract.GRPCPlugin{Impl: impl},
		},
		GRPCServer: hplugin.DefaultGRPCServer,
		Logger:     hclog.NewNullLogger(),
	})
	return nil
}
