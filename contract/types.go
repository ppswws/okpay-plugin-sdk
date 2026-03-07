package contract

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
