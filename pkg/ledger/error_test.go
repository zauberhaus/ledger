package ledger_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ec-systems/core.ledger.server/pkg/ledger"
	"github.com/ec-systems/core.ledger.server/pkg/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type TestLedger_CreateTx_Error_Test struct {
	name          string
	ledgerOptions []ledger.LedgerOption
	exec          func(context.Context, *testing.T, *ledger.Ledger, TestLedger_CreateTx_Error_Test) bool
	holder        string
	asset         types.Asset
	amount        float64
	options       []ledger.TransactionOption
	err           string
}

func TestLedger_CreateTx_Error(t *testing.T) {
	holder := randomName()

	tests := []TestLedger_CreateTx_Error_Test{
		{
			name:    "Empty holder",
			holder:  "",
			asset:   types.XRP,
			amount:  1,
			options: nil,
			err:     "accounts: holder is mandatory",
		},
		{
			name:    "Empty asset",
			holder:  randomName(),
			asset:   types.AllAssets,
			amount:  1,
			options: nil,
			err:     "invalid asset ''",
		},
		{
			name:    "Unknown asset",
			holder:  randomName(),
			asset:   types.Asset("dummy"),
			amount:  1,
			options: nil,
			err:     "invalid asset 'dummy'",
		},
		{
			name:    "Amount = 0",
			holder:  holder,
			asset:   types.XRP,
			amount:  0,
			options: nil,
			err:     fmt.Sprintf("transaction for holder %v with 0 XRP", holder),
		},
		{
			name:   "Unknown Account",
			holder: holder,
			asset:  types.XRP,
			amount: 1,
			options: []ledger.TransactionOption{
				ledger.Account("1234567890"),
			},
			err: "account 1234567890 not found",
		},
		{
			name:   "Invalid Account",
			holder: holder,
			asset:  types.XRP,
			amount: 1,
			ledgerOptions: []ledger.LedgerOption{
				ledger.MultiAccounts(),
			},
			options: []ledger.TransactionOption{
				ledger.Account("1234567890"),
			},
			err: "invalid checksum for account 1234567890",
		},
		{
			name:   "Inconsistent Holder Account Combination",
			holder: holder,
			asset:  types.XRP,
			amount: 1,
			err:    "invalid checksum for account 1234567890",
			exec: func(ctx context.Context, t *testing.T, l *ledger.Ledger, tt TestLedger_CreateTx_Error_Test) bool {
				amount := randFloats(1)[0]

				tx1, err := l.Add(ctx, tt.holder+"_other", tt.asset, amount)
				if !assert.NoError(t, err) {
					return false
				}

				tx2, err := l.Add(ctx, tt.holder, tt.asset, amount)
				if !assert.NoError(t, err) {
					return false
				}

				_, err = l.CreateTx(ctx, tx2.Holder, tx2.Asset, decimal.NewFromFloat(tt.amount), ledger.Account(tx1.Account))
				assert.EqualError(t, err, fmt.Sprintf("invalid holder %v for account %v (%v)", tx2.Holder, tx1.Account, tx1.Holder))

				return true
			},
		},
		{
			name:   "Inconsistent Asset Account Combination",
			holder: holder,
			asset:  types.XRP,
			amount: 1,
			err:    "invalid checksum for account 1234567890",
			exec: func(ctx context.Context, t *testing.T, l *ledger.Ledger, tt TestLedger_CreateTx_Error_Test) bool {
				amount := randFloats(1)[0]

				tx1, err := l.Add(ctx, tt.holder, types.Bitcoin, amount)
				if !assert.NoError(t, err) {
					return false
				}

				tx2, err := l.Add(ctx, tt.holder, tt.asset, amount)
				if !assert.NoError(t, err) {
					return false
				}

				_, err = l.CreateTx(ctx, tx2.Holder, tx2.Asset, decimal.NewFromFloat(tt.amount), ledger.Account(tx1.Account))
				assert.EqualError(t, err, fmt.Sprintf("invalid asset %v for account %v (%v)", tx2.Asset, tx1.Account, tx1.Asset))

				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client, err := newClient(ctx)
			if !assert.NoError(t, err) {
				return
			}

			defer client.Close(ctx)

			assets := cfg.Assets
			ledgerOptions := []ledger.LedgerOption{
				ledger.SupportedAssets(assets),
			}

			ledgerOptions = append(ledgerOptions, tt.ledgerOptions...)

			l := ledger.New(client, ledgerOptions...)

			if tt.exec != nil {
				tt.exec(ctx, t, l, tt)
			} else {
				_, err := l.CreateTx(ctx, tt.holder, tt.asset, decimal.NewFromFloat(tt.amount), tt.options...)
				assert.EqualError(t, err, tt.err)
			}

		})
	}
}
