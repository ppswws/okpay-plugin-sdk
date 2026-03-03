package wechatpay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-pay/gopay/wechat"

	"okpay/payment/plugin/sdk"
)

const defaultOAuthScope = "snsapi_base"

var httpClient = sdk.NewHTTPClient(sdk.HTTPClientConfig{})

type miniProgramSessionResp struct {
	OpenID  string `json:"openid"`
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type miniProgramAccessTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type miniProgramSchemeResp struct {
	OpenLink string `json:"openlink"`
	ErrCode  int    `json:"errcode"`
	ErrMsg   string `json:"errmsg"`
}

type MPAuthParams struct {
	AppID       string
	AppSecret   string
	Code        string
	RedirectURL string
	State       string
}

type MiniAuthParams struct {
	AppID     string
	AppSecret string
	Code      string
}

// GetOpenid handles full oauth flow for public accounts.
// If code is empty, it returns authURL for redirect.
// If code is present, it returns openid.
func GetOpenid(ctx context.Context, params MPAuthParams) (string, string, error) {
	code := strings.TrimSpace(params.Code)
	if code == "" {
		appID := strings.TrimSpace(params.AppID)
		redirectURL := strings.TrimSpace(params.RedirectURL)
		if appID == "" || redirectURL == "" {
			return "", "", fmt.Errorf("redirect url 不能为空")
		}
		state := strings.TrimSpace(params.State)
		base := "https://open.weixin.qq.com/connect/oauth2/authorize"
		authURL := fmt.Sprintf(
			"%s?appid=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s#wechat_redirect",
			base,
			url.QueryEscape(appID),
			url.QueryEscape(redirectURL),
			url.QueryEscape(defaultOAuthScope),
			url.QueryEscape(state),
		)
		return "", authURL, nil
	}
	resp, err := getOpenidFromMp(ctx, params.AppID, params.AppSecret, code)
	if err != nil {
		return "", "", err
	}
	openid := strings.TrimSpace(resp.Openid)
	if openid == "" {
		return "", "", fmt.Errorf("openid 为空")
	}
	return openid, "", nil
}

// getOpenidFromMp exchanges OAuth2 code for access token (公众号), returns gopay response.
func getOpenidFromMp(ctx context.Context, appID, appSecret, code string) (*wechat.Oauth2AccessToken, error) {
	appID = strings.TrimSpace(appID)
	appSecret = strings.TrimSpace(appSecret)
	code = strings.TrimSpace(code)
	if appID == "" || appSecret == "" || code == "" {
		return nil, fmt.Errorf("公众号参数缺失")
	}
	resp, err := wechat.GetOauth2AccessToken(ctx, appID, appSecret, code)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("oauth2 响应为空")
	}
	if resp.Errcode != 0 {
		return nil, fmt.Errorf("获取 openid 失败: %s", resp.Errmsg)
	}
	return resp, nil
}

// AppGetOpenid returns mini-program openid via jscode2session.
func AppGetOpenid(ctx context.Context, params MiniAuthParams) (string, error) {
	appID := strings.TrimSpace(params.AppID)
	appSecret := strings.TrimSpace(params.AppSecret)
	code := strings.TrimSpace(params.Code)
	if appID == "" || appSecret == "" || code == "" {
		return "", fmt.Errorf("小程序参数缺失")
	}
	endpoint := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		url.QueryEscape(appID),
		url.QueryEscape(appSecret),
		url.QueryEscape(code),
	)
	body, _, _, err := httpClient.Do(ctx, "GET", endpoint, "", "")
	if err != nil {
		return "", err
	}
	resp := miniProgramSessionResp{}
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

// GenerateScheme creates a mini-program scheme for browser -> mini-program jump.
func GenerateScheme(ctx context.Context, appID, appSecret, path, query string) (string, error) {
	appID = strings.TrimSpace(appID)
	appSecret = strings.TrimSpace(appSecret)
	if appID == "" || appSecret == "" {
		return "", fmt.Errorf("小程序参数缺失")
	}
	token, err := getMiniProgramAccessToken(ctx, appID, appSecret)
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
	body, _ := json.Marshal(payload)
	endpoint := fmt.Sprintf("https://api.weixin.qq.com/wxa/generatescheme?access_token=%s", url.QueryEscape(token))
	respBody, _, _, err := httpClient.Do(ctx, "POST", endpoint, string(body), "application/json")
	if err != nil {
		return "", err
	}
	resp := miniProgramSchemeResp{}
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

func getMiniProgramAccessToken(ctx context.Context, appID, appSecret string) (string, error) {
	endpoint := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		url.QueryEscape(appID),
		url.QueryEscape(appSecret),
	)
	body, _, _, err := httpClient.Do(ctx, "GET", endpoint, "", "")
	if err != nil {
		return "", err
	}
	resp := miniProgramAccessTokenResp{}
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
