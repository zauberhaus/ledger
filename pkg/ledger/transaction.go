package ledger

import (
	"time"

	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/ec-systems/core.ledger.server/pkg/types"
	"github.com/shopspring/decimal"
)

const (
	TimeFormat = "02 Jan 2006 15:04:05 Z07:00" // "02 Jan 06 15:04:01 -0700"
)

type Transaction struct {
	tx      uint64
	key     string
	ID      types.ID      `json:"ID" swaggertype:"primitive,string"`
	Account types.Account `json:"Account" swaggertype:"primitive,string"`
	Holder  string        `json:"Holder"`
	Order   string        `json:"Order,omitempty"`
	Item    string        `json:"Item,omitempty"`

	Asset  types.Asset     `json:"Asset"`
	Amount decimal.Decimal `json:"Amount"`

	Status    types.Status `json:"Status" swaggertype:"primitive,string"`
	Modified  *time.Time   `json:"Modified,omitempty"`
	Created   *time.Time   `json:"Created"`
	Reference string       `json:"Reference,omitempty"`
	User      string       `json:"User,omitempty"`
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
