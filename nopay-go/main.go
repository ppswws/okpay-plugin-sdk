package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"okpay/payment/plugin"
)

func main() {
	plugin.Serve(map[string]plugin.HandlerFunc{
		"info":     info,
		"create":   create,
		"query":    query,
		"refund":   refund,
		"transfer": transfer,
		"notify":   notify,
		"return":   ret,
	})
}

func info(ctx context.Context, req *plugin.CallRequest) (map[string]any, error) {
	return map[string]any{
		"id":   "nopay",
		"name": "NoPay 演示插件",
		"link": "",
		"paytypes": []string{
			"alipay",
			"wxpay",
		},
		"transtypes": []string{
			"pay",
			"refund",
		},
		"inputs": map[string]plugin.InputField{
			"appurl": {
				Name: "接口地址",
				Type: "input",
				Note: "必须以 http:// 或 https:// 开头，以 / 结尾",
			},
			"appid": {
				Name: "商户ID",
				Type: "input",
				Note: "",
			},
			"appkey": {
				Name: "商户密钥",
				Type: "input",
				Note: "",
			},
		},
		"note": "演示插件，仅返回固定二维码",
	}, nil
}

func create(ctx context.Context, req *plugin.CallRequest) (map[string]any, error) {
	if req == nil || req.Order == nil {
		return nil, fmt.Errorf("请求为空")
	}
	codeURL := "https://baidu.com/"
	order := plugin.DecodeOrder(req.Order)
	switch order.Type {
	case "alipay":
		return map[string]any{
			"type": "qrcode",
			"page": "alipay_qrcode",
			"url":  codeURL,
		}, nil
	case "wxpay":
		page := "wxpay_qrcode"
		if isMobile(req.Request.UA) {
			page = "wxpay_wap"
		}
		return map[string]any{
			"type": "qrcode",
			"page": page,
			"url":  codeURL,
		}, nil
	default:
		if order.Type == "" {
			return map[string]any{"type": "error", "msg": "缺少支付方式类型"}, nil
		}
		return map[string]any{"type": "error", "msg": fmt.Sprintf("不支持的支付方式: %s", order.Type)}, nil
	}
}

func query(ctx context.Context, req *plugin.CallRequest) (map[string]any, error) {
	order := plugin.DecodeOrder(req.Order)
	out := map[string]any{
		"trade_no":     order.TradeNo,
		"out_trade_no": order.OutTradeNo,
		"amount":       order.Money,
		"subject":      order.Name,
		"state":        0,
	}
	return map[string]any{
		"type": "json",
		"data": out,
	}, nil
}

func refund(ctx context.Context, req *plugin.CallRequest) (map[string]any, error) {
	refund := plugin.DecodeRefund(req.Refund)
	refundNo := refund.OutRefundNo
	if refundNo == "" {
		refundNo = refund.RefundNo
	}
	if refundNo == "" {
		refundNo = fmt.Sprintf("R%s", time.Now().Format("20060102150405"))
	}
	return map[string]any{
		"state":         1,
		"api_refund_no": refundNo,
	}, nil
}

func transfer(ctx context.Context, req *plugin.CallRequest) (map[string]any, error) {
	transfer := plugin.DecodeTransfer(req.Transfer)
	tradeNo := ""
	if transfer != nil {
		tradeNo = transfer.TradeNo
	}
	if tradeNo != "" {
		// 演示插件通过 RPC 回调完成转账。
		_ = plugin.CompleteTransferFromCall(ctx, req, plugin.CompleteTransferRequest{
			TradeNo: tradeNo,
			Status:  1,
		})
	}
	return map[string]any{
		"state": 1,
	}, nil
}

func notify(ctx context.Context, req *plugin.CallRequest) (map[string]any, error) {
	order := plugin.DecodeOrder(req.Order)
	if order.TradeNo != "" {
		_ = plugin.CompleteOrderFromCall(ctx, req, plugin.CompleteOrderRequest{
			TradeNo: order.TradeNo,
		})
	}
	return map[string]any{
		"type": "html",
		"data": "success",
	}, nil
}

func ret(ctx context.Context, req *plugin.CallRequest) (map[string]any, error) {
	return map[string]any{
		"type": "jump",
		"url":  "/",
	}, nil
}

func isMobile(ua string) bool {
	ua = strings.ToLower(ua)
	if ua == "" {
		return false
	}
	return strings.Contains(ua, "mobile") ||
		strings.Contains(ua, "android") ||
		strings.Contains(ua, "iphone") ||
		strings.Contains(ua, "ipad")
}
