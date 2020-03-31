package tools

import (
	"os"
	"strings"
	"time"
)

func GetDateNowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func GetOsEnv(keys []string) map[string]string {
	values := make(map[string]string)
	for _, key := range keys {
		value := os.Getenv(key)
		values[key] = value
	}

	return values
}

func ParseMeta(meta string) map[string]string {
	values := make(map[string]string)

	ss := strings.Split(meta, "&")
	for _, s := range ss {
		v := strings.Split(s, "=")
		if len(v) != 2 {
			continue
		}
		values[v[0]] = v[1]
	}

	return values
}
