package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"okpay/payment/plugin/contract"
)

// OrderPayload 提供插件侧使用的订单结构。
type OrderPayload struct {
	TradeNo    string    `json:"trade_no"`
	OutTradeNo string    `json:"out_trade_no"`
	APITradeNo string    `json:"api_trade_no"`
	UID        int64     `json:"uid"`
	Type       string    `json:"type"`
	Plugin     string    `json:"plugin"`
	Channel    int64     `json:"channel"`
	CID        int64     `json:"cid"`
	Code       int64     `json:"code"`
	Subject    string    `json:"subject"`
	Amount     int64     `json:"amount"`
	Real       int64     `json:"real"`
	Fee        int64     `json:"fee"`
	Get        int64     `json:"get"`
	NotifyURL  string    `json:"notify_url"`
	ReturnURL  string    `json:"return_url"`
	Param      string    `json:"param"`
	Domain     string    `json:"domain"`
	IPBuyer    string    `json:"ip_buyer"`
	IPSource   string    `json:"ip_source"`
	Buyer      string    `json:"buyer"`
	Status     int16     `json:"status"`
	Notify     int16     `json:"notify"`
	ReqMs      int32     `json:"req_ms"`
	ReqCount   int16     `json:"req_count"`
	ReqBody    string    `json:"req_body"`
	RespBody   string    `json:"resp_body"`
	Ext        string    `json:"ext"`
	Result     string    `json:"result"`
	Endtime    time.Time `json:"endtime"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// RefundPayload 提供插件侧使用的退款结构。
type RefundPayload struct {
	RefundNo    string    `json:"refund_no"`
	TradeNo     string    `json:"trade_no"`
	OutRefundNo string    `json:"out_refund_no"`
	UID         int64     `json:"uid"`
	Channel     int64     `json:"channel"`
	Amount      int64     `json:"amount"`
	ReqMs       int32     `json:"req_ms"`
	ReqBody     string    `json:"req_body"`
	RespBody    string    `json:"resp_body"`
	Status      int16     `json:"status"`
	Remark      string    `json:"remark"`
	Result      string    `json:"result"`
	Endtime     time.Time `json:"endtime"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TransferPayload 提供插件侧使用的转账结构。
type TransferPayload struct {
	TradeNo    string    `json:"trade_no"`
	APITradeNo string    `json:"api_trade_no"`
	OutTradeNo string    `json:"out_trade_no"`
	UID        int64     `json:"uid"`
	Business   int16     `json:"business"`
	Channel    int64     `json:"channel"`
	Amount     int64     `json:"amount"`
	Fee        int64     `json:"fee"`
	BankName   string    `json:"bank_name"`
	CardName   string    `json:"card_name"`
	CardNo     string    `json:"card_no"`
	BranchName string    `json:"branch_name"`
	NotifyURL  string    `json:"notify_url"`
	ReqMs      int32     `json:"req_ms"`
	ReqBody    string    `json:"req_body"`
	RespBody   string    `json:"resp_body"`
	Status     int16     `json:"status"`
	Remark     string    `json:"remark"`
	Result     string    `json:"result"`
	Endtime    time.Time `json:"endtime"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ChannelPayload 提供插件侧使用的通道结构。
type ChannelPayload struct {
	ID        int64           `json:"id"`
	Type      string          `json:"type"`
	Plugin    string          `json:"plugin"`
	Name      string          `json:"name"`
	Rate      json.Number     `json:"rate"`
	Status    int16           `json:"status"`
	Config    json.RawMessage `json:"config"`
	Daylimit  int64           `json:"daylimit"`
	Daynumber int64           `json:"daynumber"`
	Paymin    int64           `json:"paymin"`
	Paymax    int64           `json:"paymax"`
	Author    int64           `json:"author"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Order 从 InvokeRequestV2 的 parsed.order 解析订单。
func Order(req *contract.InvokeRequestV2) *OrderPayload {
	return decodeFromRequest[OrderPayload](req, "order")
}

// Refund 从 InvokeRequestV2 的 parsed.refund 解析退款。
func Refund(req *contract.InvokeRequestV2) *RefundPayload {
	return decodeFromRequest[RefundPayload](req, "refund")
}

// Transfer 从 InvokeRequestV2 的 parsed.transfer 解析代付。
func Transfer(req *contract.InvokeRequestV2) *TransferPayload {
	return decodeFromRequest[TransferPayload](req, "transfer")
}

// Channel 从 InvokeRequestV2 的 parsed.channel 解析通道。
func Channel(req *contract.InvokeRequestV2) *ChannelPayload {
	return decodeFromRequest[ChannelPayload](req, "channel")
}

func decodeJSONBytesTo[T any](data []byte) (*T, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("empty json")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var out T
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}
	// Reject trailing tokens to avoid partial/ambiguous payloads.
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("unexpected trailing json data")
		}
		return nil, fmt.Errorf("decode trailing failed: %w", err)
	}
	return &out, nil
}

func decodeFromRequest[T any](req *contract.InvokeRequestV2, key string) *T {
	if req == nil {
		return nil
	}
	v, ok := req.Parsed.Data.Fields[key]
	if !ok {
		return nil
	}
	anyVal, err := ValueToAny(v)
	if err != nil {
		return nil
	}
	raw, err := json.Marshal(anyVal)
	if err != nil {
		return nil
	}
	out, err := decodeJSONBytesTo[T](raw)
	if err != nil {
		return nil
	}
	return out
}

// ChannelConfig parses parsed.channel.config into a map.
func ChannelConfig(req *contract.InvokeRequestV2) map[string]any {
	channelCfg := readMapFromPath(req, "channel.config")
	if len(channelCfg) == 0 {
		return map[string]any{}
	}
	return channelCfg
}

// GlobalConfig parses parsed.config (global config) into map.
func GlobalConfig(req *contract.InvokeRequestV2) map[string]any {
	cfg := readMapFromPath(req, "config")
	if len(cfg) == 0 {
		return map[string]any{}
	}
	return cfg
}

func readMapFromPath(req *contract.InvokeRequestV2, path string) map[string]any {
	v, ok := Read(req, path)
	if !ok {
		return map[string]any{}
	}
	anyVal, err := ValueToAny(v)
	if err != nil {
		return map[string]any{}
	}
	m, ok := anyVal.(map[string]any)
	if !ok || m == nil {
		return map[string]any{}
	}
	return m
}
