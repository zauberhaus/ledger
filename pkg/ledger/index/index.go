package index

import (
	"strings"
)

type index struct {
	prefix string
	max    int
}

func (i *index) scan(parts ...string) string {
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}

	if len(parts) > i.max {
		parts = parts[0 : i.max-1]
	}

	index := i.prefix + ":" + strings.Join(parts, ":")
	if len(parts) > 0 && len(parts) < i.max {
		return index + ":"
	}

	return index

}

func (i *index) strip(parts ...string) []string {
	result := []string{}

	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if len(part) > 0 {
			result = append([]string{part}, result...)
		}
	}

	return result
}
