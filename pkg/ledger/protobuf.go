package ledger

// protoc --proto_path=proto --go_out=. ./proto/transaction.proto

import (
	"fmt"
	"reflect"
	"time"

	"github.com/ec-systems/core.ledger.server/pkg/ledger/protobuf"
	"github.com/ec-systems/core.ledger.server/pkg/types"

	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"
)

type ProtobufSerializer struct {
}

func (ProtobufSerializer) Marshal(v interface{}, version uint16) ([]byte, error) {
	switch o := v.(type) {
	case *Transaction:
		tx := &protobuf.Transaction{
			ID:        o.ID.Bytes(),
			Account:   o.Account.String(),
			Holder:    o.Holder,
			Order:     o.Order,
			Item:      o.Item,
			Asset:     o.Asset.String(),
			Amount:    o.Amount.String(),
			Status:    int64(o.Status),
			User:      o.User,
			Reference: o.Reference,
		}

		if o.Created != nil {
			created, err := o.Created.MarshalBinary()
			if err != nil {
				return nil, err
			}

			tx.Created = created
		}

		if o.Modified != nil {
			modified, err := o.Modified.MarshalBinary()
			if err != nil {
				return nil, err
			}

			tx.Modified = modified
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

		if tx.Created != nil {
			var tmp time.Time
			err = tmp.UnmarshalBinary(tx.Created)
			if err != nil {
				return err
			}

			o.Created = &tmp
		}

		if tx.Modified != nil {
			var tmp time.Time
			err = tmp.UnmarshalBinary(tx.Modified)
			if err != nil {
				return err
			}

			o.Modified = &tmp
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
