package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type httpEnvelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type httpEnvelopeWithData struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

const paymentConfigKey = "local_host"
const callbackSecretKey = "complete_secret"

var callbackHTTPClient = NewHTTPClient(HTTPClientConfig{
	Timeout: 5 * time.Second,
	Retry:   1,
})

func getConfigString(conf map[string]any, key string) string {
	if conf == nil {
		return ""
	}
	if val, ok := conf[key]; ok {
		if s, ok := val.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func completeViaHTTP(ctx context.Context, call *CallRequest, path string, payload any) error {
	if call == nil {
		return fmt.Errorf("call 为空")
	}
	base := getConfigString(call.Config, paymentConfigKey)
	if base == "" {
		return fmt.Errorf("payment 地址未配置")
	}
	base = strings.TrimRight(base, "/")
	url := base + path
	data, err := encodeCallbackPayload(call, payload)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	respBody, _, _, err := callbackHTTPClient.Do(ctx, http.MethodPost, url, string(data), "application/json")
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	var out httpEnvelope
	if err := json.Unmarshal([]byte(respBody), &out); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	if out.Code != 0 {
		msg := strings.TrimSpace(out.Message)
		if msg == "" {
			msg = "payment 返回错误"
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

func completeViaHTTPWithData(ctx context.Context, call *CallRequest, path string, payload any, out any) error {
	if call == nil {
		return fmt.Errorf("call 为空")
	}
	base := getConfigString(call.Config, paymentConfigKey)
	if base == "" {
		return fmt.Errorf("payment 地址未配置")
	}
	base = strings.TrimRight(base, "/")
	url := base + path
	data, err := encodeCallbackPayload(call, payload)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	respBody, _, _, err := callbackHTTPClient.Do(ctx, http.MethodPost, url, string(data), "application/json")
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	var envelope httpEnvelopeWithData
	if err := json.Unmarshal([]byte(respBody), &envelope); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	if envelope.Code != 0 {
		msg := strings.TrimSpace(envelope.Message)
		if msg == "" {
			msg = "payment 返回错误"
		}
		return fmt.Errorf("%s", msg)
	}
	if out != nil && len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("解析响应数据失败: %w", err)
		}
	}
	return nil
}

func encodeCallbackPayload(call *CallRequest, payload any) ([]byte, error) {
	if call == nil {
		return json.Marshal(payload)
	}
	secret := getConfigString(call.Config, callbackSecretKey)
	if secret == "" {
		return nil, fmt.Errorf("complete secret 未配置")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	if data == nil {
		data = map[string]any{}
	}
	data["ts"] = time.Now().UnixMilli()
	data["sign"] = SignPayload(data, secret)
	return json.Marshal(data)
}
