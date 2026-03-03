package wechatpay

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/go-pay/gopay/pkg/xhttp"
	"github.com/go-pay/gopay/wechat"
)

type V2Config struct {
	AppID     string
	MchID     string
	APIKey    string
	IsProd    bool
	CertPath  string
	KeyPath   string
	CertPEM   []byte
	KeyPEM    []byte
	NotifyURL string
}

// NewV2Client creates a gopay wechat v2 client with optional TLS certs.
func NewV2Client(cfg V2Config) (*wechat.Client, error) {
	cfg.AppID = strings.TrimSpace(cfg.AppID)
	cfg.MchID = strings.TrimSpace(cfg.MchID)
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	if cfg.AppID == "" || cfg.MchID == "" || cfg.APIKey == "" {
		return nil, fmt.Errorf("微信支付V2参数不完整")
	}
	client := wechat.NewClient(cfg.AppID, cfg.MchID, cfg.APIKey, cfg.IsProd)
	if tlsConfig, ok, err := buildTLSConfig(cfg); err != nil {
		return nil, err
	} else if ok {
		tlsClient := xhttp.NewClient().SetHttpTLSConfig(tlsConfig)
		client.SetTLSHttpClient(tlsClient)
	}
	return client, nil
}

func buildTLSConfig(cfg V2Config) (*tls.Config, bool, error) {
	if len(cfg.CertPEM) > 0 && len(cfg.KeyPEM) > 0 {
		pair, err := tls.X509KeyPair(cfg.CertPEM, cfg.KeyPEM)
		if err != nil {
			return nil, false, err
		}
		return &tls.Config{Certificates: []tls.Certificate{pair}}, true, nil
	}
	if strings.TrimSpace(cfg.CertPath) != "" && strings.TrimSpace(cfg.KeyPath) != "" {
		pair, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
		if err != nil {
			return nil, false, err
		}
		return &tls.Config{Certificates: []tls.Certificate{pair}}, true, nil
	}
	return nil, false, nil
}
