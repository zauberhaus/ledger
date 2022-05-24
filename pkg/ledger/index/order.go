package index

var Order = OrderIndex{
	index{
		prefix: "OR",
		max:    3,
	},
}

type OrderIndex struct {
	index
}

func (a *OrderIndex) Key(order string, item string, id string) []byte {
	return []byte(a.scan(order, item, id))
}
