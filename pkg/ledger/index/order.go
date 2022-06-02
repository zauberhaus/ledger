package index

import "github.com/ec-systems/core.ledger.service/pkg/types"

var Order = OrderIndex{
	index{
		prefix: "OR",
		max:    2,
	},
}

var OrderItem = OrderItemIndex{
	index{
		prefix: "OI",
		max:    4,
	},
}

type OrderIndex struct {
	index
}

func (a *OrderIndex) Key(holder string, order string) []byte {
	return []byte(a.scan(holder, order))
}

func (a *OrderIndex) Orders(holder string) string {
	return a.scan(holder)
}

type OrderItemIndex struct {
	index
}

func (a *OrderItemIndex) Key(holder string, order string, item string, id types.ID) []byte {
	return []byte(a.scan(holder, order, item, id.HexString()))
}

func (a *OrderItemIndex) Scan(holder string, order string, item string) string {
	return a.scan(a.strip(holder, order, item)...)
}
