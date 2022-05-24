package index

import "github.com/ec-systems/core.ledger.tool/pkg/types"

// AS:asset:consumer:account

var Asset = AssetIndex{
	index{
		prefix: "AS",
		max:    3,
	},
}

type AssetIndex struct {
	index
}

func (a *AssetIndex) Key(asset types.Asset, customer string, account types.Account) []byte {
	return []byte(a.scan(asset.String(), customer, account.String()))
}
