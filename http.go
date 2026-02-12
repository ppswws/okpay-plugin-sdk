package plugin

import (
	"bytes"
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
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	var out httpEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
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
