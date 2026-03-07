package sdk

import (
	"encoding/json"

	"okpay/payment/plugin/proto"
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

// BuildSubmitHTML wraps a form HTML payload with auto-submit script.
func BuildSubmitHTML(data string) string {
	return "<!DOCTYPE html><html><head><meta charset=\"utf-8\"></head><body>" +
		data +
		"<script>document.forms[0] && document.forms[0].submit();</script>" +
		"</body></html>"
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

func RespQuery(state int, apiTradeNo string) *proto.QueryResponse {
	return &proto.QueryResponse{State: int32(state), ApiTradeNo: apiTradeNo}
}

func RespRefund(state int, apiRefundNo, reqBody, respBody, result string, reqMs int32) *proto.RefundResponse {
	return &proto.RefundResponse{
		State:       int32(state),
		ApiRefundNo: apiRefundNo,
		ReqBody:     reqBody,
		RespBody:    respBody,
		Result:      result,
		ReqMs:       reqMs,
	}
}

func RespTransfer(state int, apiTradeNo, reqBody, respBody, result string, reqMs int32) *proto.TransferResponse {
	return &proto.TransferResponse{
		State:      int32(state),
		ApiTradeNo: apiTradeNo,
		ReqBody:    reqBody,
		RespBody:   respBody,
		Result:     result,
		ReqMs:      reqMs,
	}
}

func RespBalance(balance string) *proto.BalanceResponse {
	return &proto.BalanceResponse{Balance: balance}
}
