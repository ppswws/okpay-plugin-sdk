package sdk

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ppswws/okpay-plugin-sdk/contract"
	"github.com/ppswws/okpay-plugin-sdk/proto"
)

// BizDoneIn defines plugin -> kernel completion payload.
type BizDoneIn struct {
	BizType  proto.BizType  `json:"bizType"`
	BizNo    string         `json:"bizNo"`
	State    proto.BizState `json:"state"`
	ApiNo    string         `json:"apiNo,omitempty"`
	Code     string         `json:"code,omitempty"`
	Msg      string         `json:"msg,omitempty"`
	RespBody string         `json:"respBody,omitempty"`
	Buyer    string         `json:"buyer,omitempty"`
}

// CompleteBiz sends final or intermediate business state back to kernel.
func CompleteBiz(ctx context.Context, req BizDoneIn) error {
	kernel, conn, err := contract.DialKernelServiceFromContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if req.BizNo == "" {
		return fmt.Errorf("bizNo is empty")
	}
	ack, err := kernel.CompleteBiz(ctx, &proto.BizDoneReq{
		RequestId: cbReqID(req.BizNo),
		BizType:   req.BizType,
		BizNo:     req.BizNo,
		State:     req.State,
		ApiNo:     req.ApiNo,
		Code:      req.Code,
		Msg:       req.Msg,
		RespBody:  req.RespBody,
		Buyer:     req.Buyer,
	})
	if err != nil {
		return err
	}
	if ack == nil || !ack.Accepted {
		return fmt.Errorf("kernel complete biz rejected")
	}
	return nil
}

func cbReqID(bizNo string) string {
	return "cb:" + bizNo + ":" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
