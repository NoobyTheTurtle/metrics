package adapter

import (
	"fmt"
	"strings"
)

type Prefix string

const (
	GaugePrefix   Prefix = "gauge:"
	CounterPrefix Prefix = "counter:"
)

func addPrefix(name string, prefix Prefix) string {
	return fmt.Sprintf("%s%s", prefix, name)
}

func trimPrefix(key string, prefix Prefix) string {
	return strings.TrimPrefix(key, string(prefix))
}

func hasPrefix(key string, prefix Prefix) bool {
	return strings.HasPrefix(key, string(prefix))
}
