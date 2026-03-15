package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ppswws/okpay-plugin-sdk/contract"
	"github.com/ppswws/okpay-plugin-sdk/proto"
)

type RequestStats struct {
	ReqBody  string
	RespBody string
	ReqCount int16
	ReqMs    int32
}

// LockOrderExt 仅在 Ext 为空时执行 fetch 并写入 Ext；存在则直接复用 Ext。
// 插件自行决定何时缓存（建议在真正请求渠道后）。
func LockOrderExt(ctx context.Context, tradeNo string, fetch func() (any, RequestStats, error)) (map[string]any, error) {
	if tradeNo == "" {
		return nil, fmt.Errorf("tradeNo 不能为空")
	}
	if fetch == nil {
		return nil, fmt.Errorf("fetch 不能为空")
	}
	kernel, conn, err := contract.DialKernelServiceFromContext(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	cachedExt, err := lockOrderData(ctx, kernel, tradeNo, RequestStats{}, nil)
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
		return nil, err
	}
	if msg, ok := errorPayloadMsg(result); ok {
		return nil, errors.New(msg)
	}
	lockedExt, err := lockOrderData(ctx, kernel, tradeNo, stats, result)
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

func lockOrderData(ctx context.Context, kernel contract.KernelService, tradeNo string, stats RequestStats, ext any) (string, error) {
	extRaw := []byte(nil)
	if ext != nil {
		b, err := marshalJSONNoEscape(ext)
		if err != nil {
			return "", err
		}
		extRaw = b
	}
	resp, err := kernel.LockOrderExt(ctx, &proto.LockExtReq{
		RequestId: cbReqID(tradeNo),
		TradeNo:   tradeNo,
		ReqBody:   stats.ReqBody,
		RespBody:  stats.RespBody,
		ReqCount:  int32(stats.ReqCount),
		ReqMs:     stats.ReqMs,
		ExtRaw:    extRaw,
	})
	if err != nil {
		return "", err
	}
	return string(resp.GetExtRaw()), nil
}

func marshalJSONNoEscape(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSpace(buf.Bytes()), nil
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
		if v == "" {
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

func stringValue(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case json.RawMessage:
		return string(v)
	default:
		return ""
	}
}
