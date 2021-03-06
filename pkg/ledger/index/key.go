package index

import "github.com/ec-systems/core.ledger.server/pkg/types"

var Key = KeyIndex{
	index{
		prefix: "ID",
		max:    1,
	},
}

type KeyIndex struct {
	index
}

func (t *KeyIndex) Key(id types.ID) []byte {
	return []byte(t.ID(id))
}

func (t *KeyIndex) ID(id types.ID) string {
	return t.scan(id.HexString())
}
