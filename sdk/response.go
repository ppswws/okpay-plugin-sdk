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
	TypeJump  = "jump"
	TypeHTML  = "html"
	TypeJSON  = "json"
	TypePage  = "page"
	TypeError = "error"
)

func RespHTML(data string) *proto.PageResponse {
	return &proto.PageResponse{Type: TypeHTML, DataText: data}
}

// RecordNotify marks current page response as channel-notify for kernel cnotify logging.
// Biz type is inferred from invoke context snapshots: refund > transfer > order.
func RecordNotify(req *proto.InvokeContext, page *proto.PageResponse) *proto.PageResponse {
	if page == nil {
		page = RespError("empty notify response")
	}
	switch {
	case req != nil && req.GetRefund() != nil && strings.TrimSpace(req.GetRefund().GetRefundNo()) != "":
		page.NotifyBiz = "refund"
	case req != nil && req.GetTransfer() != nil && strings.TrimSpace(req.GetTransfer().GetTradeNo()) != "":
		page.NotifyBiz = "transfer"
	case req != nil && req.GetOrder() != nil && strings.TrimSpace(req.GetOrder().GetTradeNo()) != "":
		page.NotifyBiz = "order"
	default:
		page.NotifyBiz = ""
	}
	return page
}

// PostForm defines the action URL and fields for a POST auto-submit page.
type PostForm struct {
	ActionURL string
	Fields    map[string][]string
}

// BuildPostHTML builds an auto-submit POST HTML page from structured params.
func BuildPostHTML(params PostForm) (string, error) {
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
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     u.Path,
		RawQuery: u.RawQuery,
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
	return &proto.PageResponse{Type: TypeJSON, DataRaw: raw}
}

func RespError(msg string) *proto.PageResponse {
	return &proto.PageResponse{Type: TypeError, Msg: msg}
}

func RespPage(page string) *proto.PageResponse {
	return &proto.PageResponse{Type: TypePage, Page: page}
}

func RespJump(url string) *proto.PageResponse {
	return &proto.PageResponse{Type: TypeJump, Url: url}
}

func RespPageURL(page, url string) *proto.PageResponse {
	return &proto.PageResponse{Type: TypePage, Page: page, Url: url}
}

func RespPageData(page string, data any) *proto.PageResponse {
	raw, _ := json.Marshal(data)
	return &proto.PageResponse{Type: TypePage, Page: page, DataRaw: raw}
}

func RespPageFull(page, url string, data any) *proto.PageResponse {
	raw, _ := json.Marshal(data)
	return &proto.PageResponse{Type: TypePage, Page: page, Url: url, DataRaw: raw}
}

// BizOut is the single named-field input for plugin business results.
type BizOut struct {
	ApiNo   string
	Code    string
	Msg     string
	Buyer   string
	Balance string
	Stats   RequestStats
}

func Result(state proto.BizState, input BizOut) *proto.BizResult {
	apiBizNo := input.ApiNo
	if state == proto.BizState_S_FAIL {
		apiBizNo = ""
	}
	return buildResult(state, apiBizNo, input.Code, input.Msg, input.Buyer, "", input.Stats)
}

func ResultBal(input BizOut) *proto.BizResult {
	return buildResult(proto.BizState_S_OK, "", input.Code, input.Msg, "", input.Balance, input.Stats)
}

func buildResult(state proto.BizState, apiNo, code, msg, buyer, balance string, stats RequestStats) *proto.BizResult {
	return &proto.BizResult{
		State:   state,
		ApiNo:   apiNo,
		Code:    code,
		Msg:     msg,
		Buyer:   buyer,
		Balance: balance,
		Trace: &proto.RequestTrace{
			ReqMs:    stats.ReqMs,
			ReqBody:  stats.ReqBody,
			RespBody: stats.RespBody,
		},
	}
}
