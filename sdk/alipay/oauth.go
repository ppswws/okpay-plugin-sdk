package alipay

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
)

const defaultOAuthScope = "auth_base"

// BuildOAuthURL builds the Alipay oauth URL for public account auth.
func BuildOAuthURL(appID, redirectURL, state string) string {
	appID = strings.TrimSpace(appID)
	redirectURL = strings.TrimSpace(redirectURL)
	if appID == "" || redirectURL == "" {
		return ""
	}
	params := url.Values{}
	params.Set("app_id", appID)
	params.Set("scope", defaultOAuthScope)
	params.Set("redirect_uri", redirectURL)
	if strings.TrimSpace(state) != "" {
		params.Set("state", strings.TrimSpace(state))
	}
	return "https://openauth.alipay.com/oauth2/publicAppAuthorize.htm?" + params.Encode()
}

// ExchangeAuthCode exchanges auth_code for access token (gopay response).
func ExchangeAuthCode(ctx context.Context, appID, privateKey, authCode string, isProd bool) (*alipay.SystemOauthTokenResponse, error) {
	client, err := alipay.NewClient(appID, privateKey, isProd)
	if err != nil {
		return nil, err
	}
	client.SetCharset(alipay.UTF8).SetSignType(alipay.RSA2)
	bm := make(gopay.BodyMap)
	bm.Set("grant_type", "authorization_code")
	bm.Set("code", strings.TrimSpace(authCode))
	resp, err := client.SystemOauthToken(ctx, bm)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Response == nil {
		return nil, fmt.Errorf("oauth 响应为空")
	}
	return resp, nil
}

// RefreshToken refreshes access token with refresh_token (gopay response).
func RefreshToken(ctx context.Context, appID, privateKey, refreshToken string, isProd bool) (*alipay.SystemOauthTokenResponse, error) {
	client, err := alipay.NewClient(appID, privateKey, isProd)
	if err != nil {
		return nil, err
	}
	client.SetCharset(alipay.UTF8).SetSignType(alipay.RSA2)
	bm := make(gopay.BodyMap)
	bm.Set("grant_type", "refresh_token")
	bm.Set("refresh_token", strings.TrimSpace(refreshToken))
	resp, err := client.SystemOauthToken(ctx, bm)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Response == nil {
		return nil, fmt.Errorf("oauth 响应为空")
	}
	return resp, nil
}

// UserInfo fetches user info by access token (gopay response).
func UserInfo(ctx context.Context, appID, privateKey, accessToken string, isProd bool) (*alipay.UserInfoShareResponse, error) {
	client, err := alipay.NewClient(appID, privateKey, isProd)
	if err != nil {
		return nil, err
	}
	client.SetCharset(alipay.UTF8).SetSignType(alipay.RSA2)
	resp, err := client.UserInfoShare(ctx, strings.TrimSpace(accessToken))
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Response == nil {
		return nil, fmt.Errorf("用户信息为空")
	}
	return resp, nil
}
