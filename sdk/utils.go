package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"okpay/payment/plugin/contract"
)

// CreateWithHandlers 通用 create：优先直跳渠道，否则回退本地中转。
func CreateWithHandlers(ctx context.Context, req *contract.InvokeRequestV2, handlers map[string]HandlerFunc) (map[string]any, error) {
	if req == nil {
		return nil, fmt.Errorf("request 为空")
	}
	order := Order(req)
	if order == nil {
		return nil, fmt.Errorf("订单为空")
	}
	payType := stringValue(order.Type)
	if payType == "" {
		return nil, fmt.Errorf("支付方式为空")
	}
	handler := handlers[payType]
	if handler == nil {
		return nil, fmt.Errorf("不支持的支付方式")
	}
	globalCfg := GlobalConfig(req)
	siteDomain := MapString(globalCfg, "sitedomain")
	siteDomain = strings.TrimRight(siteDomain, "/")
	if siteDomain == "" {
		return nil, fmt.Errorf("缺少 sitedomain")
	}

	// 1) 先尝试让子函数直接给出跳转链接（直达渠道）
	res, err := handler(ctx, req)
	if err != nil {
		return nil, err
	}
	resType := stringValue(res["type"])
	if strings.EqualFold(resType, "jump") {
		if url := stringValue(res["url"]); url != "" {
			return RespJump(url), nil
		}
		return nil, fmt.Errorf("插件返回缺少 url")
	}
	if strings.EqualFold(resType, "error") {
		if msg := stringValue(res["msg"]); msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("渠道返回失败")
	}

	// 2) 子函数需要前端承接时，回退到本地支付入口
	return RespJump(siteDomain + "/pay/" + payType + "/" + order.TradeNo), nil
}

// IsWeChat 判断微信环境。
func IsWeChat(ua string) bool {
	ua = strings.ToLower(ua)
	return strings.Contains(ua, "micromessenger/") && !strings.Contains(ua, "windowswechat")
}

// IsAlipay 判断支付宝环境。
func IsAlipay(ua string) bool {
	ua = strings.ToLower(ua)
	return strings.Contains(ua, "alipayclient/")
}

// IsMobileQQ 判断 QQ 环境。
func IsMobileQQ(ua string) bool {
	ua = strings.ToLower(ua)
	return strings.Contains(ua, "qq/")
}

// IsMobile 判断移动端环境。
func IsMobile(ua string) bool {
	ua = strings.ToLower(ua)
	needles := []string{
		"android", "midp", "nokia", "mobile", "iphone", "ipod",
		"blackberry", "windows phone", "tablet", "ipad",
		"xiaomi", "huawei", "honor", "oppo", "vivo",
		"meizu", "realme", "oneplus", "iqoo",
	}
	for _, n := range needles {
		if n != "" && strings.Contains(ua, n) {
			return true
		}
	}
	return false
}

// ParseRequestParams reads parsed.request.params as a unified string map.
func ParseRequestParams(req *contract.InvokeRequestV2) map[string]string {
	return parsedStringMap(req, "params")
}

// QueryParam returns parsed.request.query[key].
func QueryParam(req *contract.InvokeRequestV2, key string) string {
	key = strings.TrimSpace(key)
	if req == nil || key == "" {
		return ""
	}
	return parsedStringMap(req, "query")[key]
}

// MapString reads a string-like field from map[string]any in a single place.
// It avoids plugin-local duplicates like docStringField/mapDoc.
func MapString(m map[string]any, key string) string {
	if len(m) == 0 || strings.TrimSpace(key) == "" {
		return ""
	}
	val, ok := m[key]
	if !ok {
		return ""
	}
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v)
	case json.Number:
		return strings.TrimSpace(v.String())
	default:
		return stringValue(v)
	}
}

func parsedStringMap(req *contract.InvokeRequestV2, field string) map[string]string {
	out := map[string]string{}
	if req == nil || strings.TrimSpace(field) == "" {
		return out
	}
	root, ok := req.Parsed.Data.Fields["request"]
	if !ok || root.Kind != contract.ValueKindObject || root.Object == nil {
		return out
	}
	node, ok := root.Object.Fields[field]
	if !ok || node.Kind != contract.ValueKindObject || node.Object == nil {
		return out
	}
	for key, val := range node.Object.Fields {
		switch val.Kind {
		case contract.ValueKindString:
			out[key] = val.String
		case contract.ValueKindInt64:
			out[key] = strconv.FormatInt(val.Int64, 10)
		case contract.ValueKindUInt64:
			out[key] = strconv.FormatUint(val.UInt64, 10)
		case contract.ValueKindDecimal:
			out[key] = val.Decimal
		case contract.ValueKindBool:
			out[key] = strconv.FormatBool(val.Bool)
		}
	}
	return out
}

// ReadStringSlice normalizes config value into a string slice.
func ReadStringSlice(value any) []string {
	out := []string{}
	switch v := value.(type) {
	case []string:
		for _, item := range v {
			if item != "" {
				out = append(out, item)
			}
		}
	case []any:
		for _, item := range v {
			val := stringValue(item)
			if val != "" {
				out = append(out, val)
			}
		}
	case string:
		for _, item := range strings.Split(v, ",") {
			if item != "" {
				out = append(out, item)
			}
		}
	case []byte:
		for _, item := range strings.Split(string(v), ",") {
			if item != "" {
				out = append(out, item)
			}
		}
	}
	return out
}

// String converts value to string and returns empty when value is nil/typed-nil.
// It also filters the literal "<nil>" to avoid leaking typed-nil via fmt.Sprint.
func stringValue(value any) string {
	if value == nil {
		return ""
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		if rv.IsNil() {
			return ""
		}
	}
	switch v := value.(type) {
	case string:
		out := strings.TrimSpace(v)
		if out == "<nil>" {
			return ""
		}
		return out
	case []byte:
		out := strings.TrimSpace(string(v))
		if out == "<nil>" {
			return ""
		}
		return out
	case fmt.Stringer:
		out := strings.TrimSpace(v.String())
		if out == "<nil>" {
			return ""
		}
		return out
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		out := strings.TrimSpace(fmt.Sprint(v))
		if out == "<nil>" {
			return ""
		}
		return out
	}
}

// ModeSet converts a string slice into a lookup set.
func ModeSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, v := range values {
		if v != "" {
			out[v] = true
		}
	}
	return out
}

// AllowMode returns true when modes is empty or code is enabled.
func AllowMode(modes map[string]bool, code string) bool {
	if len(modes) == 0 {
		return true
	}
	return modes[code]
}

// Read returns value at dot path from req.Parsed (single read entry).
func Read(req *contract.InvokeRequestV2, path string) (contract.Value, bool) {
	if req == nil {
		return contract.Value{}, false
	}
	parts := splitPath(path)
	if len(parts) == 0 {
		return contract.Value{}, false
	}
	cur := contract.Value{Kind: contract.ValueKindObject, Object: &req.Parsed.Data}
	for _, key := range parts {
		if cur.Kind != contract.ValueKindObject || cur.Object == nil || cur.Object.Fields == nil {
			return contract.Value{}, false
		}
		next, ok := cur.Object.Fields[key]
		if !ok {
			return contract.Value{}, false
		}
		cur = next
	}
	return cur, true
}

func splitPath(path string) []string {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	parts := strings.Split(path, ".")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
