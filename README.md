# OKPay Plugin SDK（Typed gRPC）

`okpay-plugin-sdk` 是 OKPay 的插件开发 SDK，提供：
- 插件协议与 proto 定义（`PluginService` / `KernelService`）
- 插件宿主与服务启动封装
- 常用响应、回调、锁单、终端判断、OAuth/小程序辅助能力

## 安装

```bash
go get github.com/ppswws/okpay-plugin-sdk
```

## 最小示例

```go
package main

import (
	"context"
	"log"

	"github.com/ppswws/okpay-plugin-sdk"
	"github.com/ppswws/okpay-plugin-sdk/proto"
)

type channelService struct{}

func (s *channelService) Info(ctx context.Context, _ *proto.PluginInfoRequest) (*proto.PluginInfoResponse, error) {
	return plugin.BuildInfoResponse(plugin.Manifest{
		ID:       "demo",
		Name:     "Demo",
		Paytypes: []string{"alipay"},
	}), nil
}

func (s *channelService) Handle(ctx context.Context, _ *proto.HandleRequest) (*proto.HandleResponse, error) {
	return &proto.HandleResponse{Page: plugin.RespError("not implemented")}, nil
}
func (s *channelService) Submit(context.Context, *proto.BizRequest) (*proto.BizResult, error) {
	return nil, plugin.ErrNoImplementation
}
func (s *channelService) Query(context.Context, *proto.BizRequest) (*proto.BizResult, error) {
	return nil, plugin.ErrNoImplementation
}

func main() {
	if err := plugin.Serve(&channelService{}); err != nil {
		log.Fatal(err)
	}
}
```

## 开发要点

- 使用 typed proto message，不使用 `map[string]any` 动态协议。
- 金额统一使用分（`int64`）或十进制字符串。
- 验签优先使用原始载荷（`raw_http.body_raw`、`raw_http.query_raw`）。
- 页面返回类型仅使用：`jump/html/json/page/error`。

## 常用 API

- `plugin.CreateWithHandlers`
- `plugin.CompleteBiz` / `plugin.LockOrderExt`
- `plugin.ResultOK` / `plugin.ResultPending` / `plugin.ResultFail` / `plugin.ResultBal`
- `plugin.IsWeChat` / `plugin.IsAlipay` / `plugin.IsMobileQQ` / `plugin.IsMobile`
- `plugin.BuildMPOAuthURL` / `plugin.GetMPOpenid` / `plugin.GetMiniOpenid` / `plugin.GetMiniScheme`

## Proto 生成

在 `proto` 目录执行：

```bash
buf generate
```
