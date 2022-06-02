package types

import (
	"encoding/hex"

	"github.com/google/uuid"
)

var (
	ZeroID = ID{}
)

type ID struct {
	uuid.UUID
}

func NewID(data []byte) ID {
	if len(data) != 16 {
		return ID{}
	}

	return ID{*(*[16]byte)(data)}
}

func NewRandomID() (ID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return ID{}, nil
	}

	return ID{id}, nil
}

func (id ID) String() string {
	return id.UUID.String()
}

func (id ID) HexString() string {
	return hex.EncodeToString(id.UUID[:])
}

func (id ID) Bytes() []byte {
	return id.UUID[:]
}

func (id ID) IsEmpty() bool {
	return id == ID{}
}

func (id ID) MarshalText() (text []byte, err error) {
	hex := hex.EncodeToString(id.UUID[:])
	return []byte(hex), nil
}

func (id *ID) UnmarshalText(text []byte) error {
	if data, err := hex.DecodeString(string(text)); err == nil || len(data) != 16 {
		*id = NewID(data)
	} else {
		tmp, err := uuid.Parse(string(text))
		if err != nil {
			return err
		}

		*id = ID{tmp}
	}

	return nil
}
