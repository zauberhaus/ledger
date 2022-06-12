package index

import "github.com/ec-systems/core.ledger.server/pkg/types"

var Reference = ReferenceIndex{
	index{
		prefix: "RF",
		max:    2,
	},
}

type ReferenceIndex struct {
	index
}

func (i *ReferenceIndex) Key(src types.ID, dest types.ID) []byte {
	return []byte(i.scan(src.HexString(), dest.HexString()))
}

func (i *ReferenceIndex) Source(src types.ID) string {
	return i.scan(src.HexString())
}
