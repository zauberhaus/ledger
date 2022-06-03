package ledger

// protoc --proto_path=proto --go_out=. ./proto/transaction.proto

import (
	"fmt"
	"reflect"

	"github.com/ec-systems/core.ledger.service/pkg/ledger/protobuf"
	"github.com/ec-systems/core.ledger.service/pkg/types"

	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"
)

type ProtobufSerializer struct {
}

func (ProtobufSerializer) Marshal(v interface{}, version uint16) ([]byte, error) {
	switch o := v.(type) {
	case *Transaction:
		created, err := o.Created.MarshalBinary()
		if err != nil {
			return nil, err
		}

		modified, err := o.Modified.MarshalBinary()
		if err != nil {
			return nil, err
		}

		tx := &protobuf.Transaction{
			ID:        o.ID.Bytes(),
			Account:   o.Account.String(),
			Holder:    o.Holder,
			Order:     o.Order,
			Item:      o.Item,
			Asset:     o.Asset.String(),
			Amount:    o.Amount.String(),
			Status:    int64(o.Status),
			Created:   created,
			Modified:  modified,
			User:      o.User,
			Reference: o.Reference,
		}

		return proto.Marshal(tx)

	default:
		return nil, fmt.Errorf("protobuf marshal: unsupported type: %v", reflect.TypeOf(v))

	}
}

func (ProtobufSerializer) Unmarshal(data []byte, v interface{}, version uint16) error {
	switch o := v.(type) {
	case *Transaction:
		tx := &protobuf.Transaction{}
		err := proto.Unmarshal(data, tx)
		if err != nil {
			return err
		}

		err = o.Created.UnmarshalBinary(tx.Created)
		if err != nil {
			return err
		}

		err = o.Modified.UnmarshalBinary(tx.Modified)
		if err != nil {
			return err
		}

		o.ID = types.NewID(tx.ID)
		o.Account = types.Account(tx.Account)
		o.Holder = tx.Holder
		o.Order = tx.Order
		o.Item = tx.Item
		o.Asset = types.Asset(tx.Asset)
		o.Amount, _ = decimal.NewFromString(tx.Amount)

		o.Status = types.Status(tx.Status)
		o.User = tx.User
		o.Reference = tx.Reference

	default:
		return fmt.Errorf("protobuf unmarshal: unsupported type: %v", reflect.TypeOf(v))
	}

	return nil
}
