package index

import "github.com/ec-systems/core.ledger.service/pkg/types"

var AssetTx = AssetIndex{
	index{
		prefix: "AT",
		max:    1,
	},
}

var Asset = AssetIndex{
	index{
		prefix: "AS",
		max:    1,
	},
}

type AssetIndex struct {
	index
}

func (a *AssetIndex) Key(asset types.Asset) []byte {
	return []byte(a.scan(asset.String()))
}

func (a *AssetIndex) Assets() string {
	return a.scan()
}
