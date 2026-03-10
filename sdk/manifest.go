package sdk

import (
	"fmt"
	"strings"

	"okpay/payment/plugin/proto"
)

// InputSpec defines one plugin config input field.
type InputSpec struct {
	Name         string
	Type         string
	Note         string
	Required     bool
	DefaultValue string
	Options      map[string]string
}

// Manifest defines plugin info in strongly typed form.
type Manifest struct {
	ID         string
	Name       string
	Link       string
	Paytypes   []string
	Transtypes []string
	Inputs     map[string]InputSpec
	Note       string
}

// BuildInfoResponse converts manifest to proto info response with validation.
func BuildInfoResponse(m Manifest) (*proto.PluginInfoResponse, error) {
	id := strings.TrimSpace(m.ID)
	name := strings.TrimSpace(m.Name)
	if id == "" || name == "" {
		return nil, fmt.Errorf("manifest 缺少 id/name")
	}
	out := &proto.PluginInfoResponse{
		Id:         id,
		Name:       name,
		Link:       strings.TrimSpace(m.Link),
		Paytypes:   sanitizeStringSlice(m.Paytypes),
		Transtypes: sanitizeStringSlice(m.Transtypes),
		Inputs:     map[string]*proto.InputField{},
		Note:       strings.TrimSpace(m.Note),
	}
	for key, item := range m.Inputs {
		fieldKey := strings.TrimSpace(key)
		if fieldKey == "" {
			return nil, fmt.Errorf("manifest input key 不能为空")
		}
		if strings.TrimSpace(item.Name) == "" {
			return nil, fmt.Errorf("manifest input[%s] 缺少 name", fieldKey)
		}
		if strings.TrimSpace(item.Type) == "" {
			return nil, fmt.Errorf("manifest input[%s] 缺少 type", fieldKey)
		}
		options := copyStringMap(item.Options)
		typ := strings.TrimSpace(item.Type)
		if (typ == "select" || typ == "checkbox") && len(options) == 0 {
			return nil, fmt.Errorf("manifest input[%s] type=%s 必须提供 options", fieldKey, typ)
		}
		out.Inputs[fieldKey] = &proto.InputField{
			Name:         strings.TrimSpace(item.Name),
			Type:         typ,
			Note:         strings.TrimSpace(item.Note),
			Required:     item.Required,
			DefaultValue: strings.TrimSpace(item.DefaultValue),
			Options:      options,
		}
	}
	return out, nil
}

func sanitizeStringSlice(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(in))
	for _, item := range in {
		val := strings.TrimSpace(item)
		if val != "" {
			out = append(out, val)
		}
	}
	return out
}

func copyStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for key, val := range in {
		out[key] = val
	}
	return out
}
