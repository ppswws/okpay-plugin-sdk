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

	TypeJump  = sdk.TypeJump
	TypeHTML  = sdk.TypeHTML
	TypeJSON  = sdk.TypeJSON
	TypePage  = sdk.TypePage
	TypeError = sdk.TypeError
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
type InSpec = sdk.InSpec
type RequestStats = sdk.RequestStats
type BizOut = sdk.BizOut
type BizDoneIn = sdk.BizDoneIn
type HTTPClient = sdk.HTTPClient
type HTTPClientConfig = sdk.HTTPClientConfig
type CreateHandlerFunc = sdk.CreateHandlerFunc
type PostForm = sdk.PostForm
type AliIdentity = sdk.AliIdentity

const (
	BizTypeInvalid  = proto.BizType_T_NONE
	BizTypeOrder    = proto.BizType_T_PAY
	BizTypeRefund   = proto.BizType_T_REF
	BizTypeTransfer = proto.BizType_T_XFER
	BizTypeBalance  = proto.BizType_T_BAL

	BizStateInvalid    = proto.BizState_S_NONE
	BizStateFailed     = proto.BizState_S_FAIL
	BizStateProcessing = proto.BizState_S_ING
	BizStateSucceeded  = proto.BizState_S_OK
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

func BuildInfo(manifest Manifest) (*PluginInfoResponse, error) {
	return sdk.BuildInfo(manifest)
}

func RespHTML(data string) *PageResponse { return sdk.RespHTML(data) }
func RecordNotify(req *InvokeContext, page *PageResponse) *PageResponse {
	return sdk.RecordNotify(req, page)
}
func BuildPostHTML(params PostForm) (string, error) {
	return sdk.BuildPostHTML(params)
}
func RespJSON(data any) *PageResponse                       { return sdk.RespJSON(data) }
func RespError(msg string) *PageResponse                    { return sdk.RespError(msg) }
func RespPage(page string) *PageResponse                    { return sdk.RespPage(page) }
func RespJump(url string) *PageResponse                     { return sdk.RespJump(url) }
func RespPageURL(page, url string) *PageResponse            { return sdk.RespPageURL(page, url) }
func RespPageData(page string, data any) *PageResponse      { return sdk.RespPageData(page, data) }
func RespPageFull(page, url string, data any) *PageResponse { return sdk.RespPageFull(page, url, data) }
func Result(state BizState, input BizOut) *BizResult {
	return sdk.Result(state, input)
}
func ResultBal(input BizOut) *BizResult {
	return sdk.ResultBal(input)
}

func CompleteBiz(ctx context.Context, req BizDoneIn) error {
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
func BuildAliOAuthURL(appID, redirectURL, state string, isProd bool) string {
	return sdk.BuildAliOAuthURL(appID, redirectURL, state, isProd)
}
func GetAliIdentity(ctx context.Context, appID, privateKey, authCode string, isProd bool) (AliIdentity, error) {
	return sdk.GetAliIdentity(ctx, appID, privateKey, authCode, isProd)
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
