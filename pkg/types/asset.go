//go:generate go run github.com/ec-systems/core.ledger.tool/pkg/generator/assets/

package types

import "fmt"

const (
	AllAssets Asset = ""
)

type Assets map[string]Asset

func (a Assets) Parse(txt string) (Asset, error) {
	asset, ok := a[txt]
	if !ok {
		return "", fmt.Errorf("unsupported asset: %v", txt)
	}

	return asset, nil
}

func (a Assets) Map() map[string]string {
	m := map[string]string{}

	for k, v := range a {
		m[k] = v.String()
	}

	return m
}

type Asset string

func (c Asset) String() string {
	return string(c)
}

func (c Asset) Check(a Assets) bool {
	if c == "" {
		return false
	}

	for _, v := range a {
		if v == c {
			return true
		}
	}

	return false
}
