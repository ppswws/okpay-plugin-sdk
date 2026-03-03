package sdk

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"okpay/payment/plugin/contract"
)

type NotifyResponse struct {
	BizType string
	Result  any
}

// RespNotify records cnotify and returns the response result.
func RespNotify(ctx context.Context, call *contract.CallRequest, data NotifyResponse) (map[string]any, error) {
	tradeNo := inferNotifyTradeNo(call, data.BizType)
	req := CompleteCNotifyRequest{
		BizType:      String(data.BizType),
		TradeNo:      tradeNo,
		RequestIP:    "",
		RequestURL:   "",
		RequestBody:  "",
		ResponseBody: "",
	}
	resultMap := toResponseMap(data.Result)
	req.ResponseBody = encodeResponseBody(resultMap)
	if call != nil {
		req.RequestIP = String(call.Request.IP)
		req.RequestURL = String(call.Request.URL)
		req.RequestBody = encodeRequestBody(call)
	}
	if err := CompleteCNotify(ctx, call, req); err != nil {
		log.Printf("[plugin-sdk] complete cnotify failed bizType=%s tradeNo=%s err=%v", req.BizType, req.TradeNo, err)
	}
	return resultMap, nil
}

func inferNotifyTradeNo(call *contract.CallRequest, bizType string) string {
	if call == nil {
		return ""
	}
	switch strings.ToLower(String(bizType)) {
	case contract.BizTypeOrder:
		if order := DecodeOrder(call.Order); order != nil {
			return String(order.TradeNo)
		}
	case contract.BizTypeRefund:
		if refund := DecodeRefund(call.Refund); refund != nil {
			return String(refund.RefundNo)
		}
	case contract.BizTypeTransfer:
		if transfer := DecodeTransfer(call.Transfer); transfer != nil {
			return String(transfer.TradeNo)
		}
	}
	// Fallback: try all known payloads when bizType is empty/unknown.
	if order := DecodeOrder(call.Order); order != nil && String(order.TradeNo) != "" {
		return String(order.TradeNo)
	}
	if refund := DecodeRefund(call.Refund); refund != nil && String(refund.RefundNo) != "" {
		return String(refund.RefundNo)
	}
	if transfer := DecodeTransfer(call.Transfer); transfer != nil && String(transfer.TradeNo) != "" {
		return String(transfer.TradeNo)
	}
	return ""
}

func encodeRequestBody(call *contract.CallRequest) string {
	if call == nil {
		return ""
	}
	if raw := String(call.Request.Body); raw != "" {
		return raw
	}
	return ""
}

func encodeResponseBody(result map[string]any) string {
	if result == nil {
		return ""
	}
	typ := strings.ToLower(String(result["type"]))
	switch typ {
	case "html":
		if v := String(result["data"]); v != "" {
			return v
		}
	case "json":
		if data, err := json.Marshal(result["data"]); err == nil {
			return string(data)
		}
	case "error":
		if v := String(result["msg"]); v != "" {
			return v
		}
	}
	if data, err := json.Marshal(result); err == nil {
		return string(data)
	}
	return ""
}

func toResponseMap(result any) map[string]any {
	switch v := result.(type) {
	case map[string]any:
		if v == nil {
			return RespError("notify response is nil")
		}
		return v
	default:
		log.Printf("[plugin-sdk] invalid notify response payload type=%T", result)
		return RespError("invalid notify response payload")
	}
}
