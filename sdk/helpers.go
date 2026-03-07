package sdk

import (
	"context"
	"fmt"
	"strings"

	"okpay/payment/plugin/proto"
)

// CreateHandlerFunc handles one pay type create flow and returns a page payload.
type CreateHandlerFunc func(context.Context, *proto.InvokeContext) (*proto.PageResponse, error)

// CreateWithHandlers handles create by pay type.
func CreateWithHandlers(ctx context.Context, req *proto.InvokeContext, handlers map[string]CreateHandlerFunc) (*proto.PageResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request 为空")
	}
	order := req.GetOrder()
	if order == nil {
		return nil, fmt.Errorf("订单为空")
	}
	payType := order.GetType()
	if payType == "" {
		return nil, fmt.Errorf("支付方式为空")
	}
	handler := handlers[payType]
	if handler == nil {
		return nil, fmt.Errorf("不支持的支付方式")
	}
	siteDomain := req.GetConfig().GetSiteDomain()
	if siteDomain == "" {
		return nil, fmt.Errorf("缺少 site_domain")
	}

	res, err := handler(ctx, req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("插件返回为空")
	}
	resType := strings.ToLower(strings.TrimSpace(res.GetType()))
	switch resType {
	case ResponseTypeJump:
		if strings.TrimSpace(res.GetUrl()) == "" {
			return nil, fmt.Errorf("插件返回缺少 url")
		}
		return res, nil
	case ResponseTypeError:
		msg := strings.TrimSpace(res.GetMsg())
		if msg == "" {
			msg = "渠道返回失败"
		}
		return nil, fmt.Errorf("%s", msg)
	default:
		return RespJump(siteDomain + "/pay/" + payType + "/" + order.GetTradeNo()), nil
	}
}

func IsWeChat(ua string) bool {
	ua = strings.ToLower(ua)
	return strings.Contains(ua, "micromessenger/") && !strings.Contains(ua, "windowswechat")
}

func IsAlipay(ua string) bool {
	ua = strings.ToLower(ua)
	return strings.Contains(ua, "alipayclient/")
}

func IsMobileQQ(ua string) bool {
	ua = strings.ToLower(ua)
	return strings.Contains(ua, "qq/")
}

func IsMobile(ua string) bool {
	ua = strings.ToLower(ua)
	needles := []string{
		"android", "midp", "nokia", "mobile", "iphone", "ipod",
		"blackberry", "windows phone", "tablet", "ipad",
		"xiaomi", "huawei", "honor", "oppo", "vivo",
		"meizu", "realme", "oneplus", "iqoo",
	}
	for _, n := range needles {
		if strings.Contains(ua, n) {
			return true
		}
	}
	return false
}
