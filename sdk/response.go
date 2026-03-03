package sdk

// Response type constants.
const (
	ResponseTypeJump  = "jump"
	ResponseTypeHTML  = "html"
	ResponseTypeJSON  = "json"
	ResponseTypePage  = "page"
	ResponseTypeError = "error"
)

// RespHTML builds a typed html response payload.
func RespHTML(data string) map[string]any {
	return map[string]any{
		"type": ResponseTypeHTML,
		"data": data,
	}
}

// RespHTMLWithSubmit builds a typed html response payload with submit flag.
func RespHTMLWithSubmit(data string, submit bool) map[string]any {
	out := RespHTML(data)
	if submit {
		out["submit"] = true
	}
	return out
}

// RespJSON builds a typed json response payload.
func RespJSON(data any) map[string]any {
	return map[string]any{
		"type": ResponseTypeJSON,
		"data": data,
	}
}

// RespError builds a typed error response payload.
func RespError(msg string) map[string]any {
	return map[string]any{
		"type": ResponseTypeError,
		"msg":  msg,
	}
}

// RespPage builds a typed page response payload.
func RespPage(page string) map[string]any {
	return map[string]any{
		"type": ResponseTypePage,
		"page": page,
	}
}

// RespJump builds a typed jump response payload.
func RespJump(url string) map[string]any {
	return map[string]any{
		"type": ResponseTypeJump,
		"url":  url,
	}
}

// RespJumpWithSubmit builds a typed jump payload with submit flag.
func RespJumpWithSubmit(url string, submit bool) map[string]any {
	out := RespJump(url)
	if submit {
		out["submit"] = true
	}
	return out
}

// RespPageURL builds a typed page payload with url.
func RespPageURL(page, url string) map[string]any {
	out := RespPage(page)
	if url != "" {
		out["url"] = url
	}
	return out
}

// RespPageData builds a typed page payload with data.
func RespPageData(page string, data any) map[string]any {
	out := RespPage(page)
	if data != nil {
		out["data"] = data
	}
	return out
}

// RespPageFull builds a typed page payload with url and data.
func RespPageFull(page, url string, data any) map[string]any {
	out := RespPageURL(page, url)
	if data != nil {
		out["data"] = data
	}
	return out
}

// QueryStateResponse is the payload model for query state.
type QueryStateResponse struct {
	State      int
	APITradeNo string
}

// RespQuery builds a query state response payload.
func RespQuery(data QueryStateResponse) map[string]any {
	return map[string]any{
		"state":        data.State,
		"api_trade_no": data.APITradeNo,
	}
}

// RefundStateResponse is the payload model for refund state.
type RefundStateResponse struct {
	State       int
	APIRefundNo string
	ReqBody     string
	RespBody    string
	Result      string
	ReqMs       int32
}

// RespRefund builds a refund state response payload.
func RespRefund(data RefundStateResponse) map[string]any {
	return map[string]any{
		"state":         data.State,
		"api_refund_no": data.APIRefundNo,
		"req_body":      data.ReqBody,
		"resp_body":     data.RespBody,
		"result":        data.Result,
		"req_ms":        data.ReqMs,
	}
}

// TransferStateResponse is the payload model for transfer state.
type TransferStateResponse struct {
	State      int
	APITradeNo string
	ReqBody    string
	RespBody   string
	Result     string
	ReqMs      int32
}

// RespTransfer builds a transfer state response payload.
func RespTransfer(data TransferStateResponse) map[string]any {
	return map[string]any{
		"state":        data.State,
		"api_trade_no": data.APITradeNo,
		"req_body":     data.ReqBody,
		"resp_body":    data.RespBody,
		"result":       data.Result,
		"req_ms":       data.ReqMs,
	}
}
