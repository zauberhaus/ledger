package ledger

import (
	"fmt"
	"time"

	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/ec-systems/core.ledger.service/pkg/types"
	"github.com/shopspring/decimal"
)

const (
	TimeFormat = "02 Jan 2006 15:04:05 Z07:00" // "02 Jan 06 15:04:01 -0700"
)

type Transaction struct {
	tx      uint64
	key     string
	ID      types.ID      `swaggertype:"primitive,string"`
	Account types.Account `swaggertype:"primitive,string"`
	Holder  string
	Order   string `json:"Order,omitempty"`
	Item    string `json:"Item,omitempty"`

	Asset  types.Asset
	Amount decimal.Decimal

	Status    types.Status
	Modified  time.Time
	Created   time.Time
	Reference string `json:",omitempty"`
	User      string `json:",omitempty"`
}

func (tx *Transaction) Copy() *Transaction {
	return &Transaction{
		tx:  tx.tx,
		key: tx.key,

		ID:      tx.ID,
		Account: tx.Account,
		Holder:  tx.Holder,
		Order:   tx.Order,
		Item:    tx.Item,

		Asset:  tx.Asset,
		Amount: tx.Amount,

		Status:    tx.Status,
		Modified:  tx.Modified,
		Created:   tx.Created,
		Reference: tx.Reference,
		User:      tx.User,
	}
}

func (tx *Transaction) Parse(e *schema.Entry) error {
	return Unmarshal(e, tx)
}

func (t *Transaction) Bytes(format types.Format) ([]byte, error) {
	return Marshal(t, format, Version)
}

func (t *Transaction) SetTX(tx uint64) {
	t.tx = tx
}

func (t *Transaction) TX() uint64 {
	return t.tx
}

func (t *Transaction) SetKey(key string) {
	t.key = key
}

func (t *Transaction) Key() string {
	return t.key
}

func (t *Transaction) OrderRow(items bool) []string {
	row := []string{
		fmt.Sprintf("%v", t.TX()),
		t.Created.Format(TimeFormat),
		t.Order,
	}

	if items {
		row = append(row, t.Item)
		row = append(row, t.Asset.String())
		row = append(row, t.Status.String())
		row = append(row, t.Amount.String())
	}

	return row
}

func (t *Transaction) Change() []string {
	return []string{
		fmt.Sprintf("%v", t.TX()), t.Modified.Format(TimeFormat), t.Status.String()}
}
