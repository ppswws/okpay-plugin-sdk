package sdk

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/ppswws/okpay-plugin-sdk/contract"
)

// Serve starts plugin process with strongly-typed grpc/protobuf contract.
func Serve(impl contract.PluginService) error {
	if impl == nil {
		return fmt.Errorf("version plugin implementation is nil")
	}
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: contract.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			contract.PluginName: &contract.GRPCPlugin{Impl: impl},
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     hclog.NewNullLogger(),
	})
	return nil
}
