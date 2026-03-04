package contract

import "strings"

// Event constants used by InvokeRequestV2.Event.
const (
	EventPayCreate      = "pay.create"
	EventPayNotify      = "pay.notify"
	EventPayQuery       = "pay.query"
	EventRefundCreate   = "refund.create"
	EventRefundNotify   = "refund.notify"
	EventTransferCreate = "transfer.create"
	EventTransferNotify = "transfer.notify"
)

// InvokeRequestV2 is the new lossless envelope for plugin invocation.
// It keeps raw bytes and parsed values side-by-side to avoid signature drift.
type InvokeRequestV2 struct {
	Version string         `json:"version,omitempty"`
	TraceID string         `json:"trace_id,omitempty"`
	Action  string         `json:"action,omitempty"` // free-form handler name
	Event   string         `json:"event,omitempty"`  // normalized event class
	Route   string         `json:"route,omitempty"`  // callback route discriminator
	Raw     RawEnvelope    `json:"raw"`
	Parsed  ParsedEnvelope `json:"parsed"`
}

// InvokeResponseV2 is the standardized plugin return model.
type InvokeResponseV2 struct {
	OK      bool              `json:"ok"`
	Error   *PluginError      `json:"error,omitempty"`
	Present map[string]Value  `json:"present,omitempty"`
	Effect  map[string]Value  `json:"effect,omitempty"`
	Result  map[string]Value  `json:"result,omitempty"`
	Debug   map[string]string `json:"debug,omitempty"`
}

// PluginError marks system/business/retryable failures.
type PluginError struct {
	Kind    string            `json:"kind,omitempty"` // system/business/retryable
	Code    string            `json:"code,omitempty"`
	Message string            `json:"message,omitempty"`
	Detail  map[string]string `json:"detail,omitempty"`
}

// RawEnvelope stores exact wire/raw data for signature and forensic checks.
type RawEnvelope struct {
	HTTPMethod       string     `json:"http_method,omitempty"`
	HTTPURL          string     `json:"http_url,omitempty"`
	HTTPQueryRaw     string     `json:"http_query_raw,omitempty"`
	HTTPBodyRaw      []byte     `json:"http_body_raw,omitempty"`
	HTTPHeadersRaw   []HeaderKV `json:"http_headers_raw,omitempty"`
	RequestIP        string     `json:"request_ip,omitempty"`
	UserAgent        string     `json:"user_agent,omitempty"`
	ChannelReqRaw    []byte     `json:"channel_req_raw,omitempty"`
	ChannelRespRaw   []byte     `json:"channel_resp_raw,omitempty"`
	ChannelNotifyRaw []byte     `json:"channel_notify_raw,omitempty"`
}

// ParsedEnvelope stores lossless structured data.
type ParsedEnvelope struct {
	Data ObjectValue `json:"data"`
}

// HeaderKV keeps header key/value as-is and in-order.
type HeaderKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ValueKind is a lossless value type marker.
type ValueKind string

const (
	ValueKindNull    ValueKind = "null"
	ValueKindString  ValueKind = "string"
	ValueKindBool    ValueKind = "bool"
	ValueKindInt64   ValueKind = "int64"
	ValueKindUInt64  ValueKind = "uint64"
	ValueKindDecimal ValueKind = "decimal"
	ValueKindBytes   ValueKind = "bytes"
	ValueKindObject  ValueKind = "object"
	ValueKindArray   ValueKind = "array"
)

// Value is a lossless dynamic type. Missing field and null are different:
// - missing: key not present in object map
// - null: key exists with Kind == ValueKindNull
type Value struct {
	Kind    ValueKind    `json:"kind"`
	String  string       `json:"string,omitempty"`
	Bool    bool         `json:"bool,omitempty"`
	Int64   int64        `json:"int64,omitempty"`
	UInt64  uint64       `json:"uint64,omitempty"`
	Decimal string       `json:"decimal,omitempty"` // exact decimal string
	Bytes   []byte       `json:"bytes,omitempty"`
	Object  *ObjectValue `json:"object,omitempty"`
	Array   []Value      `json:"array,omitempty"`
}

// ObjectValue keeps fields in dynamic object form.
type ObjectValue struct {
	Fields map[string]Value `json:"fields,omitempty"`
}

func NullValue() Value {
	return Value{Kind: ValueKindNull}
}

func StringValue(v string) Value {
	return Value{Kind: ValueKindString, String: v}
}

func BoolValue(v bool) Value {
	return Value{Kind: ValueKindBool, Bool: v}
}

func Int64Value(v int64) Value {
	return Value{Kind: ValueKindInt64, Int64: v}
}

func UInt64Value(v uint64) Value {
	return Value{Kind: ValueKindUInt64, UInt64: v}
}

func DecimalValue(v string) Value {
	return Value{Kind: ValueKindDecimal, Decimal: strings.TrimSpace(v)}
}

func BytesValue(v []byte) Value {
	copyBytes := make([]byte, len(v))
	copy(copyBytes, v)
	return Value{Kind: ValueKindBytes, Bytes: copyBytes}
}

func ObjectMapValue(fields map[string]Value) Value {
	out := make(map[string]Value, len(fields))
	for k, v := range fields {
		out[k] = v
	}
	return Value{Kind: ValueKindObject, Object: &ObjectValue{Fields: out}}
}

func ArrayValue(items []Value) Value {
	out := make([]Value, len(items))
	copy(out, items)
	return Value{Kind: ValueKindArray, Array: out}
}
