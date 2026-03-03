package plugin

import (
	"context"
	"time"

	"okpay/payment/plugin/sdk"
)

// ---- Serve ------------------------------------------------------------

func Serve(funcs map[string]HandlerFunc, opts ...ServeOption) {
	sdk.Serve(funcs, opts...)
}

func WithServeCallTimeout(timeout time.Duration) ServeOption {
	return sdk.WithServeCallTimeout(timeout)
}

// ---- Decode helpers ---------------------------------------------------

func DecodeOrder(raw any) *OrderPayload {
	return sdk.DecodeOrder(raw)
}

func DecodeRefund(raw any) *RefundPayload {
	return sdk.DecodeRefund(raw)
}

func DecodeTransfer(raw any) *TransferPayload {
	return sdk.DecodeTransfer(raw)
}

func DecodeChannel(raw any) *ChannelPayload {
	return sdk.DecodeChannel(raw)
}

func DecodeConfig(req *CallRequest) map[string]any {
	return sdk.DecodeConfig(req)
}

// ---- Response helpers -------------------------------------------------

func RespHTML(data string) map[string]any {
	return sdk.RespHTML(data)
}

func RespHTMLWithSubmit(data string, submit bool) map[string]any {
	return sdk.RespHTMLWithSubmit(data, submit)
}

func RespJSON(data any) map[string]any {
	return sdk.RespJSON(data)
}

func RespError(msg string) map[string]any {
	return sdk.RespError(msg)
}

func RespPage(page string) map[string]any {
	return sdk.RespPage(page)
}

func RespJump(url string) map[string]any {
	return sdk.RespJump(url)
}

func RespJumpWithSubmit(url string, submit bool) map[string]any {
	return sdk.RespJumpWithSubmit(url, submit)
}

func RespPageURL(page, url string) map[string]any {
	return sdk.RespPageURL(page, url)
}

func RespPageData(page string, data any) map[string]any {
	return sdk.RespPageData(page, data)
}

func RespPageFull(page, url string, data any) map[string]any {
	return sdk.RespPageFull(page, url, data)
}

func RespQuery(data QueryStateResponse) map[string]any {
	return sdk.RespQuery(data)
}

func RespRefund(data RefundStateResponse) map[string]any {
	return sdk.RespRefund(data)
}

func RespTransfer(data TransferStateResponse) map[string]any {
	return sdk.RespTransfer(data)
}

func RespNotify(ctx context.Context, call *CallRequest, data NotifyResponse) (map[string]any, error) {
	return sdk.RespNotify(ctx, call, data)
}

// ---- Create & lock helpers -------------------------------------------

func CreateWithHandlers(ctx context.Context, req *CallRequest, handlers map[string]HandlerFunc) (map[string]any, error) {
	return sdk.CreateWithHandlers(ctx, req, handlers)
}

func LockOrderExt(ctx context.Context, call *CallRequest, tradeNo string, fetch func() (any, RequestStats, error)) (map[string]any, error) {
	return sdk.LockOrderExt(ctx, call, tradeNo, fetch)
}

// ---- Complete callbacks ----------------------------------------------

func CompleteOrder(ctx context.Context, call *CallRequest, req CompleteOrderRequest) error {
	return sdk.CompleteOrder(ctx, call, req)
}

func CompleteRefund(ctx context.Context, call *CallRequest, req CompleteRefundRequest) error {
	return sdk.CompleteRefund(ctx, call, req)
}

func CompleteTransfer(ctx context.Context, call *CallRequest, req CompleteTransferRequest) error {
	return sdk.CompleteTransfer(ctx, call, req)
}

func CompleteCNotify(ctx context.Context, call *CallRequest, req CompleteCNotifyRequest) error {
	return sdk.CompleteCNotify(ctx, call, req)
}

// ---- HTTP client ------------------------------------------------------

func NewHTTPClient(cfg HTTPClientConfig) *HTTPClient {
	return sdk.NewHTTPClient(cfg)
}

// ---- Misc -------------------------------------------------------------

func IsWeChat(ua string) bool {
	return sdk.IsWeChat(ua)
}

func IsAlipay(ua string) bool {
	return sdk.IsAlipay(ua)
}

func IsMobileQQ(ua string) bool {
	return sdk.IsMobileQQ(ua)
}

func IsMobile(ua string) bool {
	return sdk.IsMobile(ua)
}

func ReadStringSlice(value any) []string {
	return sdk.ReadStringSlice(value)
}

func String(value any) string {
	return sdk.String(value)
}

func DecodeJSONMap(raw string) (map[string]any, error) {
	return sdk.DecodeJSONMap(raw)
}

func ModeSet(values []string) map[string]bool {
	return sdk.ModeSet(values)
}

func AllowMode(modes map[string]bool, code string) bool {
	return sdk.AllowMode(modes, code)
}
