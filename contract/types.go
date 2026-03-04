package contract

import (
	"context"
	"encoding/json"
)

// PaymentChannel 定义支付插件标准接口，主程序通过 go-plugin 调用。
type PaymentChannel interface {
	// InvokeV2 为唯一插件调用入口，保留 action 的自由命名能力。
	InvokeV2(ctx context.Context, req *InvokeRequestV2) (*InvokeResponseV2, error)
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
	Raw        map[string]any        `json:"-"`
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

// ToMap 转为 map（用于存储 info 字段）。
func (p *PluginInfo) ToMap() map[string]any {
	if p == nil {
		return map[string]any{}
	}
	if len(p.Raw) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(p.Raw))
	for key, val := range p.Raw {
		out[key] = val
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
