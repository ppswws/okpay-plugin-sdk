package plugin

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// CreateWithHandlers 通用 create：优先直跳渠道，否则回退本地中转。
func CreateWithHandlers(ctx context.Context, req *CallRequest, handlers map[string]HandlerFunc) (map[string]any, error) {
	if req == nil {
		return nil, fmt.Errorf("request 为空")
	}
	order := DecodeOrder(req.Order)
	if order == nil {
		return nil, fmt.Errorf("订单为空")
	}
	payType := strings.TrimSpace(order.Type)
	if payType == "" {
		return nil, fmt.Errorf("支付方式为空")
	}
	handler := handlers[payType]
	if handler == nil {
		return nil, fmt.Errorf("不支持的支付方式")
	}
	siteDomain := strings.TrimRight(fmt.Sprint(req.Config["sitedomain"]), "/")
	if siteDomain == "" {
		return nil, fmt.Errorf("缺少 sitedomain")
	}

	// 1) 先尝试让子函数直接给出跳转链接（直达渠道）
	res, err := handler(ctx, req)
	if err != nil {
		return nil, err
	}
	resType, _ := res["type"].(string)
	if strings.EqualFold(resType, "jump") {
		if url, _ := res["url"].(string); strings.TrimSpace(url) != "" {
			return map[string]any{"type": "jump", "url": strings.TrimSpace(url)}, nil
		}
		return nil, fmt.Errorf("插件返回缺少 url")
	}
	if strings.EqualFold(resType, "error") {
		if msg, _ := res["msg"].(string); strings.TrimSpace(msg) != "" {
			return nil, fmt.Errorf("%s", strings.TrimSpace(msg))
		}
		return nil, fmt.Errorf("渠道返回失败")
	}

	// 2) 子函数需要前端承接时，回退到本地支付入口
	return map[string]any{"type": "jump", "url": siteDomain + "/pay/" + payType + "/" + order.TradeNo}, nil
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
			val := fmt.Sprint(item)
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

// GetQuery fetches a query param safely from CallRequest.
func GetQuery(req *CallRequest, key string) string {
	if req == nil || req.Request.Query == nil {
		return ""
	}
	value, ok := req.Request.Query[key]
	if !ok || isNilValue(value) {
		return ""
	}
	return fmt.Sprint(value)
}

// isNilValue handles typed-nil values stored in interface{}.
func isNilValue(value any) bool {
	if value == nil {
		return true
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
