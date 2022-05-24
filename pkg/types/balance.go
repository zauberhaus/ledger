package types

type AccountBalance struct {
	Count uint
	Sum   float64
}

type Balance struct {
	Sum      float64
	Status   Status
	Accounts map[Account]*AccountBalance
	Count    uint
}

func NewBalance(status Status) *Balance {
	return &Balance{
		Status:   status,
		Accounts: map[Account]*AccountBalance{},
	}
}

func (b *Balance) Add(account Account, amount float64, status Status) {
	if status >= b.Status {
		b.Sum += amount
		b.Count++

		acc, ok := b.Accounts[account]
		if !ok {
			acc = &AccountBalance{}
			b.Accounts[account] = acc
		}

		acc.Sum += amount
		acc.Count++
	}
}
