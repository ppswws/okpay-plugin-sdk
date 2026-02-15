package plugin

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// SignPayload 生成请求签名（MD5）。
func SignPayload(payload map[string]any, secret string) string {
	if payload == nil {
		payload = map[string]any{}
	}
	secret = strings.TrimSpace(secret)
	pairs := make([]string, 0, len(payload))
	for k, v := range payload {
		if strings.EqualFold(strings.TrimSpace(k), "sign") {
			continue
		}
		if isEmptyValue(v) {
			continue
		}
		valStr, ok := valueToString(v)
		if !ok || valStr == "" {
			continue
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, valStr))
	}
	sort.Strings(pairs)
	raw := strings.Join(pairs, "&")
	if secret != "" {
		if raw != "" {
			raw += "&"
		}
		raw += "key=" + secret
	}
	sum := md5.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func isEmptyValue(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val) == ""
	case []any:
		return len(val) == 0
	case map[string]any:
		return len(val) == 0
	case json.RawMessage:
		return len(bytes.TrimSpace(val)) == 0
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() == 0
	}
	return false
}

func valueToString(v any) (string, bool) {
	switch val := v.(type) {
	case string:
		return val, true
	case json.Number:
		return val.String(), true
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), true
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32), true
	case int:
		return strconv.Itoa(val), true
	case int64:
		return strconv.FormatInt(val, 10), true
	case int32:
		return strconv.FormatInt(int64(val), 10), true
	case int16:
		return strconv.FormatInt(int64(val), 10), true
	case int8:
		return strconv.FormatInt(int64(val), 10), true
	case uint:
		return strconv.FormatUint(uint64(val), 10), true
	case uint64:
		return strconv.FormatUint(val, 10), true
	case uint32:
		return strconv.FormatUint(uint64(val), 10), true
	case uint16:
		return strconv.FormatUint(uint64(val), 10), true
	case uint8:
		return strconv.FormatUint(uint64(val), 10), true
	case bool:
		if val {
			return "true", true
		}
		return "false", true
	default:
		data, err := json.Marshal(val)
		if err != nil {
			return "", false
		}
		return string(data), true
	}
}
