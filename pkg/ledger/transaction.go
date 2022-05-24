package ledger

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ec-systems/core.ledger.tool/pkg/types"
)

const (
	Version = uint16(1)

	TimeFormat = "02 Jan 06 15:04:00 -0700"
)

type Transaction struct {
	tx       uint64 `json:""`
	key      string `json:""`
	ID       string
	Account  types.Account
	Customer string
	Order    string
	Item     string

	Asset  types.Asset
	Amount float64

	Status    types.Status
	Date      time.Time
	Created   time.Time
	Reference interface{}
	Version   uint16
}

func ParseTransaction(id uint64, key string, data []byte) (*Transaction, error) {
	tx := &Transaction{}

	err := json.Unmarshal(data, tx)
	if err != nil {
		return nil, fmt.Errorf("parse transaction error: %v", err)
	}

	tx.tx = id
	tx.key = key

	return tx, nil
}

func (t *Transaction) Bytes() []byte {
	data, err := json.Marshal(t)
	if err != nil {
		return []byte(err.Error())
	}

	return data
}

func (t *Transaction) TX() uint64 {
	return t.tx
}

func (t *Transaction) Key() string {
	return t.key
}

func (t *Transaction) Change() []string {
	return []string{fmt.Sprintf("%v", t.TX()), t.Date.Format(TimeFormat), t.Status.String()}
}

func (t *Transaction) Row(keys bool) []string {
	row := []string{}

	row = append(row, fmt.Sprintf("%v", t.tx))
	row = append(row, t.ID)
	row = append(row, t.Created.Format(TimeFormat))
	row = append(row, t.Date.Format(TimeFormat))
	row = append(row, t.Customer)
	row = append(row, string(t.Account))
	row = append(row, t.Order)
	row = append(row, t.Item)
	row = append(row, string(t.Asset))
	row = append(row, string(t.Status.String()))
	row = append(row, fmt.Sprintf("%.8f", t.Amount))

	if t.Reference != nil {
		data, err := json.Marshal(t.Reference)
		if err != nil {
			row = append(row, fmt.Sprintf("%v", err))
		} else {
			row = append(row, string(data))
		}
	} else {
		row = append(row, "")
	}

	if keys {
		row = append(row, t.key)
	}

	return row
}
