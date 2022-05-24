package main

import (
	"fmt"
	"sort"
	"strings"
)

func GroupBindings(env map[string][]string) ([]map[string]string, error) {
	groups := make(map[string]map[string]string)
	group := ""
	var items map[string]string
	for k, l := range env {
		if len(l) > 1 {
			return nil, fmt.Errorf("More than one env binding for %v", k)
		}

		parts := strings.Split(k, ".")
		name := fmt.Sprintf("%v.", len(parts)) + parts[0]
		if len(parts) > 2 {
			name += "." + parts[1]
		}

		if name != group {
			group = name
			if i, ok := groups[group]; ok {
				items = i
			} else {
				items = make(map[string]string)
				groups[name] = items
			}
		}

		items[k] = l[0]
	}

	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := []map[string]string{}
	for _, k := range keys {
		result = append(result, groups[k])
	}

	return result, nil
}
