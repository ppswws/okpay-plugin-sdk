package sdk

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/ppswws/okpay-plugin-sdk/contract"
	"github.com/ppswws/okpay-plugin-sdk/proto"
)

// RecordNotify records cnotify and returns the original payload.
func RecordNotify(ctx context.Context, req *proto.InvokeContext, bizType string, result *proto.PageResponse) *proto.PageResponse {
	if result == nil {
		result = RespError("invalid notify response payload")
	}
	notifyReq := CompleteCNotifyInput{
		BizType:      normalizeBizType(bizType),
		TradeNo:      inferNotifyTradeNo(req, bizType),
		RequestIP:    req.GetRequest().GetIp(),
		RequestURL:   req.GetRequest().GetUrl(),
		RequestBody:  string(req.GetRequest().GetBody()),
		ResponseBody: encodeResponseBody(result),
	}
	if err := CompleteCNotify(ctx, notifyReq); err != nil {
		log.Printf("[plugin-sdk] complete cnotify failed bizType=%s tradeNo=%s err=%v", notifyReq.BizType, notifyReq.TradeNo, err)
	}
	return result
}

func normalizeBizType(in string) string {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case contract.BizTypeOrder:
		return contract.BizTypeOrder
	case contract.BizTypeRefund:
		return contract.BizTypeRefund
	case contract.BizTypeTransfer:
		return contract.BizTypeTransfer
	default:
		return ""
	}
}

func inferNotifyTradeNo(req *proto.InvokeContext, bizType string) string {
	if req == nil {
		return ""
	}
	switch bizType {
	case contract.BizTypeOrder:
		return req.GetOrder().GetTradeNo()
	case contract.BizTypeRefund:
		return req.GetRefund().GetRefundNo()
	case contract.BizTypeTransfer:
		return req.GetTransfer().GetTradeNo()
	}
	if tradeNo := strings.TrimSpace(req.GetTradeNo()); tradeNo != "" {
		return tradeNo
	}
	if tradeNo := strings.TrimSpace(req.GetOrder().GetTradeNo()); tradeNo != "" {
		return tradeNo
	}
	if refundNo := strings.TrimSpace(req.GetRefund().GetRefundNo()); refundNo != "" {
		return refundNo
	}
	if tradeNo := strings.TrimSpace(req.GetTransfer().GetTradeNo()); tradeNo != "" {
		return tradeNo
	}
	return ""
}

func encodeResponseBody(result *proto.PageResponse) string {
	if result == nil {
		return ""
	}
	typ := strings.ToLower(strings.TrimSpace(result.GetType()))
	switch typ {
	case ResponseTypeHTML:
		if v := strings.TrimSpace(result.GetDataText()); v != "" {
			return v
		}
	case ResponseTypeJSON:
		if len(result.GetDataJsonRaw()) > 0 {
			return string(result.GetDataJsonRaw())
		}
	case ResponseTypeError:
		if v := strings.TrimSpace(result.GetMsg()); v != "" {
			return v
		}
	}
	if data, err := json.Marshal(result); err == nil {
		return string(data)
	}
	return ""
}
