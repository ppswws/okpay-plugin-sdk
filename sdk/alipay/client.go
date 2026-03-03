package alipay

import (
	"fmt"
	"strings"

	"github.com/go-pay/gopay/alipay"
)

type ClientConfig struct {
	AppID                   string
	PrivateKey              string
	IsProd                  bool
	NotifyURL               string
	ReturnURL               string
	Charset                 string
	SignType                string
	AppCertPath             string
	AliPayRootCertPath      string
	AliPayPublicCertPath    string
	AppCertContent          []byte
	AliPayRootCertContent   []byte
	AliPayPublicCertContent []byte
	AliPayPublicKey         []byte
}

// NewClient creates a gopay alipay client and applies cert/public-key config.
func NewClient(cfg ClientConfig) (*alipay.Client, error) {
	cfg.AppID = strings.TrimSpace(cfg.AppID)
	cfg.PrivateKey = strings.TrimSpace(cfg.PrivateKey)
	if cfg.AppID == "" || cfg.PrivateKey == "" {
		return nil, fmt.Errorf("支付宝参数不完整")
	}
	client, err := alipay.NewClient(cfg.AppID, cfg.PrivateKey, cfg.IsProd)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.Charset) != "" {
		client.SetCharset(cfg.Charset)
	}
	if strings.TrimSpace(cfg.SignType) != "" {
		client.SetSignType(cfg.SignType)
	}
	if strings.TrimSpace(cfg.NotifyURL) != "" {
		client.SetNotifyUrl(cfg.NotifyURL)
	}
	if strings.TrimSpace(cfg.ReturnURL) != "" {
		client.SetReturnUrl(cfg.ReturnURL)
	}
	if len(cfg.AppCertContent) > 0 && len(cfg.AliPayRootCertContent) > 0 && len(cfg.AliPayPublicCertContent) > 0 {
		if err := client.SetCertSnByContent(cfg.AppCertContent, cfg.AliPayRootCertContent, cfg.AliPayPublicCertContent); err != nil {
			return nil, err
		}
	} else if strings.TrimSpace(cfg.AppCertPath) != "" && strings.TrimSpace(cfg.AliPayRootCertPath) != "" && strings.TrimSpace(cfg.AliPayPublicCertPath) != "" {
		if err := client.SetCertSnByPath(cfg.AppCertPath, cfg.AliPayRootCertPath, cfg.AliPayPublicCertPath); err != nil {
			return nil, err
		}
	}
	if len(cfg.AliPayPublicKey) > 0 {
		client.AutoVerifySign(cfg.AliPayPublicKey)
	}
	return client, nil
}
