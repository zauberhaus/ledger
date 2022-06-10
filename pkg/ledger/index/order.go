package index

var Order = OrderIndex{
	index{
		prefix: "OR",
		max:    2,
	},
}

var OrderItem = OrderItemIndex{
	index{
		prefix: "OI",
		max:    1,
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

func (a *OrderItemIndex) Key(order string) []byte {
	return []byte(a.scan(order))
}

func (a *OrderItemIndex) Scan(order string) string {
	return a.scan(a.strip(order)...)
}
