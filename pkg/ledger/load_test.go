package ledger_test

import (
	"context"
	"testing"
	"time"

	"github.com/ec-systems/core.ledger.server/pkg/ledger"
	"github.com/ec-systems/core.ledger.server/pkg/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func Test_Add_Remove_100(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	l := ledger.New(client,
		ledger.SupportedAssets(cfg.Assets),
		ledger.MultiAccounts(true),
		ledger.Format(types.Protobuf),
	)

	holder = randomName()

	asset := randomAsset(assets)

	runs := 20

	amounts1 := randFloats(runs, 8, 18)

	start := time.Now()
	cnt := 0

	for i := 0; i < runs; i++ {

		if i%10 == 0 {
			asset = randomAsset(assets)
		}

		half := amounts1[i].Div(decimal.NewFromFloat(2))
		balance := half.Add(half)
		h := half.String()
		b1 := balance.String()
		b2 := amounts1[i].String()

		if !assert.Equal(t, amounts1[i].String(), balance.String()) {
			return
		}

		diff := amounts1[i].Div(balance).String()

		_ = diff
		_ = h
		_ = b1
		_ = b2

		tx, err := l.Add(ctx, holder, asset, half)
		if !assert.NoError(t, err) {
			return
		}

		if !assert.NoError(t, err) || !assert.NotNil(t, tx) || !assert.NotZero(t, tx) {
			return
		}

		cnt++

		tx, err = l.Add(ctx, holder, asset, half)
		if !assert.NoError(t, err) {
			return
		}

		if !assert.NoError(t, err) || !assert.NotNil(t, tx) || !assert.NotZero(t, tx) {
			return
		}

		cnt++

		tx, err = l.Remove(ctx, holder, asset, balance)
		if !assert.NoError(t, err) {
			return
		}

		if !assert.NoError(t, err) || !assert.NotNil(t, tx) || !assert.NotZero(t, tx) {
			return
		}

		cnt++
	}

	seconds := time.Since(start).Seconds()
	ps := float64(cnt) / seconds

	t.Logf("Took %.3f seconds to create %v transactions\n", seconds, cnt)
	t.Logf("%.2f transactions per second\n", ps)
}
