package plugin

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ReturnOrOK builds a return response:
// - no ReturnURL -> ok page
// - ReturnURL exists -> redirect unless return_timeout exceeded
func ReturnOrOK(req *CallRequest) map[string]any {
	if req == nil {
		return map[string]any{"type": "page", "page": "ok"}
	}
	order := DecodeOrder(req.Order)
	if order == nil {
		return map[string]any{"type": "page", "page": "ok"}
	}
	returnURL := strings.TrimSpace(order.ReturnURL)
	if returnURL == "" {
		return map[string]any{"type": "page", "page": "ok"}
	}
	timeoutSec := parseTimeoutSeconds(fmt.Sprint(req.Config["return_timeout"]))
	if timeoutSec > 0 && !order.Endtime.IsZero() {
		if time.Since(order.Endtime) > time.Duration(timeoutSec)*time.Second {
			return map[string]any{"type": "page", "page": "ok"}
		}
	}
	return map[string]any{"type": "return", "url": returnURL}
}

func parseTimeoutSeconds(val string) int64 {
	val = strings.TrimSpace(val)
	if val == "" {
		return 0
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return n
}
