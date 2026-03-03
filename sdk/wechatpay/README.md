# WeChatPay SDK (Go)

微信支付客户端创建封装（V2/V3）。只做 gopay 客户端初始化，不做支付/退款/分账等高层封装。函数均位于 `payment/plugin/sdk/wechatpay`。

## 功能范围

- APIv2 / APIv3 客户端初始化（基于 gopay）
- 支持证书文件或证书内容（V2）
- 支持自动验签或公钥验签（V3）

不包含：
- 下单、退款、分账、转账等业务方法（请直接调用 gopay 客户端）
- 公众号/小程序 OpenID 获取（见 `payment/plugin/wechat`）

## 环境要求

- Go 1.20+

## 使用方法

### 1. APIv2 客户端（支付/退款等 API 由 gopay 提供）

```go
import "okpay/payment/plugin/sdk/wechatpay"

client, err := wechatpay.NewV2Client(wechatpay.V2Config{
  AppID:  appID,
  MchID:  mchID,
  APIKey: apiKey,
  IsProd: true,
  CertPath: certPath,
  KeyPath:  keyPath,
})
```

创建后，直接使用 gopay 的 V2 API（示例仅展示调用方式，参数以 gopay 文档为准）：

```go
resp, err := client.UnifiedOrder(ctx, bm)
```

### 2. APIv3 客户端（支付/退款等 API 由 gopay 提供）

```go
client, err := wechatpay.NewV3Client(wechatpay.V3Config{
  MchID:      mchID,
  SerialNo:   serialNo,
  ApiV3Key:   apiV3Key,
  PrivateKey: privateKey,
  AutoVerify: true,
})
```

创建后，直接使用 gopay 的 V3 API（示例仅展示调用方式，参数以 gopay 文档为准）：

```go
resp, err := client.V3TransactionNative(ctx, bm)
```

## 能力说明

| 函数 | 说明 |
| --- | --- |
| NewV2Client | 创建微信支付 V2 客户端 |
| NewV3Client | 创建微信支付 V3 客户端 |

## 说明

- 本模块只负责初始化客户端；业务能力由 gopay 提供。
- 公众号/小程序的 OpenID 获取在 `payment/plugin/wechat`。
