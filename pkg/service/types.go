package service

import (
	"time"

	"github.com/ec-systems/core.ledger.service/pkg/ledger"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Holder struct {
	Name     string     `json:"Name"`
	Accounts []*Account `json:"Accounts,omitempty"`
}

type Account struct {
	Account string `json:"Account"`
	Asset   string `json:"Asset"`
}

type Balance struct {
	Asset    string            `json:"Asset"`
	Sum      decimal.Decimal   `json:"Sum"`
	Accounts []*AccountBalance `json:"Accounts,omitempty"`
	Count    uint              `json:"Count"`
}
type AccountBalance struct {
	ID    string          `json:"ID"`
	Count uint            `json:"Count"`
	Sum   decimal.Decimal `json:"Sum"`
}

type Transaction struct {
	ID      uuid.UUID `json:"ID"`
	Account string    `json:"Account"`
	Holder  string    `json:"Holder"`
	Order   string    `json:"Order,omitempty"`
	Item    string    `json:"Item,omitempty"`

	Asset  string          `json:"Asset"`
	Amount decimal.Decimal `json:"Amount"`

	Status    string    `json:"Status"`
	Modified  time.Time `json:"Modified,omitempty"`
	Created   time.Time `json:"Created"`
	Reference string    `json:"Reference,omitempty"`
	User      string    `json:"User,omitempty"`
}

func (t *Transaction) Set(l *ledger.Ledger, tx *ledger.Transaction) {
	t.ID = tx.ID.UUID
	t.Account = tx.Account.String()
	t.Holder = tx.Holder
	t.Order = tx.Order
	t.Item = tx.Item
	t.Asset = tx.Asset.String()
	t.Amount = tx.Amount
	t.Status = tx.Status.String(l.SupportedStatus())
	t.Modified = tx.Modified
	t.Created = tx.Created
	t.Reference = tx.Reference
	t.User = tx.User
}

type Asset struct {
	Symbol string `json:"Symbol"`
	Name   string `json:"Name"`
}

type Assets []Asset

func (a Assets) Len() int           { return len(a) }
func (a Assets) Less(i, j int) bool { return a[i].Symbol < a[j].Symbol }
func (a Assets) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type Status struct {
	ID   int    `json:"ID"`
	Name string `json:"Name"`
}

type Statuses []Status

func (s Statuses) Len() int           { return len(s) }
func (s Statuses) Less(i, j int) bool { return s[i].ID < s[j].ID }
func (s Statuses) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
