package plugin

import (
	"encoding/json"
	"time"
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
	Name       string    `json:"name"`
	Money      int64     `json:"money"`
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
	Err        string    `json:"err"`
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
	Money       int64     `json:"money"`
	ReqMs       int32     `json:"req_ms"`
	ReqBody     string    `json:"req_body"`
	RespBody    string    `json:"resp_body"`
	Status      int16     `json:"status"`
	Remark      string    `json:"remark"`
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
	Channel    int64     `json:"channel"`
	Money      int64     `json:"money"`
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
	Rate      float64         `json:"rate"`
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

// DecodeOrder 将 map 转为 OrderPayload（失败返回 nil）。
func DecodeOrder(raw map[string]any) *OrderPayload {
	return decodeTo[OrderPayload](raw)
}

// DecodeRefund 将 map 转为 RefundPayload（失败返回 nil）。
func DecodeRefund(raw map[string]any) *RefundPayload {
	return decodeTo[RefundPayload](raw)
}

// DecodeTransfer 将 map 转为 TransferPayload（失败返回 nil）。
func DecodeTransfer(raw map[string]any) *TransferPayload {
	return decodeTo[TransferPayload](raw)
}

// DecodeChannel 将 map 转为 ChannelPayload（失败返回 nil）。
func DecodeChannel(raw map[string]any) *ChannelPayload {
	return decodeTo[ChannelPayload](raw)
}

func decodeTo[T any](raw map[string]any) *T {
	if raw == nil {
		return nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	return &out
}
