package ledger

import (
	"bytes"
	"encoding/gob"
)

type GOBSerializer struct {
}

func (GOBSerializer) Marshal(v interface{}, version uint16) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (GOBSerializer) Unmarshal(data []byte, v interface{}, version uint16) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	return decoder.Decode(v)
}
