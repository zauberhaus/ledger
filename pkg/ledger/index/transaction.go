package index

import "github.com/ec-systems/core.ledger.service/pkg/types"

var Transaction = TransactionIndex{
	index{
		prefix: "TX",
		max:    4,
	},
}

type TransactionIndex struct {
	index
}

func (a *TransactionIndex) Key(holder string, asset types.Asset, account types.Account, id types.ID) []byte {
	return []byte(a.scan(holder, asset.String(), account.String(), id.HexString()))
}

func (a *TransactionIndex) Scan(holder string, asset types.Asset, account types.Account) string {
	path := a.strip(holder, asset.String(), account.String())
	return a.scan(path...)
}
