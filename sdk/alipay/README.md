# Alipay SDK (Go)

支付宝基础能力封装（V2/V3 + OAuth），全部基于 gopay，不做自定义协议实现。函数均位于 `payment/plugin/sdk/alipay`。

## 功能特点

- 支付宝 V2 / V3 客户端初始化
- 支持公钥/证书两种模式
- OAuth 授权、换取 token、刷新 token、用户信息

## 环境要求

- Go 1.20+

## 使用方法

### 1. 创建 V2 客户端

```go
import "okpay/payment/plugin/sdk/alipay"

client, err := alipay.NewClient(alipay.ClientConfig{
  AppID:      appID,
  PrivateKey: privateKey,
  IsProd:     true,
  NotifyURL:  notifyURL,
  ReturnURL:  returnURL,
  SignType:   "RSA2",
  Charset:    "utf-8",
  AppCertPath:        appCertPath,
  AliPayRootCertPath: rootCertPath,
  AliPayPublicCertPath: publicCertPath,
})
```

### 2. 创建 V3 客户端

```go
client, err := alipay.NewV3Client(alipay.V3Config{
  AppID:      appID,
  PrivateKey: privateKey,
  IsProd:     true,
  AppCertPath:        appCertPath,
  AliPayRootCertPath: rootCertPath,
  AliPayPublicCertPath: publicCertPath,
})
```

### 3. OAuth 授权

```go
authURL := alipay.BuildOAuthURL(appID, redirectURL, state)

tokenResp, err := alipay.ExchangeAuthCode(ctx, appID, privateKey, authCode, isProd)
refreshResp, err := alipay.RefreshToken(ctx, appID, privateKey, refreshToken, isProd)
userResp, err := alipay.UserInfo(ctx, appID, privateKey, accessToken, isProd)
```

## 能力说明

| 函数 | 说明 |
| --- | --- |
| NewClient | 创建支付宝 V2 客户端 |
| NewV3Client | 创建支付宝 V3 客户端 |
| BuildOAuthURL | 生成 OAuth2 授权链接 |
| ExchangeAuthCode | 通过 auth_code 换取 token |
| RefreshToken | 刷新 token |
| UserInfo | 获取用户信息 |

## 备注

- 支付宝投诉能力暂不在此目录实现（按需求先不处理）。
- 具体业务 API 由 gopay 客户端直接调用。
