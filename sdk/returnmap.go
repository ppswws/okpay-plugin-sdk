package sdk

import (
	"encoding/json"
	"fmt"

	"github.com/ppswws/okpay-plugin-sdk/proto"
)

// BuildReturnMap serializes PageResponse to lock-order ext payload.
func BuildReturnMap(page *proto.PageResponse) map[string]any {
	if page == nil {
		return map[string]any{"type": TypeError, "msg": "empty page response"}
	}
	out := map[string]any{"type": page.GetType()}
	if page.GetPage() != "" {
		out["page"] = page.GetPage()
	}
	if page.GetUrl() != "" {
		out["url"] = page.GetUrl()
	}
	if page.GetMsg() != "" {
		out["msg"] = page.GetMsg()
	}
	if len(page.GetDataRaw()) > 0 {
		var data any
		if err := json.Unmarshal(page.GetDataRaw(), &data); err == nil {
			out["data"] = data
		}
	}
	if page.GetDataText() != "" {
		out["data"] = page.GetDataText()
	}
	return out
}

// BuildReturnPage parses lock-order ext payload back to PageResponse.
func BuildReturnPage(payload map[string]any) *proto.PageResponse {
	errMsg := ""
	resp := &proto.PageResponse{}
	if payload == nil {
		errMsg = "empty page payload"
	} else {
		resp = &proto.PageResponse{
			Type: mapString(payload, "type"),
			Page: mapString(payload, "page"),
			Url:  mapString(payload, "url"),
			Msg:  mapString(payload, "msg"),
		}
		if data, ok := payload["data"]; ok && data != nil {
			switch resp.GetType() {
			case TypeHTML:
				resp.DataText = fmt.Sprint(data)
			default:
				raw, _ := json.Marshal(data)
				resp.DataRaw = raw
			}
		}
		if resp.GetType() == "" {
			errMsg = "invalid page payload"
		}
	}
	if errMsg != "" {
		resp = RespError(errMsg)
	}
	return resp
}

func mapString(m map[string]any, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}
