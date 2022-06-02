//go:generate go run github.com/ec-systems/core.ledger.service/pkg/generator/assets/

package types

import (
	"fmt"
	"sort"
)

const (
	AllAssets Asset = ""
	BNB       Asset = "BNB"
	XRP       Asset = "XRP"
	Cardano   Asset = "ADA"
	Bitcoin   Asset = "BTC"
	Ethereum  Asset = "ETH"
	Tether    Asset = "USDT"
	USDCoin   Asset = "USDC"
)

type Assets map[Asset]string

func (a Assets) Parse(txt string) (Asset, error) {
	if txt == "" {
		return AllAssets, nil
	}

	asset := Asset(txt)
	if _, ok := a[asset]; ok {
		return asset, nil
	}

	return "", fmt.Errorf("unsupported asset: %v", txt)
}

/*
func (a Assets) Map() map[string]string {
	m := map[string]string{}

	for k, v := range a {
		m[k] = v.String()
	}

	return m
}
*/

func (a Assets) MarshalYAML() (interface{}, error) {
	max := 20

	if len(a) < max {
		max = len(a)
	}

	keys := []string{}
	for k := range a {
		keys = append(keys, k.String())
	}

	sort.Strings(keys)

	list := []string{}
	for i := 0; i < max; i++ {
		list = append(list, keys[i])
	}

	if len(a) > max {
		list = append(list, "...")
	}

	return list, nil
}

type Asset string

func (a Assets) Name(asset Asset) string {
	name, ok := a[asset]
	if ok {
		return name
	}

	return "Unknown"
}

func (c Asset) String() string {
	return string(c)
}

func (c Asset) Check(a Assets) bool {
	_, ok := a[c]
	return ok
}
