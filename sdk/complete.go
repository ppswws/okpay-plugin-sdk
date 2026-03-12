package sdk

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ppswws/okpay-plugin-sdk/contract"
	"github.com/ppswws/okpay-plugin-sdk/proto"
)

// CompleteBizInput defines plugin -> kernel completion payload.
type CompleteBizInput struct {
	BizType     proto.BizType  `json:"bizType"`
	BizNo       string         `json:"bizNo"`
	State       proto.BizState `json:"state"`
	APIBizNo    string         `json:"apiBizNo,omitempty"`
	ChannelCode string         `json:"channelCode,omitempty"`
	ChannelMsg  string         `json:"channelMsg,omitempty"`
	RespBody    string         `json:"respBody,omitempty"`
	Buyer       string         `json:"buyer,omitempty"`
}

// CompleteBiz sends final or intermediate business state back to kernel.
func CompleteBiz(ctx context.Context, req CompleteBizInput) error {
	kernel, conn, err := contract.DialKernelServiceFromContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if req.BizNo == "" {
		return fmt.Errorf("bizNo is empty")
	}
	ack, err := kernel.CompleteBiz(ctx, &proto.CompleteBizRequest{
		RequestId:   callbackRequestID(req.BizNo),
		BizType:     req.BizType,
		BizNo:       req.BizNo,
		State:       req.State,
		ApiBizNo:    req.APIBizNo,
		ChannelCode: req.ChannelCode,
		ChannelMsg:  req.ChannelMsg,
		RespBody:    req.RespBody,
		Buyer:       req.Buyer,
	})
	if err != nil {
		return err
	}
	if ack == nil || !ack.Accepted {
		return fmt.Errorf("kernel complete biz rejected")
	}
	return nil
}

func callbackRequestID(bizNo string) string {
	return "cb:" + bizNo + ":" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
