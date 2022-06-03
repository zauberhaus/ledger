package ledger

import (
	"time"

	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/ec-systems/core.ledger.service/pkg/ledger/index"
	"github.com/ec-systems/core.ledger.service/pkg/types"
)

func (l *Ledger) CancelOperations(tx *Transaction) ([]interface{}, *Transaction, error) {
	id, err := l.NewID()
	if err != nil {
		return nil, nil, err
	}

	ref, err := types.NewReference(
		struct {
			ID     string
			Status types.Status
		}{
			tx.ID.String(),
			types.Canceled,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	cancel := tx.Copy()
	cancel.ID = id
	cancel.Status = types.Finished
	cancel.Amount = tx.Amount.Neg()
	cancel.Reference = ref.String()

	ops := []interface{}{}

	op1, key, err := l.CreateOperations(cancel)
	if err != nil {
		return nil, nil, err
	}

	ops = append(ops, op1...)
	cancel.key = key

	tx.Status = types.Canceled

	op2, _, err := l.UpdateOperations(tx)
	if err != nil {
		return nil, nil, err
	}

	ops = append(ops, op2...)

	ops = append(ops, l.RefOperation(tx, cancel))

	return ops, cancel, nil
}

func (l *Ledger) RefOperation(src *Transaction, dest *Transaction) *schema.Op_Ref {
	return &schema.Op_Ref{
		Ref: &schema.ReferenceRequest{
			ReferencedKey: []byte(dest.key),
			Key:           index.Reference.Key(src.ID, dest.ID),
			BoundRef:      false,
		},
	}
}

func (l *Ledger) UpdateOperations(tx *Transaction) ([]interface{}, string, error) {
	tx.Modified = time.Now()

	if tx.Created.IsZero() {
		tx.Created = tx.Modified
	}

	if !tx.Account.Check() {
		return nil, "", NewError(BadRequestError, "checksum check failed for '%v'", tx.Account)
	}

	if !tx.Asset.Check(l.assets) {
		return nil, "", NewError(BadRequestError, "invalid asset '%v'", tx.Asset)
	}

	if tx.Holder == "" {
		return nil, "", NewError(BadRequestError, "holder is empty")
	}

	if tx.ID.IsEmpty() {
		return nil, "", NewError(BadRequestError, "holder '%v' transaction id is empty", tx.Holder)
	}

	if tx.Amount.IsZero() {
		return nil, "", NewError(BadRequestError, "amount is zero: %v", tx.ID)
	}

	data, err := tx.Bytes(l.format)
	if err != nil {
		return nil, "", NewError(InternalError, "marshal transaction failed: %v", err)
	}

	kv := &schema.Op_Kv{
		Kv: &schema.KeyValue{
			Key:   index.Key.Key(tx.ID),
			Value: data,
		},
	}

	return []interface{}{
		kv,
	}, string(kv.Kv.Key), nil
}

func (l *Ledger) CreateOperations(tx *Transaction) ([]interface{}, string, error) {
	tx.Modified = time.Now()

	if tx.Created.IsZero() {
		tx.Created = tx.Modified
	}

	if !tx.Account.Check() {
		return nil, "", NewError(BadRequestError, "checksum check failed for '%v'", tx.Account)
	}

	if !tx.Asset.Check(l.assets) {
		return nil, "", NewError(BadRequestError, "invalid asset '%v'", tx.Asset)
	}

	if tx.Holder == "" {
		return nil, "", NewError(BadRequestError, "holder is empty")
	}

	if tx.ID.IsEmpty() {
		return nil, "", NewError(BadRequestError, "holder '%v' transaction id is empty", tx.Holder)
	}

	if tx.Amount.IsZero() {
		return nil, "", nil
	}

	data, err := tx.Bytes(l.format)
	if err != nil {
		return nil, "", NewError(InternalError, "marshal transaction failed: %v", err)
	}

	kv := &schema.Op_Kv{
		Kv: &schema.KeyValue{
			Key:   index.Key.Key(tx.ID),
			Value: data,
		},
	}

	var order *schema.Op_Ref
	var item *schema.Op_Ref
	if tx.Order != "" || tx.Item != "" {
		order = &schema.Op_Ref{
			Ref: &schema.ReferenceRequest{
				ReferencedKey: kv.Kv.Key,
				Key:           index.Order.Key(tx.Holder, tx.Order),
				BoundRef:      false,
			},
		}

		item = &schema.Op_Ref{
			Ref: &schema.ReferenceRequest{
				ReferencedKey: kv.Kv.Key,
				Key:           index.OrderItem.Key(tx.Holder, tx.Order, tx.Item, tx.ID),
				BoundRef:      false,
			},
		}
	}

	transaction := &schema.Op_Ref{
		Ref: &schema.ReferenceRequest{
			ReferencedKey: kv.Kv.Key,
			Key:           index.Transaction.Key(tx.Holder, tx.Asset, tx.Account, tx.ID),
			BoundRef:      false,
		},
	}

	holder := &schema.Op_Ref{
		Ref: &schema.ReferenceRequest{
			ReferencedKey: kv.Kv.Key,
			Key:           index.Holder.Key(tx.Holder, tx.Asset, tx.Account),
			BoundRef:      false,
		},
	}

	account := &schema.Op_Ref{
		Ref: &schema.ReferenceRequest{
			ReferencedKey: kv.Kv.Key,
			Key:           index.Account.Key(tx.Account),
			BoundRef:      false,
		},
	}

	assetTx := &schema.Op_Ref{
		Ref: &schema.ReferenceRequest{
			ReferencedKey: kv.Kv.Key,
			Key:           index.AssetTx.Key(tx.Asset, tx.Holder, tx.Account, tx.ID),
			BoundRef:      false,
		},
	}

	asset := &schema.Op_Ref{
		Ref: &schema.ReferenceRequest{
			ReferencedKey: kv.Kv.Key,
			Key:           index.Asset.Key(tx.Asset),
			BoundRef:      false,
		},
	}

	return []interface{}{
		kv,
		order,
		item,
		transaction,
		holder,
		account,
		asset,
		assetTx,
	}, string(kv.Kv.Key), nil
}
