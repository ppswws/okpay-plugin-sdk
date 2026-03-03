package plugin

import (
	"io"
	"time"

	"github.com/hashicorp/go-hclog"
	"okpay/payment/plugin/host"
)

func NewManager(opts ...Option) (*Manager, error) {
	return host.NewManager(opts...)
}

func WithDir(dir string) Option {
	return host.WithDir(dir)
}

func WithCallTimeout(timeout time.Duration) Option {
	return host.WithCallTimeout(timeout)
}

func WithPluginLogWriters(stdout, stderr io.Writer) Option {
	return host.WithPluginLogWriters(stdout, stderr)
}

func WithCallObserver(observer CallObserver) Option {
	return host.WithCallObserver(observer)
}

func WithPluginLogger(logger hclog.Logger) Option {
	return host.WithPluginLogger(logger)
}
