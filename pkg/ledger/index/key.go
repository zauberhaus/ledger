package index

var Key = KeyIndex{
	index{
		prefix: "ID",
		max:    1,
	},
}

type KeyIndex struct {
	index
}

func (t *KeyIndex) Key(id string) []byte {
	return []byte(t.scan(id))
}

func (t *KeyIndex) ID(id string) string {
	return t.scan(id)
}
