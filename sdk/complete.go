package sdk

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ppswws/okpay-plugin-sdk/contract"
	"github.com/ppswws/okpay-plugin-sdk/proto"
)

// CompleteOrderInput 表示订单完成回调参数。
type CompleteOrderInput struct {
	TradeNo    string `json:"tradeNo"`
	APITradeNo string `json:"apiTradeNo,omitempty"`
	Buyer      string `json:"buyer,omitempty"`
	TS         int64  `json:"ts,omitempty"`
	Sign       string `json:"sign,omitempty"`
}

// CompleteRefundInput 表示退款完成回调参数。
type CompleteRefundInput struct {
	RefundNo    string `json:"refundNo"`
	Status      int16  `json:"status"`
	APIRefundNo string `json:"apiRefundNo,omitempty"`
	RespBody    string `json:"respBody,omitempty"`
	TS          int64  `json:"ts,omitempty"`
	Sign        string `json:"sign,omitempty"`
}

// CompleteTransferInput 表示代付完成回调参数。
type CompleteTransferInput struct {
	TradeNo    string `json:"tradeNo"`
	Status     int16  `json:"status"`
	APITradeNo string `json:"apiTradeNo,omitempty"`
	Result     string `json:"result,omitempty"`
	TS         int64  `json:"ts,omitempty"`
	Sign       string `json:"sign,omitempty"`
}

// CompleteCNotifyInput 表示渠道回调日志参数。
type CompleteCNotifyInput struct {
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
func CompleteOrder(ctx context.Context, req CompleteOrderInput) error {
	kernel, conn, err := contract.DialKernelServiceFromContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	requestID := callbackRequestID(req.TradeNo)
	ack, err := kernel.CompleteOrder(ctx, &proto.CompleteOrderRequest{
		RequestId:  requestID,
		TradeNo:    req.TradeNo,
		ApiTradeNo: req.APITradeNo,
		Buyer:      req.Buyer,
	})
	if err != nil {
		return err
	}
	if ack == nil || !ack.Accepted {
		return fmt.Errorf("kernel complete order rejected")
	}
	return nil
}

// CompleteRefund 通过反向 RPC 通知核心完成退款。
func CompleteRefund(ctx context.Context, req CompleteRefundInput) error {
	kernel, conn, err := contract.DialKernelServiceFromContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	requestID := callbackRequestID(req.RefundNo)
	ack, err := kernel.CompleteRefund(ctx, &proto.CompleteRefundRequest{
		RequestId:   requestID,
		RefundNo:    req.RefundNo,
		Status:      int32(req.Status),
		ApiRefundNo: req.APIRefundNo,
		RespBody:    req.RespBody,
	})
	if err != nil {
		return err
	}
	if ack == nil || !ack.Accepted {
		return fmt.Errorf("kernel complete refund rejected")
	}
	return nil
}

// CompleteTransfer 通过反向 RPC 通知核心完成代付。
func CompleteTransfer(ctx context.Context, req CompleteTransferInput) error {
	kernel, conn, err := contract.DialKernelServiceFromContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	requestID := callbackRequestID(req.TradeNo)
	ack, err := kernel.CompleteTransfer(ctx, &proto.CompleteTransferRequest{
		RequestId:  requestID,
		TradeNo:    req.TradeNo,
		Status:     int32(req.Status),
		ApiTradeNo: req.APITradeNo,
		Result:     req.Result,
	})
	if err != nil {
		return err
	}
	if ack == nil || !ack.Accepted {
		return fmt.Errorf("kernel complete transfer rejected")
	}
	return nil
}

// CompleteCNotify 通过反向 RPC 通知核心记录渠道回调日志。
func CompleteCNotify(ctx context.Context, req CompleteCNotifyInput) error {
	kernel, conn, err := contract.DialKernelServiceFromContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	requestID := callbackRequestID(req.TradeNo)
	ack, err := kernel.RecordCNotify(ctx, &proto.RecordCNotifyRequest{
		RequestId:       requestID,
		BizType:         req.BizType,
		TradeNo:         req.TradeNo,
		RequestIp:       req.RequestIP,
		RequestUrl:      req.RequestURL,
		RequestBodyRaw:  []byte(req.RequestBody),
		ResponseBodyRaw: []byte(req.ResponseBody),
	})
	if err != nil {
		return err
	}
	if ack == nil || !ack.Accepted {
		return fmt.Errorf("kernel record cnotify rejected")
	}
	return nil
}

func callbackRequestID(bizNo string) string {
	return "cb:" + bizNo + ":" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
