package plugin

import "context"

// CompleteOrderRequest 表示订单完成回调参数。
type CompleteOrderRequest struct {
	TradeNo    string `json:"tradeNo"`
	APITradeNo string `json:"apiTradeNo,omitempty"`
	Buyer      string `json:"buyer,omitempty"`
}

// CompleteRefundRequest 表示退款完成回调参数。
type CompleteRefundRequest struct {
	RefundNo    string `json:"refundNo"`
	Status      int16  `json:"status"`
	APIRefundNo string `json:"apiRefundNo,omitempty"`
	RespBody    string `json:"respBody,omitempty"`
}

// CompleteTransferRequest 表示转账完成回调参数。
type CompleteTransferRequest struct {
	TradeNo    string `json:"tradeNo"`
	Status     int16  `json:"status"`
	APITradeNo string `json:"apiTradeNo,omitempty"`
	Result     string `json:"result,omitempty"`
}

// CompleteOrder 通过反向 RPC 通知核心完成订单。
func CompleteOrder(ctx context.Context, call *CallRequest, req CompleteOrderRequest) error {
	return completeViaHTTP(ctx, call, "/api/complete/order", &req)
}

// CompleteRefund 通过反向 RPC 通知核心完成退款。
func CompleteRefund(ctx context.Context, call *CallRequest, req CompleteRefundRequest) error {
	return completeViaHTTP(ctx, call, "/api/complete/refund", &req)
}

// CompleteTransfer 通过反向 RPC 通知核心完成转账。
func CompleteTransfer(ctx context.Context, call *CallRequest, req CompleteTransferRequest) error {
	return completeViaHTTP(ctx, call, "/api/complete/transfer", &req)
}
