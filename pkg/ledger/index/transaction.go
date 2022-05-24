package index

import "github.com/ec-systems/core.ledger.tool/pkg/types"

var Transaction = TransactionIndex{
	index{
		prefix: "TX",
		max:    4,
	},
}

type TransactionIndex struct {
	index
}

func (a *TransactionIndex) Key(customer string, asset types.Asset, account types.Account, id string) []byte {
	return []byte(a.scan(customer, asset.String(), account.String(), id))
}

func (a *TransactionIndex) Scan(customer string, asset types.Asset, account types.Account) string {
	path := a.strip(customer, asset.String(), account.String())
	return a.scan(path...)
}
