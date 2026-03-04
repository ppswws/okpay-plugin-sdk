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
func RespNotify(ctx context.Context, call *contract.InvokeRequestV2, data NotifyResponse) (map[string]any, error) {
	tradeNo := inferNotifyTradeNo(call, data.BizType)
	req := CompleteCNotifyRequest{
		BizType:      stringValue(data.BizType),
		TradeNo:      tradeNo,
		RequestIP:    "",
		RequestURL:   "",
		RequestBody:  "",
		ResponseBody: "",
	}
	resultMap := toResponseMap(data.Result)
	req.ResponseBody = encodeResponseBody(resultMap)
	if call != nil {
		req.RequestIP = stringValue(call.Raw.RequestIP)
		req.RequestURL = stringValue(call.Raw.HTTPURL)
		req.RequestBody = encodeRequestBody(call)
	}
	if err := CompleteCNotify(ctx, call, req); err != nil {
		log.Printf("[plugin-sdk] complete cnotify failed bizType=%s tradeNo=%s err=%v", req.BizType, req.TradeNo, err)
	}
	return resultMap, nil
}

func inferNotifyTradeNo(call *contract.InvokeRequestV2, bizType string) string {
	if call == nil {
		return ""
	}
	switch strings.ToLower(stringValue(bizType)) {
	case contract.BizTypeOrder:
		if order := Order(call); order != nil {
			return stringValue(order.TradeNo)
		}
	case contract.BizTypeRefund:
		if refund := Refund(call); refund != nil {
			return stringValue(refund.RefundNo)
		}
	case contract.BizTypeTransfer:
		if transfer := Transfer(call); transfer != nil {
			return stringValue(transfer.TradeNo)
		}
	}
	// Fallback: try all known payloads when bizType is empty/unknown.
	if order := Order(call); order != nil && stringValue(order.TradeNo) != "" {
		return stringValue(order.TradeNo)
	}
	if refund := Refund(call); refund != nil && stringValue(refund.RefundNo) != "" {
		return stringValue(refund.RefundNo)
	}
	if transfer := Transfer(call); transfer != nil && stringValue(transfer.TradeNo) != "" {
		return stringValue(transfer.TradeNo)
	}
	return ""
}

func encodeRequestBody(call *contract.InvokeRequestV2) string {
	if call == nil {
		return ""
	}
	if raw := stringValue(string(call.Raw.HTTPBodyRaw)); raw != "" {
		return raw
	}
	return ""
}

func encodeResponseBody(result map[string]any) string {
	if result == nil {
		return ""
	}
	typ := strings.ToLower(stringValue(result["type"]))
	switch typ {
	case "html":
		if v := stringValue(result["data"]); v != "" {
			return v
		}
	case "json":
		if data, err := json.Marshal(result["data"]); err == nil {
			return string(data)
		}
	case "error":
		if v := stringValue(result["msg"]); v != "" {
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
