package ledger

import (
	"encoding/binary"
	"fmt"

	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/ec-systems/core.ledger.service/pkg/types"
)

type DatabaseObject interface {
	SetTX(uint64)
	SetKey(string)
}

type Serializer interface {
	Marshal(interface{}, uint16) ([]byte, error)
	Unmarshal([]byte, interface{}, uint16) error
}

var Serializers = map[types.Format]Serializer{
	types.JSON:     &JSONSerializer{},
	types.Protobuf: &ProtobufSerializer{},
	types.GOB:      &GOBSerializer{},
}

func Unmarshal(e *schema.Entry, v interface{}) error {
	data := e.Value

	version := binary.LittleEndian.Uint16(data[0:2])
	format := types.Format(binary.LittleEndian.Uint16(data[2:4]))

	data = data[4:]

	serializer, ok := Serializers[format]
	if !ok {
		return fmt.Errorf("unknown serializer %v", format)
	}

	err := serializer.Unmarshal(data, v, version)
	if err != nil {
		return fmt.Errorf("unmarshal error: %v (%v)", err, format)
	}

	tx, ok := v.(DatabaseObject)
	if ok {
		tx.SetTX(e.Tx)
		tx.SetKey(string(e.Key))
	}

	return nil
}

func Marshal(v interface{}, format types.Format, version uint16) ([]byte, error) {
	serializer, ok := Serializers[format]
	if !ok {
		return nil, fmt.Errorf("unknown serializer %v", format)
	}

	data, err := serializer.Marshal(v, version)
	if err != nil {
		return nil, err
	}

	h1 := make([]byte, 2)
	binary.LittleEndian.PutUint16(h1, version)

	h2 := make([]byte, 2)
	binary.LittleEndian.PutUint16(h2, uint16(format))

	header := append(h1, h2...)

	return append(header, data...), nil
}
