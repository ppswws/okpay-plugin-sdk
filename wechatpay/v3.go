package wechatpay

import (
	"fmt"
	"strings"

	v3 "github.com/go-pay/gopay/wechat/v3"
)

type V3Config struct {
	MchID       string
	SerialNo    string
	ApiV3Key    string
	PrivateKey  string
	AutoVerify  bool
	WXPublicKey []byte
	WXPublicSN  string
}

// NewV3Client creates a gopay wechat v3 client and configures auto verify when requested.
func NewV3Client(cfg V3Config) (*v3.ClientV3, error) {
	cfg.MchID = strings.TrimSpace(cfg.MchID)
	cfg.SerialNo = strings.TrimSpace(cfg.SerialNo)
	cfg.ApiV3Key = strings.TrimSpace(cfg.ApiV3Key)
	cfg.PrivateKey = strings.TrimSpace(cfg.PrivateKey)
	if cfg.MchID == "" || cfg.SerialNo == "" || cfg.ApiV3Key == "" || cfg.PrivateKey == "" {
		return nil, fmt.Errorf("微信支付V3参数不完整")
	}
	client, err := v3.NewClientV3(cfg.MchID, cfg.SerialNo, cfg.ApiV3Key, cfg.PrivateKey)
	if err != nil {
		return nil, err
	}
	if len(cfg.WXPublicKey) > 0 && strings.TrimSpace(cfg.WXPublicSN) != "" {
		if err := client.AutoVerifySignByPublicKey(cfg.WXPublicKey, strings.TrimSpace(cfg.WXPublicSN)); err != nil {
			return nil, err
		}
	} else if cfg.AutoVerify {
		if err := client.AutoVerifySign(); err != nil {
			return nil, err
		}
	}
	return client, nil
}
