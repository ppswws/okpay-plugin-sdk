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
type RPCPlugin = contract.RPCPlugin
type RPCServer = contract.RPCServer
type RPCClient = contract.RPCClient
type InvokeV2Args = contract.InvokeV2Args
type InvokeRequestV2 = contract.InvokeRequestV2
type InvokeResponseV2 = contract.InvokeResponseV2
type PluginError = contract.PluginError
type RawEnvelope = contract.RawEnvelope
type ParsedEnvelope = contract.ParsedEnvelope
type HeaderKV = contract.HeaderKV
type Value = contract.Value
type ValueKind = contract.ValueKind
type ObjectValue = contract.ObjectValue

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

const (
	EventPayCreate      = contract.EventPayCreate
	EventPayNotify      = contract.EventPayNotify
	EventPayQuery       = contract.EventPayQuery
	EventRefundCreate   = contract.EventRefundCreate
	EventRefundNotify   = contract.EventRefundNotify
	EventTransferCreate = contract.EventTransferCreate
	EventTransferNotify = contract.EventTransferNotify
)

const (
	ValueKindNull    = contract.ValueKindNull
	ValueKindString  = contract.ValueKindString
	ValueKindBool    = contract.ValueKindBool
	ValueKindInt64   = contract.ValueKindInt64
	ValueKindUInt64  = contract.ValueKindUInt64
	ValueKindDecimal = contract.ValueKindDecimal
	ValueKindBytes   = contract.ValueKindBytes
	ValueKindObject  = contract.ValueKindObject
	ValueKindArray   = contract.ValueKindArray
)

// ---- Host types -------------------------------------------------------

type Manager = host.Manager
type Option = host.Option
type CallObserver = host.CallObserver

const (
	CallTimeout = host.CallTimeout
)
