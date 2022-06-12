package index

import "github.com/ec-systems/core.ledger.server/pkg/types"

var Account = AccountIndex{
	index{
		prefix: "AC",
		max:    1,
	},
}

type AccountIndex struct {
	index
}

func (c *AccountIndex) Key(account types.Account) []byte {
	return []byte(c.scan(account.String()))
}
