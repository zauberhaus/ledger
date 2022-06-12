package index

import (
	"github.com/ec-systems/core.ledger.server/pkg/types"
)

var Transaction = TransactionIndex{
	index{
		prefix: "TX",
		max:    1,
	},
}

type TransactionIndex struct {
	index
}

func (a *TransactionIndex) Key(account types.Account) []byte {
	return []byte(a.scan(account.String()))
}

func (a *TransactionIndex) Scan(account types.Account) string {
	path := a.strip(account.String())
	return a.scan(path...)
}
