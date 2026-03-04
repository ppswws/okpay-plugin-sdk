package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"okpay/payment/plugin/contract"
)

type RequestStats struct {
	ReqBody  string
	RespBody string
	ReqCount int16
	ReqMs    int32
}

// LockOrderExt 仅在 Ext 为空时执行 fetch 并写入 Ext；存在则直接复用 Ext。
// 插件自行决定何时缓存（建议在真正请求渠道后）。
func LockOrderExt(ctx context.Context, call *contract.InvokeRequestV2, tradeNo string, fetch func() (any, RequestStats, error)) (map[string]any, error) {
	if call == nil {
		return nil, fmt.Errorf("call 不能为空")
	}
	if strings.TrimSpace(tradeNo) == "" {
		return nil, fmt.Errorf("tradeNo 不能为空")
	}
	if fetch == nil {
		return nil, fmt.Errorf("fetch 不能为空")
	}
	cachedExt, err := lockOrderData(ctx, call, tradeNo, RequestStats{}, nil)
	if err != nil {
		return nil, err
	}
	if cachedExt != "" {
		if cached, ok := extractPayloadFromAny(cachedExt); ok && cached != nil {
			return cached, nil
		}
	}
	result, stats, err := fetch()
	if err != nil {
		_, _ = lockOrderData(ctx, call, tradeNo, stats, nil)
		return nil, err
	}
	if msg, ok := errorPayloadMsg(result); ok {
		_, _ = lockOrderData(ctx, call, tradeNo, stats, nil)
		return nil, errors.New(msg)
	}
	lockedExt, err := lockOrderData(ctx, call, tradeNo, stats, result)
	if err != nil {
		return nil, err
	}
	if lockedExt != "" {
		if cached, ok := extractPayloadFromAny(lockedExt); ok && cached != nil {
			return cached, nil
		}
	}
	if out, ok := extractPayloadFromAny(result); ok {
		return out, nil
	}
	return nil, nil
}

type lockOrderDataResponse struct {
	Ext string `json:"ext,omitempty"`
}

func lockOrderData(ctx context.Context, call *contract.InvokeRequestV2, tradeNo string, stats RequestStats, ext any) (string, error) {
	extStr := ""
	if ext != nil {
		b, err := json.Marshal(ext)
		if err != nil {
			return "", err
		}
		extStr = string(b)
	}
	var resp lockOrderDataResponse
	if err := completeViaHTTPWithData(ctx, call, "/api/complete/orderdata/lock", map[string]any{
		"tradeNo":  tradeNo,
		"reqBody":  stats.ReqBody,
		"respBody": stats.RespBody,
		"reqCount": stats.ReqCount,
		"reqMs":    stats.ReqMs,
		"ext":      extStr,
	}, &resp); err != nil {
		return "", err
	}
	return resp.Ext, nil
}

func extractPayloadFromAny(value any) (map[string]any, bool) {
	switch v := value.(type) {
	case map[string]any:
		return v, true
	case json.RawMessage:
		if len(v) == 0 {
			return nil, false
		}
		var out map[string]any
		if err := json.Unmarshal(v, &out); err != nil {
			return nil, false
		}
		return out, true
	case []byte:
		if len(v) == 0 {
			return nil, false
		}
		var out map[string]any
		if err := json.Unmarshal(v, &out); err != nil {
			return nil, false
		}
		return out, true
	case string:
		if strings.TrimSpace(v) == "" {
			return nil, false
		}
		var out map[string]any
		if err := json.Unmarshal([]byte(v), &out); err != nil {
			return nil, false
		}
		return out, true
	default:
		return nil, false
	}
}

func errorPayloadMsg(value any) (string, bool) {
	payload, ok := extractPayloadFromAny(value)
	if !ok || payload == nil {
		return "", false
	}
	t, ok := payload["type"].(string)
	if !ok || t != "error" {
		return "", false
	}
	msg := stringValue(payload["msg"])
	if msg == "" {
		msg = "支付通道返回错误"
	}
	return msg, true
}
