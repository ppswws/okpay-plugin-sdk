package sdk

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"sort"
	"strings"

	"github.com/ppswws/okpay-plugin-sdk/proto"
)

// Response type constants.
const (
	ResponseTypeJump  = "jump"
	ResponseTypeHTML  = "html"
	ResponseTypeJSON  = "json"
	ResponseTypePage  = "page"
	ResponseTypeError = "error"
)

func RespHTML(data string) *proto.PageResponse {
	return &proto.PageResponse{Type: ResponseTypeHTML, DataText: data}
}

// SubmitFormParams defines the action URL and fields for a POST auto-submit page.
type SubmitFormParams struct {
	ActionURL string
	Fields    map[string][]string
}

// BuildSubmitHTML builds an auto-submit POST HTML page from structured params.
func BuildSubmitHTML(params SubmitFormParams) (string, error) {
	actionURL := strings.TrimSpace(params.ActionURL)
	if actionURL == "" {
		return "", fmt.Errorf("action url is empty")
	}
	u, err := url.Parse(actionURL)
	if err != nil {
		return "", fmt.Errorf("parse action url failed: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("action url is incomplete")
	}
	actionURL = (&url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   u.Path,
	}).String()
	if len(params.Fields) == 0 {
		return "", fmt.Errorf("submit fields are empty")
	}
	keys := make([]string, 0, len(params.Fields))
	for k := range params.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString(`<form id="okpay_submit_form" method="post" accept-charset="utf-8" enctype="application/x-www-form-urlencoded" action="`)
	b.WriteString(html.EscapeString(actionURL))
	b.WriteString(`">`)
	for _, k := range keys {
		for _, v := range params.Fields[k] {
			b.WriteString(`<input type="hidden" name="`)
			b.WriteString(html.EscapeString(k))
			b.WriteString(`" value="`)
			b.WriteString(html.EscapeString(v))
			b.WriteString(`">`)
		}
	}
	b.WriteString(`</form>`)
	return "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\"></head><body>" +
		b.String() +
		"<script>document.forms[0] && document.forms[0].submit();</script>" +
		"</body></html>", nil
}

func RespJSON(data any) *proto.PageResponse {
	raw, _ := json.Marshal(data)
	return &proto.PageResponse{Type: ResponseTypeJSON, DataJsonRaw: raw}
}

func RespError(msg string) *proto.PageResponse {
	return &proto.PageResponse{Type: ResponseTypeError, Msg: msg}
}

func RespPage(page string) *proto.PageResponse {
	return &proto.PageResponse{Type: ResponseTypePage, Page: page}
}

func RespJump(url string) *proto.PageResponse {
	return &proto.PageResponse{Type: ResponseTypeJump, Url: url}
}

func RespPageURL(page, url string) *proto.PageResponse {
	return &proto.PageResponse{Type: ResponseTypePage, Page: page, Url: url}
}

func RespPageData(page string, data any) *proto.PageResponse {
	raw, _ := json.Marshal(data)
	return &proto.PageResponse{Type: ResponseTypePage, Page: page, DataJsonRaw: raw}
}

func RespPageFull(page, url string, data any) *proto.PageResponse {
	raw, _ := json.Marshal(data)
	return &proto.PageResponse{Type: ResponseTypePage, Page: page, Url: url, DataJsonRaw: raw}
}

// BizResultInput is the single named-field input for plugin business results.
type BizResultInput struct {
	ApiNo   string
	Code    string
	Msg     string
	Balance string
	Stats   RequestStats
	// Legacy aliases for smooth migration.
	APIBizNo    string
	ChannelCode string
	ChannelMsg  string
}

func ResultOK(input BizResultInput) *proto.BizResult {
	return buildResult(proto.BizState_BIZ_STATE_SUCCEEDED, pick(input.ApiNo, input.APIBizNo), pick(input.Code, input.ChannelCode), pick(input.Msg, input.ChannelMsg), "", input.Stats)
}

func ResultPending(input BizResultInput) *proto.BizResult {
	return buildResult(proto.BizState_BIZ_STATE_PROCESSING, pick(input.ApiNo, input.APIBizNo), pick(input.Code, input.ChannelCode), pick(input.Msg, input.ChannelMsg), "", input.Stats)
}

func ResultFail(input BizResultInput) *proto.BizResult {
	return buildResult(proto.BizState_BIZ_STATE_FAILED, "", pick(input.Code, input.ChannelCode), pick(input.Msg, input.ChannelMsg), "", input.Stats)
}

func ResultBal(input BizResultInput) *proto.BizResult {
	return buildResult(proto.BizState_BIZ_STATE_SUCCEEDED, "", pick(input.Code, input.ChannelCode), pick(input.Msg, input.ChannelMsg), input.Balance, input.Stats)
}

func pick(short, legacy string) string {
	if short != "" {
		return short
	}
	return legacy
}

func buildResult(state proto.BizState, apiBizNo, channelCode, channelMsg, balance string, stats RequestStats) *proto.BizResult {
	return &proto.BizResult{
		State:       state,
		ApiBizNo:    apiBizNo,
		ChannelCode: channelCode,
		ChannelMsg:  channelMsg,
		Balance:     balance,
		Trace: &proto.RequestTrace{
			ReqMs:    stats.ReqMs,
			ReqBody:  stats.ReqBody,
			RespBody: stats.RespBody,
		},
	}
}
