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
func CreateWithHandlers(ctx context.Context, req *contract.CallRequest, handlers map[string]HandlerFunc) (map[string]any, error) {
	if req == nil {
		return nil, fmt.Errorf("request 为空")
	}
	order := DecodeOrder(req.Order)
	if order == nil {
		return nil, fmt.Errorf("订单为空")
	}
	payType := String(order.Type)
	if payType == "" {
		return nil, fmt.Errorf("支付方式为空")
	}
	handler := handlers[payType]
	if handler == nil {
		return nil, fmt.Errorf("不支持的支付方式")
	}
	siteDomain := strings.TrimRight(String(req.Config["sitedomain"]), "/")
	if siteDomain == "" {
		return nil, fmt.Errorf("缺少 sitedomain")
	}

	// 1) 先尝试让子函数直接给出跳转链接（直达渠道）
	res, err := handler(ctx, req)
	if err != nil {
		return nil, err
	}
	resType := String(res["type"])
	if strings.EqualFold(resType, "jump") {
		if url := String(res["url"]); url != "" {
			return RespJump(url), nil
		}
		return nil, fmt.Errorf("插件返回缺少 url")
	}
	if strings.EqualFold(resType, "error") {
		if msg := String(res["msg"]); msg != "" {
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
			val := String(item)
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
func String(value any) string {
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

// DecodeJSONMap decodes JSON object and keeps number lexemes as json.Number.
// This preserves numeric text style (e.g. 1.00 vs 1) for signature scenarios.
func DecodeJSONMap(raw string) (map[string]any, error) {
	out := map[string]any{}
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
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
