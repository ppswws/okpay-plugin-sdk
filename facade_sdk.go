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

func Order(req *InvokeRequestV2) *OrderPayload {
	return sdk.Order(req)
}

func Refund(req *InvokeRequestV2) *RefundPayload {
	return sdk.Refund(req)
}

func Transfer(req *InvokeRequestV2) *TransferPayload {
	return sdk.Transfer(req)
}

func Channel(req *InvokeRequestV2) *ChannelPayload {
	return sdk.Channel(req)
}

func ChannelConfig(req *InvokeRequestV2) map[string]any {
	return sdk.ChannelConfig(req)
}

func GlobalConfig(req *InvokeRequestV2) map[string]any {
	return sdk.GlobalConfig(req)
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

func RespBalance(balance string) map[string]any {
	return sdk.RespBalance(balance)
}

func RespNotify(ctx context.Context, call *InvokeRequestV2, data NotifyResponse) (map[string]any, error) {
	return sdk.RespNotify(ctx, call, data)
}

// ---- Create & lock helpers -------------------------------------------

func CreateWithHandlers(ctx context.Context, req *InvokeRequestV2, handlers map[string]HandlerFunc) (map[string]any, error) {
	return sdk.CreateWithHandlers(ctx, req, handlers)
}

func LockOrderExt(ctx context.Context, call *InvokeRequestV2, tradeNo string, fetch func() (any, RequestStats, error)) (map[string]any, error) {
	return sdk.LockOrderExt(ctx, call, tradeNo, fetch)
}

// ---- Complete callbacks ----------------------------------------------

func CompleteOrder(ctx context.Context, call *InvokeRequestV2, req CompleteOrderRequest) error {
	return sdk.CompleteOrder(ctx, call, req)
}

func CompleteRefund(ctx context.Context, call *InvokeRequestV2, req CompleteRefundRequest) error {
	return sdk.CompleteRefund(ctx, call, req)
}

func CompleteTransfer(ctx context.Context, call *InvokeRequestV2, req CompleteTransferRequest) error {
	return sdk.CompleteTransfer(ctx, call, req)
}

func CompleteCNotify(ctx context.Context, call *InvokeRequestV2, req CompleteCNotifyRequest) error {
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

func ParseRequestParams(req *InvokeRequestV2) map[string]string {
	return sdk.ParseRequestParams(req)
}

func QueryParam(req *InvokeRequestV2, key string) string {
	return sdk.QueryParam(req, key)
}

func MapString(m map[string]any, key string) string {
	return sdk.MapString(m, key)
}

func Read(req *InvokeRequestV2, path string) (Value, bool) {
	return sdk.Read(req, path)
}

func ReadStringSlice(value any) []string {
	return sdk.ReadStringSlice(value)
}

func DecodeLosslessJSON(raw []byte) (Value, error) {
	return sdk.DecodeLosslessJSON(raw)
}

func DecodeLosslessJSONObject(raw []byte) (*ObjectValue, error) {
	return sdk.DecodeLosslessJSONObject(raw)
}

func ValueToAny(v Value) (any, error) {
	return sdk.ValueToAny(v)
}

func AnyToValue(v any) (Value, error) {
	return sdk.AnyToValue(v)
}

func ValueMapToAnyMap(in map[string]Value) (map[string]any, error) {
	return sdk.ValueMapToAnyMap(in)
}

func ModeSet(values []string) map[string]bool {
	return sdk.ModeSet(values)
}

func AllowMode(modes map[string]bool, code string) bool {
	return sdk.AllowMode(modes, code)
}
