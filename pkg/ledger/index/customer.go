package index

import "github.com/ec-systems/core.ledger.tool/pkg/types"

var Customer = CustomerIndex{
	index{
		prefix: "CU",
		max:    3,
	},
}

type CustomerIndex struct {
	index
}

func (c *CustomerIndex) Key(customer string, asset types.Asset, account types.Account) []byte {
	return []byte(c.scan(customer, asset.String(), account.String()))
}

func (c *CustomerIndex) Accounts(customer string, asset types.Asset) string {
	return c.scan(customer, asset.String())
}

func (c *CustomerIndex) All() string {
	return c.scan()
}
