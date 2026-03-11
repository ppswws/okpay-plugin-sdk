package plugin

import (
	"context"
	"io"
	"time"

	"github.com/hashicorp/go-hclog"
	"okpay/payment/plugin/contract"
	"okpay/payment/plugin/host"
	"okpay/payment/plugin/proto"
	"okpay/payment/plugin/sdk"
)

// Contract and host types.
type PluginService = contract.PluginService
type KernelService = contract.KernelService
type PluginInfo = contract.PluginInfo
type InputField = contract.InputField
type Manager = host.Manager
type Option = host.Option
type CallObserver = host.CallObserver

const (
	PluginName  = contract.PluginName
	CallTimeout = host.CallTimeout

	BizTypeOrder    = contract.BizTypeOrder
	BizTypeRefund   = contract.BizTypeRefund
	BizTypeTransfer = contract.BizTypeTransfer

	ResponseTypeJump  = sdk.ResponseTypeJump
	ResponseTypeHTML  = sdk.ResponseTypeHTML
	ResponseTypeJSON  = sdk.ResponseTypeJSON
	ResponseTypePage  = sdk.ResponseTypePage
	ResponseTypeError = sdk.ResponseTypeError
)

var (
	HandshakeConfig     = contract.HandshakeConfig
	ErrNoImplementation = contract.ErrNoImplementation
)

// Proto aliases exposed to plugins.
type InvokeContext = proto.InvokeContext
type PluginInfoResponse = proto.PluginInfoResponse
type CreateRequest = proto.CreateRequest
type CreateResponse = proto.CreateResponse
type QueryRequest = proto.QueryRequest
type QueryResponse = proto.QueryResponse
type RefundRequest = proto.RefundRequest
type RefundResponse = proto.RefundResponse
type TransferRequest = proto.TransferRequest
type TransferResponse = proto.TransferResponse
type BalanceRequest = proto.BalanceRequest
type BalanceResponse = proto.BalanceResponse
type InvokeFuncRequest = proto.InvokeFuncRequest
type InvokeFuncResponse = proto.InvokeFuncResponse
type PageResponse = proto.PageResponse

type Manifest = sdk.Manifest
type InputSpec = sdk.InputSpec
type RequestStats = sdk.RequestStats
type CompleteOrderInput = sdk.CompleteOrderInput
type CompleteRefundInput = sdk.CompleteRefundInput
type CompleteTransferInput = sdk.CompleteTransferInput
type CompleteCNotifyInput = sdk.CompleteCNotifyInput
type HTTPClient = sdk.HTTPClient
type HTTPClientConfig = sdk.HTTPClientConfig
type CreateHandlerFunc = sdk.CreateHandlerFunc

func NewManager(opts ...Option) (*Manager, error)  { return host.NewManager(opts...) }
func WithDir(dir string) Option                    { return host.WithDir(dir) }
func WithCallTimeout(timeout time.Duration) Option { return host.WithCallTimeout(timeout) }
func WithPluginLogWriters(stdout, stderr io.Writer) Option {
	return host.WithPluginLogWriters(stdout, stderr)
}
func WithCallObserver(observer CallObserver) Option { return host.WithCallObserver(observer) }
func WithPluginLogger(logger hclog.Logger) Option   { return host.WithPluginLogger(logger) }
func WithKernelService(kernel KernelService) Option { return host.WithKernelService(kernel) }

func Serve(impl PluginService) error { return sdk.Serve(impl) }

func BuildInfoResponse(manifest Manifest) (*PluginInfoResponse, error) {
	return sdk.BuildInfoResponse(manifest)
}

func RespHTML(data string) *PageResponse                    { return sdk.RespHTML(data) }
func BuildSubmitHTML(data string) string                    { return sdk.BuildSubmitHTML(data) }
func RespJSON(data any) *PageResponse                       { return sdk.RespJSON(data) }
func RespError(msg string) *PageResponse                    { return sdk.RespError(msg) }
func RespPage(page string) *PageResponse                    { return sdk.RespPage(page) }
func RespJump(url string) *PageResponse                     { return sdk.RespJump(url) }
func RespPageURL(page, url string) *PageResponse            { return sdk.RespPageURL(page, url) }
func RespPageData(page string, data any) *PageResponse      { return sdk.RespPageData(page, data) }
func RespPageFull(page, url string, data any) *PageResponse { return sdk.RespPageFull(page, url, data) }
func RespQuery(state int, apiTradeNo string) *QueryResponse { return sdk.RespQuery(state, apiTradeNo) }
func RespRefund(state int, apiRefundNo, reqBody, respBody, result string, reqMs int32) *RefundResponse {
	return sdk.RespRefund(state, apiRefundNo, reqBody, respBody, result, reqMs)
}
func RespTransfer(state int, apiTradeNo, reqBody, respBody, result string, reqMs int32) *TransferResponse {
	return sdk.RespTransfer(state, apiTradeNo, reqBody, respBody, result, reqMs)
}
func RespBalance(balance string) *BalanceResponse { return sdk.RespBalance(balance) }

func RecordNotify(ctx context.Context, req *InvokeContext, bizType string, result *PageResponse) *PageResponse {
	return sdk.RecordNotify(ctx, req, bizType, result)
}

func CompleteOrder(ctx context.Context, req CompleteOrderInput) error {
	return sdk.CompleteOrder(ctx, req)
}
func CompleteRefund(ctx context.Context, req CompleteRefundInput) error {
	return sdk.CompleteRefund(ctx, req)
}
func CompleteTransfer(ctx context.Context, req CompleteTransferInput) error {
	return sdk.CompleteTransfer(ctx, req)
}
func CompleteCNotify(ctx context.Context, req CompleteCNotifyInput) error {
	return sdk.CompleteCNotify(ctx, req)
}

func LockOrderExt(ctx context.Context, tradeNo string, fetch func() (any, RequestStats, error)) (map[string]any, error) {
	return sdk.LockOrderExt(ctx, tradeNo, fetch)
}

func CreateWithHandlers(ctx context.Context, req *InvokeContext, handlers map[string]CreateHandlerFunc) (*PageResponse, error) {
	return sdk.CreateWithHandlers(ctx, req, handlers)
}

func NewHTTPClient(cfg HTTPClientConfig) *HTTPClient { return sdk.NewHTTPClient(cfg) }
func IsWeChat(ua string) bool                        { return sdk.IsWeChat(ua) }
func IsAlipay(ua string) bool                        { return sdk.IsAlipay(ua) }
func IsMobileQQ(ua string) bool                      { return sdk.IsMobileQQ(ua) }
func IsMobile(ua string) bool                        { return sdk.IsMobile(ua) }
func BuildMPOAuthURL(appID, redirectURL, state string) string {
	return sdk.BuildMPOAuthURL(appID, redirectURL, state)
}
func GetMPOpenid(ctx context.Context, appID, appSecret, code string) (string, error) {
	return sdk.GetMPOpenid(ctx, appID, appSecret, code)
}
func GetMiniOpenid(ctx context.Context, appID, appSecret, code string) (string, error) {
	return sdk.GetMiniOpenid(ctx, appID, appSecret, code)
}
func GetMiniScheme(ctx context.Context, appID, appSecret, path, query string) (string, error) {
	return sdk.GetMiniScheme(ctx, appID, appSecret, path, query)
}
