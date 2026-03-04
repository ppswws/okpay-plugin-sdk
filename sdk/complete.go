package sdk

import (
	"context"

	"okpay/payment/plugin/contract"
)

// CompleteOrderRequest 表示订单完成回调参数。
type CompleteOrderRequest struct {
	TradeNo    string `json:"tradeNo"`
	APITradeNo string `json:"apiTradeNo,omitempty"`
	Buyer      string `json:"buyer,omitempty"`
	TS         int64  `json:"ts,omitempty"`
	Sign       string `json:"sign,omitempty"`
}

// CompleteRefundRequest 表示退款完成回调参数。
type CompleteRefundRequest struct {
	RefundNo    string `json:"refundNo"`
	Status      int16  `json:"status"`
	APIRefundNo string `json:"apiRefundNo,omitempty"`
	RespBody    string `json:"respBody,omitempty"`
	TS          int64  `json:"ts,omitempty"`
	Sign        string `json:"sign,omitempty"`
}

// CompleteTransferRequest 表示代付完成回调参数。
type CompleteTransferRequest struct {
	TradeNo    string `json:"tradeNo"`
	Status     int16  `json:"status"`
	APITradeNo string `json:"apiTradeNo,omitempty"`
	Result     string `json:"result,omitempty"`
	TS         int64  `json:"ts,omitempty"`
	Sign       string `json:"sign,omitempty"`
}

// CompleteCNotifyRequest 表示渠道回调日志参数。
type CompleteCNotifyRequest struct {
	BizType      string `json:"bizType"`
	TradeNo      string `json:"tradeNo"`
	RequestIP    string `json:"requestIp,omitempty"`
	RequestURL   string `json:"requestUrl,omitempty"`
	RequestBody  string `json:"requestBody,omitempty"`
	ResponseBody string `json:"responseBody,omitempty"`
	TS           int64  `json:"ts,omitempty"`
	Sign         string `json:"sign,omitempty"`
}

// CompleteOrder 通过反向 RPC 通知核心完成订单。
func CompleteOrder(ctx context.Context, call *contract.InvokeRequestV2, req CompleteOrderRequest) error {
	return completeViaHTTP(ctx, call, "/api/complete/order", map[string]any{
		"tradeNo":    req.TradeNo,
		"apiTradeNo": req.APITradeNo,
		"buyer":      req.Buyer,
	})
}

// CompleteRefund 通过反向 RPC 通知核心完成退款。
func CompleteRefund(ctx context.Context, call *contract.InvokeRequestV2, req CompleteRefundRequest) error {
	return completeViaHTTP(ctx, call, "/api/complete/refund", map[string]any{
		"refundNo":    req.RefundNo,
		"status":      req.Status,
		"apiRefundNo": req.APIRefundNo,
		"respBody":    req.RespBody,
	})
}

// CompleteTransfer 通过反向 RPC 通知核心完成代付。
func CompleteTransfer(ctx context.Context, call *contract.InvokeRequestV2, req CompleteTransferRequest) error {
	return completeViaHTTP(ctx, call, "/api/complete/transfer", map[string]any{
		"tradeNo":    req.TradeNo,
		"status":     req.Status,
		"apiTradeNo": req.APITradeNo,
		"result":     req.Result,
	})
}

// CompleteCNotify 通过反向 RPC 通知核心记录渠道回调日志。
func CompleteCNotify(ctx context.Context, call *contract.InvokeRequestV2, req CompleteCNotifyRequest) error {
	return completeViaHTTP(ctx, call, "/api/complete/cnotify", map[string]any{
		"bizType":      req.BizType,
		"tradeNo":      req.TradeNo,
		"requestIp":    req.RequestIP,
		"requestUrl":   req.RequestURL,
		"requestBody":  req.RequestBody,
		"responseBody": req.ResponseBody,
	})
}
