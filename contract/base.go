package contract

import (
	"errors"

	"github.com/hashicorp/go-plugin"
)

const (
	// PluginName is used by go-plugin Dispense key.
	PluginName = "payment_channel"
)

// HandshakeConfig constrains host/plugin handshake.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  2,
	MagicCookieKey:   "OKPAY_PLUGIN",
	MagicCookieValue: "okpay-payment",
}

var (
	// ErrNoImplementation indicates plugin implementation is not ready.
	ErrNoImplementation = errors.New("插件实现不可用")
)
