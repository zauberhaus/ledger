package index

import "github.com/ec-systems/core.ledger.service/pkg/types"

var AssetTx = AssetTxIndex{
	index{
		prefix: "AT",
		max:    4,
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

type AssetTxIndex struct {
	index
}

func (a *AssetTxIndex) Key(asset types.Asset, holder string, account types.Account, id types.ID) []byte {
	return []byte(a.scan(asset.String(), holder, account.String(), id.HexString()))
}

func (a *AssetTxIndex) Asset(asset types.Asset) string {
	return a.scan(asset.String())
}
