package plugin

import (
	"context"
	"io"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/ppswws/okpay-plugin-sdk/contract"
	"github.com/ppswws/okpay-plugin-sdk/host"
	"github.com/ppswws/okpay-plugin-sdk/proto"
	"github.com/ppswws/okpay-plugin-sdk/sdk"
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
type HandleRequest = proto.HandleRequest
type HandleResponse = proto.HandleResponse
type BizRequest = proto.BizRequest
type BizResult = proto.BizResult
type BizType = proto.BizType
type BizState = proto.BizState
type RequestTrace = proto.RequestTrace
type PageResponse = proto.PageResponse

type Manifest = sdk.Manifest
type InputSpec = sdk.InputSpec
type RequestStats = sdk.RequestStats
type BizResultInput = sdk.BizResultInput
type CompleteBizInput = sdk.CompleteBizInput
type HTTPClient = sdk.HTTPClient
type HTTPClientConfig = sdk.HTTPClientConfig
type CreateHandlerFunc = sdk.CreateHandlerFunc
type SubmitFormParams = sdk.SubmitFormParams

const (
	BizTypeInvalid  = proto.BizType_BIZ_TYPE_INVALID
	BizTypeOrder    = proto.BizType_BIZ_TYPE_ORDER
	BizTypeRefund   = proto.BizType_BIZ_TYPE_REFUND
	BizTypeTransfer = proto.BizType_BIZ_TYPE_TRANSFER
	BizTypeBalance  = proto.BizType_BIZ_TYPE_BALANCE

	BizStateInvalid    = proto.BizState_BIZ_STATE_INVALID
	BizStateFailed     = proto.BizState_BIZ_STATE_FAILED
	BizStateProcessing = proto.BizState_BIZ_STATE_PROCESSING
	BizStateSucceeded  = proto.BizState_BIZ_STATE_SUCCEEDED
)

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

func RespHTML(data string) *PageResponse { return sdk.RespHTML(data) }
func BuildSubmitHTML(params SubmitFormParams) (string, error) {
	return sdk.BuildSubmitHTML(params)
}
func RespJSON(data any) *PageResponse                       { return sdk.RespJSON(data) }
func RespError(msg string) *PageResponse                    { return sdk.RespError(msg) }
func RespPage(page string) *PageResponse                    { return sdk.RespPage(page) }
func RespJump(url string) *PageResponse                     { return sdk.RespJump(url) }
func RespPageURL(page, url string) *PageResponse            { return sdk.RespPageURL(page, url) }
func RespPageData(page string, data any) *PageResponse      { return sdk.RespPageData(page, data) }
func RespPageFull(page, url string, data any) *PageResponse { return sdk.RespPageFull(page, url, data) }
func ResultOK(input BizResultInput) *BizResult {
	return sdk.ResultOK(input)
}
func ResultPending(input BizResultInput) *BizResult {
	return sdk.ResultPending(input)
}
func ResultFail(input BizResultInput) *BizResult {
	return sdk.ResultFail(input)
}
func ResultBal(input BizResultInput) *BizResult {
	return sdk.ResultBal(input)
}

func CompleteBiz(ctx context.Context, req CompleteBizInput) error {
	return sdk.CompleteBiz(ctx, req)
}

func LockOrderExt(ctx context.Context, tradeNo string, fetch func() (any, RequestStats, error)) (map[string]any, error) {
	return sdk.LockOrderExt(ctx, tradeNo, fetch)
}
func BuildReturnMap(page *PageResponse) map[string]any     { return sdk.BuildReturnMap(page) }
func BuildReturnPage(payload map[string]any) *PageResponse { return sdk.BuildReturnPage(payload) }

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
