package alipay

import (
	"fmt"
	"os"
	"strings"

	v3 "github.com/go-pay/gopay/alipay/v3"
)

type V3Config struct {
	AppID                   string
	PrivateKey              string
	IsProd                  bool
	AppCertPath             string
	AliPayRootCertPath      string
	AliPayPublicCertPath    string
	AppCertContent          []byte
	AliPayRootCertContent   []byte
	AliPayPublicCertContent []byte
}

// NewV3Client creates a gopay alipay v3 client and applies cert/public-key config.
func NewV3Client(cfg V3Config) (*v3.ClientV3, error) {
	cfg.AppID = strings.TrimSpace(cfg.AppID)
	cfg.PrivateKey = strings.TrimSpace(cfg.PrivateKey)
	if cfg.AppID == "" || cfg.PrivateKey == "" {
		return nil, fmt.Errorf("支付宝V3参数不完整")
	}
	client, err := v3.NewClientV3(cfg.AppID, cfg.PrivateKey, cfg.IsProd)
	if err != nil {
		return nil, err
	}
	appCert, rootCert, publicCert, err := readCerts(cfg)
	if err != nil {
		return nil, err
	}
	if len(appCert) > 0 && len(rootCert) > 0 && len(publicCert) > 0 {
		if err := client.SetCert(appCert, rootCert, publicCert); err != nil {
			return nil, err
		}
	}
	return client, nil
}

func readCerts(cfg V3Config) ([]byte, []byte, []byte, error) {
	if len(cfg.AppCertContent) > 0 || len(cfg.AliPayRootCertContent) > 0 || len(cfg.AliPayPublicCertContent) > 0 {
		return cfg.AppCertContent, cfg.AliPayRootCertContent, cfg.AliPayPublicCertContent, nil
	}
	if strings.TrimSpace(cfg.AppCertPath) == "" || strings.TrimSpace(cfg.AliPayRootCertPath) == "" || strings.TrimSpace(cfg.AliPayPublicCertPath) == "" {
		return nil, nil, nil, nil
	}
	appCert, err := os.ReadFile(cfg.AppCertPath)
	if err != nil {
		return nil, nil, nil, err
	}
	rootCert, err := os.ReadFile(cfg.AliPayRootCertPath)
	if err != nil {
		return nil, nil, nil, err
	}
	publicCert, err := os.ReadFile(cfg.AliPayPublicCertPath)
	if err != nil {
		return nil, nil, nil, err
	}
	return appCert, rootCert, publicCert, nil
}
