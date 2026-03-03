package plugin

import (
	"okpay/payment/plugin/contract"
	"okpay/payment/plugin/host"
	"okpay/payment/plugin/sdk"
)

// ---- Contract types ---------------------------------------------------

type PaymentChannel = contract.PaymentChannel
type PluginInfo = contract.PluginInfo
type InputField = contract.InputField
type CallRequest = contract.CallRequest
type HTTPRequest = contract.HTTPRequest
type RPCPlugin = contract.RPCPlugin
type RPCServer = contract.RPCServer
type RPCClient = contract.RPCClient
type InvokeArgs = contract.InvokeArgs

const (
	PluginName = contract.PluginName
)

var (
	HandshakeConfig     = contract.HandshakeConfig
	ErrNoImplementation = contract.ErrNoImplementation
)

// ---- SDK types --------------------------------------------------------

type HandlerFunc = sdk.HandlerFunc
type ServeOption = sdk.ServeOption

type OrderPayload = sdk.OrderPayload
type RefundPayload = sdk.RefundPayload
type TransferPayload = sdk.TransferPayload
type ChannelPayload = sdk.ChannelPayload

type RequestStats = sdk.RequestStats
type QueryStateResponse = sdk.QueryStateResponse
type RefundStateResponse = sdk.RefundStateResponse
type TransferStateResponse = sdk.TransferStateResponse
type NotifyResponse = sdk.NotifyResponse

type CompleteOrderRequest = sdk.CompleteOrderRequest
type CompleteRefundRequest = sdk.CompleteRefundRequest
type CompleteTransferRequest = sdk.CompleteTransferRequest
type CompleteCNotifyRequest = sdk.CompleteCNotifyRequest

type HTTPClient = sdk.HTTPClient
type HTTPClientConfig = sdk.HTTPClientConfig

const (
	BizTypeOrder    = contract.BizTypeOrder
	BizTypeRefund   = contract.BizTypeRefund
	BizTypeTransfer = contract.BizTypeTransfer
)

const (
	ResponseTypeJump  = sdk.ResponseTypeJump
	ResponseTypeHTML  = sdk.ResponseTypeHTML
	ResponseTypeJSON  = sdk.ResponseTypeJSON
	ResponseTypePage  = sdk.ResponseTypePage
	ResponseTypeError = sdk.ResponseTypeError
)

// ---- Host types -------------------------------------------------------

type Manager = host.Manager
type Option = host.Option
type CallObserver = host.CallObserver

const (
	CallTimeout = host.CallTimeout
)
