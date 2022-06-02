package types

import "github.com/shopspring/decimal"

type AccountBalance struct {
	Count uint
	Sum   decimal.Decimal
}

type Balance struct {
	Sum      decimal.Decimal
	Accounts map[Account]*AccountBalance
	Count    uint
}

func NewBalance(status Status) *Balance {
	return &Balance{
		Sum:      decimal.Zero,
		Count:    0,
		Accounts: map[Account]*AccountBalance{},
	}
}

func NewAccountBalance() *AccountBalance {
	return &AccountBalance{
		Count: 0,
		Sum:   decimal.Zero,
	}
}

func (b *Balance) Add(account Account, amount decimal.Decimal, status Status) {
	b.Sum = b.Sum.Add(amount)
	b.Count++

	acc, ok := b.Accounts[account]
	if !ok {
		acc = NewAccountBalance()
		b.Accounts[account] = acc
	}

	acc.Sum = acc.Sum.Add(amount)
	acc.Count++
}
