package index

import "github.com/ec-systems/core.ledger.service/pkg/types"

var Holder = HolderIndex{
	index{
		prefix: "CU",
		max:    3,
	},
}

type HolderIndex struct {
	index
}

func (c *HolderIndex) Key(holder string, asset types.Asset, account types.Account) []byte {
	return []byte(c.scan(holder, asset.String(), account.String()))
}

func (c *HolderIndex) Accounts(holder string, asset types.Asset) string {
	return c.scan(holder, asset.String())
}

func (c *HolderIndex) All() string {
	return c.scan()
}
