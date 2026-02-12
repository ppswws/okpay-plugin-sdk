package plugin

import "strings"

const paymentConfigKey = "sys.payment"

func getConfigString(conf map[string]any, key string) string {
	if conf == nil {
		return ""
	}
	if val, ok := conf[key]; ok {
		if s, ok := val.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}
