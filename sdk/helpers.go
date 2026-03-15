package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-pay/gopay/alipay"
	"github.com/ppswws/okpay-plugin-sdk/proto"
)

// ---- 通用能力 ------------------------------------------------------

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
	case TypeJump:
		if strings.TrimSpace(res.GetUrl()) == "" {
			return nil, fmt.Errorf("插件返回缺少 url")
		}
		return res, nil
	case TypeError:
		msg := strings.TrimSpace(res.GetMsg())
		if msg == "" {
			msg = "渠道返回失败"
		}
		return nil, fmt.Errorf("%s", msg)
	default:
		return RespJump(siteDomain + "/pay/" + payType + "/" + order.GetTradeNo()), nil
	}
}

// ---- 环境判断 ------------------------------------------------------

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

// ---- 微信辅助 ------------------------------------------------------

const wxScope = "snsapi_base"

var httpc = NewHTTPClient(HTTPClientConfig{})

type miniSessResp struct {
	OpenID  string `json:"openid"`
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type miniTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type miniSchemeResp struct {
	OpenLink string `json:"openlink"`
	ErrCode  int    `json:"errcode"`
	ErrMsg   string `json:"errmsg"`
}

// BuildMPOAuthURL returns a WeChat public-account oauth authorize URL.
func BuildMPOAuthURL(appID, redirectURL, state string) string {
	appID = strings.TrimSpace(appID)
	redirectURL = strings.TrimSpace(redirectURL)
	if appID == "" || redirectURL == "" {
		return ""
	}
	base := "https://open.weixin.qq.com/connect/oauth2/authorize"
	return fmt.Sprintf(
		"%s?appid=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s#wechat_redirect",
		base,
		url.QueryEscape(appID),
		url.QueryEscape(redirectURL),
		url.QueryEscape(wxScope),
		url.QueryEscape(strings.TrimSpace(state)),
	)
}

// GetMPOpenid exchanges公众号 oauth code to openid.
func GetMPOpenid(ctx context.Context, appID, appSecret, code string) (string, error) {
	appID = strings.TrimSpace(appID)
	appSecret = strings.TrimSpace(appSecret)
	code = strings.TrimSpace(code)
	if appID == "" || appSecret == "" || code == "" {
		return "", fmt.Errorf("公众号参数缺失")
	}
	endpoint := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		url.QueryEscape(appID),
		url.QueryEscape(appSecret),
		url.QueryEscape(code),
	)
	body, _, _, err := httpc.Do(ctx, "GET", endpoint, "", "")
	if err != nil {
		return "", err
	}
	resp := struct {
		OpenID  string `json:"openid"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}{}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return "", fmt.Errorf("oauth2 响应解析失败: %w", err)
	}
	if resp.ErrCode != 0 {
		return "", fmt.Errorf("获取 openid 失败: %s", resp.ErrMsg)
	}
	openid := strings.TrimSpace(resp.OpenID)
	if openid == "" {
		return "", fmt.Errorf("openid 为空")
	}
	return openid, nil
}

// GetMiniOpenid returns mini-program openid via jscode2session.
func GetMiniOpenid(ctx context.Context, appID, appSecret, code string) (string, error) {
	appID = strings.TrimSpace(appID)
	appSecret = strings.TrimSpace(appSecret)
	code = strings.TrimSpace(code)
	if appID == "" || appSecret == "" || code == "" {
		return "", fmt.Errorf("小程序参数缺失")
	}
	endpoint := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		url.QueryEscape(appID),
		url.QueryEscape(appSecret),
		url.QueryEscape(code),
	)
	body, _, _, err := httpc.Do(ctx, "GET", endpoint, "", "")
	if err != nil {
		return "", err
	}
	resp := miniSessResp{}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return "", fmt.Errorf("小程序 openid 解析失败: %w", err)
	}
	if resp.ErrCode != 0 {
		return "", fmt.Errorf("小程序 openid 获取失败: %s", resp.ErrMsg)
	}
	if strings.TrimSpace(resp.OpenID) == "" {
		return "", fmt.Errorf("小程序 openid 为空")
	}
	return strings.TrimSpace(resp.OpenID), nil
}

// GetMiniScheme creates a mini-program scheme for browser jump.
func GetMiniScheme(ctx context.Context, appID, appSecret, path, query string) (string, error) {
	appID = strings.TrimSpace(appID)
	appSecret = strings.TrimSpace(appSecret)
	if appID == "" || appSecret == "" {
		return "", fmt.Errorf("小程序参数缺失")
	}
	token, err := getMiniToken(ctx, appID, appSecret)
	if err != nil {
		return "", err
	}
	payload := map[string]any{
		"jump_wxa": map[string]any{
			"path":  strings.TrimSpace(path),
			"query": strings.TrimSpace(query),
		},
		"is_expire": false,
	}
	raw, _ := json.Marshal(payload)
	endpoint := fmt.Sprintf("https://api.weixin.qq.com/wxa/generatescheme?access_token=%s", url.QueryEscape(token))
	respBody, _, _, err := httpc.Do(ctx, "POST", endpoint, string(raw), "application/json")
	if err != nil {
		return "", err
	}
	resp := miniSchemeResp{}
	if err := json.Unmarshal([]byte(respBody), &resp); err != nil {
		return "", fmt.Errorf("小程序 scheme 解析失败: %w", err)
	}
	if resp.ErrCode != 0 {
		return "", fmt.Errorf("生成小程序 scheme 失败: %s", resp.ErrMsg)
	}
	if strings.TrimSpace(resp.OpenLink) == "" {
		return "", fmt.Errorf("生成小程序 scheme 失败: openlink 为空")
	}
	return strings.TrimSpace(resp.OpenLink), nil
}

func getMiniToken(ctx context.Context, appID, appSecret string) (string, error) {
	endpoint := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		url.QueryEscape(appID),
		url.QueryEscape(appSecret),
	)
	body, _, _, err := httpc.Do(ctx, "GET", endpoint, "", "")
	if err != nil {
		return "", err
	}
	resp := miniTokenResp{}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return "", fmt.Errorf("获取小程序 access_token 解析失败: %w", err)
	}
	if resp.ErrCode != 0 {
		return "", fmt.Errorf("获取小程序 access_token 失败: %s", resp.ErrMsg)
	}
	if strings.TrimSpace(resp.AccessToken) == "" {
		return "", fmt.Errorf("获取小程序 access_token 失败: 为空")
	}
	return strings.TrimSpace(resp.AccessToken), nil
}

// ---- 支付宝辅助 ----------------------------------------------------

// AliIdentity stores buyer identity resolved from alipay oauth token.
type AliIdentity struct {
	UserID      string
	OpenID      string
	AccessToken string
}

// BuildAliOAuthURL returns Alipay oauth authorize URL by environment.
func BuildAliOAuthURL(appID, redirectURL, state string, isProd bool) string {
	appID = strings.TrimSpace(appID)
	redirectURL = strings.TrimSpace(redirectURL)
	if appID == "" || redirectURL == "" {
		return ""
	}
	base := "https://openauth.alipay.com/oauth2/publicAppAuthorize.htm"
	if !isProd {
		base = "https://openauth-sandbox.dl.alipaydev.com/oauth2/publicAppAuthorize.htm"
	}
	q := url.Values{}
	q.Set("app_id", appID)
	q.Set("scope", "auth_base")
	q.Set("redirect_uri", redirectURL)
	if strings.TrimSpace(state) != "" {
		q.Set("state", strings.TrimSpace(state))
	}
	return base + "?" + q.Encode()
}

// GetAliIdentity exchanges auth_code for buyer identity.
func GetAliIdentity(ctx context.Context, appID, privateKey, authCode string, isProd bool) (AliIdentity, error) {
	appID = strings.TrimSpace(appID)
	privateKey = strings.TrimSpace(privateKey)
	authCode = strings.TrimSpace(authCode)
	if appID == "" || privateKey == "" || authCode == "" {
		return AliIdentity{}, fmt.Errorf("支付宝 oauth 参数缺失")
	}
	client, err := alipay.NewClient(appID, privateKey, isProd)
	if err != nil {
		return AliIdentity{}, fmt.Errorf("初始化支付宝客户端失败: %w", err)
	}
	client.SetCharset(alipay.UTF8).SetSignType(alipay.RSA2)
	resp, err := client.SystemOauthToken(ctx, map[string]any{
		"grant_type": "authorization_code",
		"code":       authCode,
	})
	if err != nil {
		return AliIdentity{}, fmt.Errorf("支付宝 oauth 换取令牌失败: %w", err)
	}
	if resp == nil || resp.Response == nil {
		return AliIdentity{}, fmt.Errorf("支付宝 oauth 返回为空")
	}
	if resp.ErrorResponse != nil {
		code := strings.TrimSpace(resp.ErrorResponse.Code)
		msg := strings.TrimSpace(resp.ErrorResponse.SubMsg)
		if msg == "" {
			msg = strings.TrimSpace(resp.ErrorResponse.Msg)
		}
		if msg == "" {
			msg = "未知错误"
		}
		return AliIdentity{}, fmt.Errorf("支付宝 oauth 失败[%s]%s", code, msg)
	}
	out := AliIdentity{
		UserID:      strings.TrimSpace(resp.Response.UserId),
		OpenID:      strings.TrimSpace(resp.Response.OpenId),
		AccessToken: strings.TrimSpace(resp.Response.AccessToken),
	}
	if out.UserID == "" && out.OpenID == "" {
		return AliIdentity{}, fmt.Errorf("支付宝 oauth 未返回用户标识")
	}
	return out, nil
}
