package ledger

import "encoding/json"

type JSONSerializer struct {
}

func (JSONSerializer) Marshal(v interface{}, version uint16) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONSerializer) Unmarshal(data []byte, v interface{}, version uint16) error {
	return json.Unmarshal(data, v)
}
