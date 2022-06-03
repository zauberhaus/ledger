package types

import (
	"encoding/json"
)

type Reference string

func NewReference(i interface{}) (Reference, error) {
	data, err := json.Marshal(i)
	if err != nil {
		return "", err
	}

	return Reference(string(data)), nil
}

func (r Reference) String() string {
	return string(r)
}

func (r Reference) Content() interface{} {
	var content interface{}
	err := json.Unmarshal([]byte(r), &content)
	if err != nil {
		return err
	}

	return content
}
