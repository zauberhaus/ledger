package ledger

import (
	"github.com/ec-systems/core.ledger.server/pkg/types"
)

type LedgerOption interface {
	Set(*Ledger)
}

type LedgerOptionFunc func(*Ledger)

func (f LedgerOptionFunc) Set(l *Ledger) {
	f(l)
}

func Format(value types.Format) LedgerOption {
	return LedgerOptionFunc(func(l *Ledger) {
		l.format = value
	})
}

func ReadOnly(value ...bool) LedgerOption {
	return LedgerOptionFunc(func(l *Ledger) {
		if len(value) == 0 {
			l.readOnly = true
		} else {
			l.readOnly = value[0]
		}
	})
}

func Overdraw(value ...bool) LedgerOption {
	return LedgerOptionFunc(func(l *Ledger) {
		if len(value) == 0 {
			l.overdraw = true
		} else {
			l.overdraw = value[0]
		}
	})
}

func MultiAccounts(value ...bool) LedgerOption {
	return LedgerOptionFunc(func(l *Ledger) {
		if len(value) == 0 {
			l.multi = true
		} else {
			l.multi = value[0]
		}
	})
}

func SupportedAssets(assets types.Assets) LedgerOption {
	return LedgerOptionFunc(func(l *Ledger) {
		if len(assets) > 0 {
			l.assets = assets
		}
	})
}

func SupportedStatuses(statuses types.Statuses) LedgerOption {
	return LedgerOptionFunc(func(l *Ledger) {
		if len(statuses) > 0 {
			l.statuses = statuses
		}
	})
}

func Collector(collectors ...types.MetricsCollector) LedgerOption {
	return LedgerOptionFunc(func(l *Ledger) {
		l.collectors = append(l.collectors, collectors...)
	})
}

type TransactionOption interface {
	Set(*Transaction)
}

type TransactionOptionFunc func(*Transaction)

func (f TransactionOptionFunc) Set(tx *Transaction) {
	f(tx)
}

func Account(account types.Account) TransactionOption {
	return TransactionOptionFunc(func(tx *Transaction) {
		tx.Account = account
	})
}

func Reference(ref string) TransactionOption {
	return TransactionOptionFunc(func(tx *Transaction) {
		tx.Reference = ref
	})
}

func OrderID(id string) TransactionOption {
	return TransactionOptionFunc(func(tx *Transaction) {
		tx.Order = id
	})
}

func OrderItemID(id string) TransactionOption {
	return TransactionOptionFunc(func(tx *Transaction) {
		tx.Item = id
	})
}
