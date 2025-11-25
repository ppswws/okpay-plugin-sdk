package plugin

import (
	"context"
	"encoding/json"
)

// PaymentChannel 定义支付插件标准接口，主程序通过 go-plugin 调用。
type PaymentChannel interface {
	// Call 通用扩展调用（函数名由插件自定义，如 info/create/query/refund/notify/return）。
	Call(ctx context.Context, funcName string, req *CallRequest) (map[string]any, error)
}

// PluginInfo 插件自描述元信息（与插件表字段一致）。
type PluginInfo struct {
	ID         string                `json:"id"`
	Name       string                `json:"name"`
	Link       string                `json:"link,omitempty"`
	Paytypes   []string              `json:"paytypes,omitempty"`
	Transtypes []string              `json:"transtypes,omitempty"`
	Inputs     map[string]InputField `json:"inputs,omitempty"`
	Note       string                `json:"note,omitempty"`
}

// InputField 定义插件动态表单的输入项。
type InputField struct {
	Name     string            `json:"name,omitempty"`
	Type     string            `json:"type,omitempty"`
	Note     string            `json:"note,omitempty"`
	Required bool              `json:"required,omitempty"`
	Default  any               `json:"default,omitempty"`
	Options  map[string]string `json:"options,omitempty"`
}

// CallRequest 通用调用上下文。
type CallRequest struct {
	Channel map[string]any `json:"channel"` // 通道配置
	Order   map[string]any `json:"order"`   // 订单信息
	Config  map[string]any `json:"conf"`    // 全局配置（如 notify/return 前缀）
	Request HTTPRequest    `json:"req"`     // HTTP 请求上下文
}

// HTTPRequest 对外暴露的请求上下文。
type HTTPRequest struct {
	Method  string            `json:"method,omitempty"`
	Query   map[string]any    `json:"query,omitempty"`
	Body    map[string]any    `json:"body,omitempty"`
	IP      string            `json:"ip,omitempty"`
	UA      string            `json:"ua,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// ToMap 转为 map（用于存储 info 字段）。
func (p *PluginInfo) ToMap() map[string]any {
	if p == nil {
		return map[string]any{}
	}
	out := map[string]any{
		"id":   p.ID,
		"name": p.Name,
	}
	if p.Link != "" {
		out["link"] = p.Link
	}
	if len(p.Paytypes) > 0 {
		out["paytypes"] = p.Paytypes
	}
	if len(p.Transtypes) > 0 {
		out["transtypes"] = p.Transtypes
	}
	if len(p.Inputs) > 0 {
		out["inputs"] = p.Inputs
	}
	if p.Note != "" {
		out["note"] = p.Note
	}
	return out
}

// AsJSON 返回 PluginInfo 的 JSON 字符串（用于存储 info 字段）。
func (p *PluginInfo) AsJSON() string {
	data, err := json.Marshal(p.ToMap())
	if err != nil {
		return ""
	}
	return string(data)
}
