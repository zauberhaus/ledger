package ledger_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"hash/crc64"
	"math/rand"
	"testing"
	"time"

	immudb "github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/stream"
	"github.com/ec-systems/core.ledger.service/pkg/client"
	"github.com/ec-systems/core.ledger.service/pkg/config"
	"github.com/ec-systems/core.ledger.service/pkg/ledger"
	"github.com/ec-systems/core.ledger.service/pkg/ledger/index"
	"github.com/ec-systems/core.ledger.service/pkg/logger"
	"github.com/ec-systems/core.ledger.service/pkg/types"
	"github.com/goombaio/namegenerator"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	CLIENT_OPTIONS_ADDRESS                  = "localhost"
	CLIENT_OPTIONS_PORT                     = 3322
	CLIENT_OPTIONS_USERNAME                 = "immudb"
	CLIENT_OPTIONS_PASSWORD                 = "immudb"
	CLIENT_OPTIONS_MTLS                     = false
	CLIENT_OPTIONS_DATABASE                 = "test"
	CLIENT_OPTIONS_MTLS_OPTIONS_CERTIFICATE = "../../certs/tls.crt"
	CLIENT_OPTIONS_MTLS_OPTIONS_CLIENT_CAS  = "../../certs/ca.crt"
	CLIENT_OPTIONS_MTLS_OPTIONS_PKEY        = "../../certs/tls.key"
	CLIENT_OPTIONS_MTLS_OPTIONS_SERVERNAME  = "ledger-immudb-primary"
	CLIENT_OPTIONS_TOKEN_FILE_NAME          = "./token"
)

var (
	holder = randomName()

	zero  = decimal.Zero
	one   = decimal.NewFromInt(1)
	two   = decimal.NewFromInt(2)
	three = decimal.NewFromInt(3)

	cfg = config.Config{
		LogLevel:  logger.InfoLevel,
		Assets:    types.DefaultAssetMap,
		Statuses:  types.DefaultStatusMap,
		BatchSize: 25,
		Format:    types.Protobuf,
		ClientOptions: &immudb.Options{
			Dir:                "./test_data",
			Address:            CLIENT_OPTIONS_ADDRESS,
			Port:               CLIENT_OPTIONS_PORT,
			Username:           CLIENT_OPTIONS_USERNAME,
			Password:           CLIENT_OPTIONS_PASSWORD,
			Database:           CLIENT_OPTIONS_DATABASE,
			MTLs:               CLIENT_OPTIONS_MTLS,
			Auth:               true,
			HealthCheckRetries: 5,
			HeartBeatFrequency: time.Minute * 1,
			StreamChunkSize:    stream.DefaultChunkSize,
			MaxRecvMsgSize:     4 * 1024 * 1024,
			TokenFileName:      CLIENT_OPTIONS_TOKEN_FILE_NAME,
			MTLsOptions: immudb.MTLsOptions{
				Certificate: CLIENT_OPTIONS_MTLS_OPTIONS_CERTIFICATE,
				ClientCAs:   CLIENT_OPTIONS_MTLS_OPTIONS_CLIENT_CAS,
				Pkey:        CLIENT_OPTIONS_MTLS_OPTIONS_PKEY,
				Servername:  CLIENT_OPTIONS_MTLS_OPTIONS_SERVERNAME,
			},
			Config:      "configs/immuclient.toml",
			DialOptions: []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
		},
	}

	assets = types.Assets{}

	formats = []types.Format{
		types.JSON,
		types.Protobuf,
	}
)

func Test_Add(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	for _, f := range formats {
		t.Run(t.Name()+"_"+f.String(), func(t *testing.T) {

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.MultiAccounts(true),
				ledger.Format(f),
			)

			asset := randomAsset(assets)
			holder = randomName()
			order := "order1"
			item := "test2"
			amounts := randFloats(2, 100)
			reference := "test"

			length := len(amounts[0].String())
			_ = length

			tx, err := l.Add(ctx, holder, asset, amounts[0],
				ledger.OrderID(order),
				ledger.OrderItemID(item),
				ledger.Reference(reference),
			)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx) {
				return
			}

			assert.Greater(t, tx.TX(), uint64(0))

			assert.Equal(t, holder, tx.Holder)
			assert.Equal(t, order, tx.Order)
			assert.Equal(t, item, tx.Item)
			assert.Equal(t, asset, tx.Asset)
			assert.Equal(t, amounts[0], tx.Amount)
			assert.Equal(t, index.Key.ID(tx.ID), tx.Key())
			assert.Equal(t, types.Created, tx.Status)
			assert.Equal(t, reference, tx.Reference)

			tx2, err := l.Get(ctx, tx.ID)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx) {
				return
			}

			check(t, tx, tx2)
		})
	}
}

func Test_Add_Account(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	for _, f := range formats {
		t.Run(t.Name()+"_"+f.String(), func(t *testing.T) {

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.MultiAccounts(true),
				ledger.Format(f),
			)

			asset := randomAsset(assets)
			holder = randomName()
			account := newAccount(holder, asset)
			order := "order1"
			item := "test2"
			amounts := randFloats(2)
			reference := "test"

			tx, err := l.Add(ctx, holder, asset, amounts[0],
				ledger.Account(account),
				ledger.OrderID(order),
				ledger.OrderItemID(item),
				ledger.Reference(reference),
			)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx) {
				return
			}

			assert.Greater(t, tx.TX(), uint64(0))

			assert.Equal(t, holder, tx.Holder)
			assert.Equal(t, account, tx.Account)
			assert.Equal(t, order, tx.Order)
			assert.Equal(t, item, tx.Item)
			assert.Equal(t, asset, tx.Asset)
			assert.Equal(t, amounts[0], tx.Amount)
			assert.Equal(t, index.Key.ID(tx.ID), tx.Key())
			assert.Equal(t, types.Created, tx.Status)
			assert.Equal(t, reference, tx.Reference)

			tx2, err := l.Get(ctx, tx.ID)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx) {
				return
			}

			check(t, tx, tx2)
		})
	}
}

func Test_Remove(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	for _, f := range formats {
		t.Run(t.Name()+"_"+f.String(), func(t *testing.T) {

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.Overdraw(true),
				ledger.Format(f),
			)

			asset := randomAsset(assets)
			holder = randomName()
			account := newAccount(holder, asset)
			order := "order1"
			item := "test2"

			amount := randFloats(1)[0]
			assert.NoError(t, err)

			reference, err := types.NewReference(float64(99))
			if !assert.NoError(t, err) {
				return
			}

			tx, err := l.Remove(ctx, holder, asset, amount,
				ledger.OrderID(order),
				ledger.OrderItemID(item),
				ledger.Reference(reference.String()),
			)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx) {
				return
			}

			assert.Greater(t, tx.TX(), uint64(0))

			assert.Equal(t, holder, tx.Holder)
			assert.Equal(t, account, tx.Account)
			assert.Equal(t, order, tx.Order)
			assert.Equal(t, item, tx.Item)
			assert.Equal(t, asset, tx.Asset)
			assert.Equal(t, amount.Neg().String(), tx.Amount.String())
			assert.Equal(t, reference, types.Reference(tx.Reference))
			assert.Equal(t, index.Key.ID(tx.ID), tx.Key())
			assert.Equal(t, types.Created, tx.Status)

			tx2, err := l.Get(ctx, tx.ID)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx) {
				return
			}

			check(t, tx, tx2)
		})
	}
}

func Test_Cancel(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	for _, f := range formats {
		t.Run(t.Name()+"_"+f.String(), func(t *testing.T) {
			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.Overdraw(false),
				ledger.Format(f),
			)

			asset := randomAsset(assets)
			holder = randomName()
			order := "order1"
			item := "test2"

			amount := randFloats(1)[0]
			assert.NoError(t, err)

			reference, err := types.NewReference(float64(99))
			if !assert.NoError(t, err) {
				return
			}

			assert.NoError(t, err)

			tx, err := l.Add(ctx, holder, asset, amount,
				ledger.OrderID(order),
				ledger.OrderItemID(item),
				ledger.Reference(reference.String()),
			)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx) {
				return
			}

			assert.NotZero(t, tx.TX())
			assert.Equal(t, holder, tx.Holder)
			assert.Equal(t, order, tx.Order)
			assert.Equal(t, item, tx.Item)
			assert.Equal(t, asset, tx.Asset)
			assert.Equal(t, amount.String(), tx.Amount.String())
			assert.Equal(t, reference, types.Reference(tx.Reference))
			assert.Equal(t, index.Key.ID(tx.ID), tx.Key())
			assert.Equal(t, types.Created, tx.Status)

			tx2, err := l.Cancel(ctx, holder, asset, tx.Account, tx.ID)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx) {
				return
			}

			ref, err := types.NewReference(struct {
				ID     string
				Status types.Status
			}{
				tx.ID.String(),
				types.Canceled,
			})

			assert.NoError(t, err)

			assert.NotZero(t, tx2.TX())
			assert.Equal(t, holder, tx2.Holder)
			assert.Equal(t, tx.Account, tx2.Account)
			assert.Equal(t, order, tx2.Order)
			assert.Equal(t, item, tx2.Item)
			assert.Equal(t, asset, tx2.Asset)
			assert.Equal(t, amount.Neg().String(), tx2.Amount.String())
			assert.Equal(t, ref.String(), tx2.Reference)
			assert.NotEqual(t, index.Key.ID(tx.ID), tx2.Key())
			assert.Equal(t, types.Finished, tx2.Status)

			tx1, err := l.Get(ctx, tx.ID)
			if !assert.NoError(t, err) {
				return
			}

			check(t, tx, tx1)
		})
	}
}

func Test_Status(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	for _, f := range formats {
		t.Run(t.Name()+"_"+f.String(), func(t *testing.T) {

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.MultiAccounts(false),
				ledger.Format(f),
			)

			amounts := randFloats(1)

			asset := types.BNB
			holder = randomName()
			order := "my_order"
			item := "1"
			reference, err := types.NewReference(struct {
				ID   int
				Name string
			}{12, "test"})

			if !assert.NoError(t, err) {
				return
			}

			tx1, err := l.Add(ctx, holder, asset, amounts[0],
				ledger.OrderID(order),
				ledger.OrderItemID(item),
				ledger.Reference(reference.String()),
			)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx1) {
				return
			}

			assert.Greater(t, tx1.TX(), uint64(0))

			time.Sleep(2 * time.Second)

			tx1.Status = types.Finished

			tx2, err := l.Status(ctx, tx1, types.Finished)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx2) {
				return
			}

			history := []*ledger.Transaction{}
			err = l.History(ctx, tx2.ID, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
				history = append(history, tx)
				return true, nil
			})
			if !assert.NoError(t, err) {
				return
			}

			txs := []*ledger.Transaction{}
			err = l.Transactions(ctx, tx2.Holder, types.AllAssets, types.AllAccounts, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
				txs = append(txs, tx)
				return true, nil
			})
			if !assert.NoError(t, err) {
				return
			}

			assert.Len(t, txs, 1)
			assert.Equal(t, types.Finished, txs[0].Status)
		})
	}

}

func Test_History(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	for _, f := range formats {
		t.Run(t.Name()+"_"+f.String(), func(t *testing.T) {

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.MultiAccounts(false),
				ledger.Format(f),
			)

			amounts := randFloats(1)

			asset := types.BNB
			holder = randomName()
			order := "my_order"
			item := "1"

			reference, err := types.NewReference(struct {
				ID   int
				Name string
			}{12, "test"})

			if !assert.NoError(t, err) {
				return
			}

			tx1, err := l.Add(ctx, holder, asset, amounts[0],
				ledger.OrderID(order),
				ledger.OrderItemID(item),
				ledger.Reference(reference.String()),
			)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx1) {
				return
			}

			assert.Greater(t, tx1.TX(), uint64(0))

			for i := 0; i < 10; i++ {
				tx2, err := l.Status(ctx, tx1, types.Finished)
				if !assert.NoError(t, err) {
					return
				}

				if !assert.NotNil(t, tx2) {
					return
				}

				tx3, err := l.Status(ctx, tx1, types.Created)
				if !assert.NoError(t, err) {
					return
				}

				if !assert.NotNil(t, tx3) {
					return
				}
			}

			history := []*ledger.Transaction{}
			err = l.History(ctx, tx1.ID, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
				history = append(history, tx)
				return true, nil
			})
			if !assert.NoError(t, err) {
				return
			}

			if assert.Len(t, history, 21) {
				for _, h := range history {
					assert.Equal(t, tx1.ID, h.ID)
				}
			}
		})
	}

}

func Test_Remove_MultipleAccounts(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	for _, f := range formats {
		t.Run(t.Name()+"_"+f.String(), func(t *testing.T) {

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.Overdraw(false),
				ledger.MultiAccounts(true),
				ledger.Format(f),
			)

			amount := randFloats(1)[0]

			asset := randomAsset(assets)

			tx1, ok := add(ctx, t, l, randomName(), asset, amount)

			if !ok {
				return
			}

			account := newAccount("dummy", tx1.Asset)

			_, ok = add(ctx, t, l, tx1.Holder, asset, amount, ledger.Account(account))

			if !ok {
				return
			}

			_, err = l.Remove(ctx, tx1.Holder, tx1.Asset, amount)

			if !assert.NoError(t, err) {
				return
			}

			_, err = l.Remove(ctx, tx1.Holder, tx1.Asset, amount)

			if !assert.NoError(t, err) {
				return
			}

			b, err := l.Balance(ctx, tx1.Holder, types.AllAssets, types.AllAccounts, types.Created)

			assert.NoError(t, err)

			if assert.Contains(t, b, tx1.Asset) {
				assert.True(t, b[tx1.Asset].Sum.IsZero())
				assert.Equal(t, uint(4), b[tx1.Asset].Count)
			}
		})
	}

}

func Test_Overdraw(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	for _, f := range formats {
		t.Run(t.Name()+"_"+f.String(), func(t *testing.T) {

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.Overdraw(false),
				ledger.Format(f),
			)

			amount := randFloats(1)[0]
			copy := amount

			asset := randomAsset(assets)
			tx1, ok := add(ctx, t, l, randomName(), asset, amount)

			if !ok {
				return
			}

			assert.Equal(t, copy.String(), amount.String())

			_, err = remove(ctx, t, l, tx1.Holder, tx1.Asset, amount)

			if err != nil {
				return
			}

			assert.Equal(t, copy.String(), amount.String())

			_, err = l.Remove(ctx, tx1.Holder, tx1.Asset, amount)

			assert.EqualError(t, err, fmt.Sprintf("balance too low to remove %v %v for holder %v", asset, amount, tx1.Holder))
		})
	}
}

func Test_Overdraw_MultipleAccounts(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	for _, f := range formats {
		t.Run(t.Name()+"_"+f.String(), func(t *testing.T) {

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.Overdraw(false),
				ledger.MultiAccounts(),
				ledger.Format(f),
			)

			asset := randomAsset(assets)

			tx1, ok := add(ctx, t, l, randomName(), asset, one)

			if !ok {
				return
			}

			account := newAccount(randomName(), tx1.Asset)

			_, ok = add(ctx, t, l, tx1.Holder, asset, two, ledger.Account(account))

			if !ok {
				return
			}

			_, err = l.Remove(ctx, tx1.Holder, tx1.Asset, three)

			assert.EqualError(t, err, fmt.Sprintf("no account found with enough balance to remove %v %v for holder %v", asset, three, tx1.Holder))

			_, err = l.Remove(ctx, tx1.Holder, tx1.Asset, two)

			assert.NoError(t, err)

			_, err = l.Remove(ctx, tx1.Holder, tx1.Asset, one)

			assert.NoError(t, err)
		})
	}
}

func Test_Order(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	for _, f := range formats {
		t.Run(t.Name()+"_"+f.String(), func(t *testing.T) {

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.SupportedStatuses(cfg.Statuses),
				ledger.Format(f),
			)

			asset1 := randomAsset(assets)
			asset2 := randomAsset(assets)

			amounts := randFloats(4)

			tx1, ok := add(ctx, t, l, randomName(), asset1, amounts[0],
				ledger.OrderID("order1"),
				ledger.OrderItemID("001"),
			)

			if !ok {
				return
			}

			_, ok = add(ctx, t, l, tx1.Holder, asset2, amounts[1],
				ledger.OrderID("order1"),
				ledger.OrderItemID("002"),
			)

			if !ok {
				return
			}

			_, ok = add(ctx, t, l, tx1.Holder, asset2, amounts[2],
				ledger.OrderID("order2"),
				ledger.OrderItemID("001"),
			)

			if !ok {
				return
			}

			_, ok = add(ctx, t, l, tx1.Holder, asset1, amounts[3],
				ledger.OrderID("order2"),
				ledger.OrderItemID("001"),
			)

			if !ok {
				return
			}

			orders := []*ledger.Transaction{}
			err = l.Orders(ctx, tx1.Holder, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
				if assert.Equal(t, tx1.Holder, tx.Holder) {
					orders = append(orders, tx)
					return true, nil
				} else {
					return false, nil
				}
			})

			if !assert.NoError(t, err) {
				return
			}

			assert.Len(t, orders, 2)

			orders = []*ledger.Transaction{}
			err = l.OrderItems(ctx, tx1.Holder, tx1.Order, "", func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
				if assert.Equal(t, tx1.Holder, tx.Holder) && assert.Equal(t, tx1.Order, tx.Order) {
					orders = append(orders, tx)
					return true, nil
				} else {
					return false, nil
				}
			})

			if !assert.NoError(t, err) {
				return
			}

			assert.Len(t, orders, 2)
		})
	}

}

func Test_Accounts(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	for _, f := range formats {
		t.Run(t.Name()+"_"+f.String(), func(t *testing.T) {

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.MultiAccounts(false),
				ledger.Format(f),
			)

			asset := randomAsset(assets)
			tx, err := l.Add(ctx, randomName(), asset, one,
				ledger.OrderID("o1"),
				ledger.OrderItemID("001"),
			)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx) {
				return
			}

			assert.Greater(t, tx.TX(), uint64(0))

			asset = types.Bitcoin
			tx, err = l.Add(ctx, tx.Holder, asset, one,
				ledger.OrderID("o2"),
				ledger.OrderItemID("001"),
			)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx) {
				return
			}

			assert.Greater(t, tx.TX(), uint64(0))
			account := tx.Account

			accounts, err := l.Accounts(ctx, tx.Holder, tx.Asset)

			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, accounts) {
				return
			}

			assert.Len(t, accounts, 1)
			assert.Contains(t, accounts, account)

			accounts, err = l.Accounts(ctx, tx.Holder, types.AllAssets)

			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, accounts) {
				return
			}

			assert.Len(t, accounts, 2)
			assert.Contains(t, accounts, account)
			assert.Contains(t, accounts, tx.Account)
		})
	}
}

func Test_AccountInfo(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	for _, f := range formats {
		t.Run(t.Name()+"_"+f.String(), func(t *testing.T) {

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.Format(f),
			)

			asset := randomAsset(assets)
			tx, err := l.Add(ctx, randomName(), asset, one)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx) {
				return
			}

			assert.Greater(t, tx.TX(), uint64(0))

			info, err := l.AccountInfo(ctx, tx.Account)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, info) {
				return
			}

			assert.Equal(t, tx.Account, info.Account)
			assert.Equal(t, tx.Holder, info.Holder)
			assert.Equal(t, tx.Asset, info.Asset)
		})
	}

}

func Test_Asset_Balances(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	for _, f := range formats {
		t.Run(t.Name()+"_"+f.String(), func(t *testing.T) {

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.Format(f),
			)

			asset1 := randomAsset(assets)
			asset2 := randomAsset(assets)

			sumAsset1 := decimal.Zero
			amounts := randFloats(4)
			holder = randomName()

			for i := 0; i < len(amounts)-1; i++ {
				tx, ok := add(ctx, t, l, holder, asset1, amounts[i],
					ledger.OrderID("order1"),
					ledger.OrderItemID("001"),
				)

				if !ok {
					return
				}

				sumAsset1 = sumAsset1.Add(tx.Amount)
			}

			_, ok := add(ctx, t, l, holder, asset2, amounts[len(amounts)-1],
				ledger.OrderID("order1"),
				ledger.OrderItemID("002"),
			)

			if !ok {
				return
			}

			b, err := l.AssetBalance(ctx, types.AllAssets)

			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, b) {
				return
			}

			assert.True(t, len(b) >= 2)
			if assert.Contains(t, b, asset1) && assert.Contains(t, b, asset2) {
				assert.NotZero(t, b[asset1])
				assert.NotZero(t, b[asset2])
			}
		})
	}
}

func Test_Transactions(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	for _, f := range formats {
		t.Run(t.Name()+"_"+f.String(), func(t *testing.T) {

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.MultiAccounts(),
				ledger.Format(f),
			)

			amounts := randFloats(3)

			asset1 := randomAsset(assets)
			tx1, err := l.Add(ctx, randomName(), asset1, amounts[0],
				ledger.OrderID("o1"),
				ledger.OrderItemID("001"),
			)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx1) {
				return
			}

			assert.Greater(t, tx1.TX(), uint64(0))

			asset2 := types.Bitcoin
			tx2, err := l.Add(ctx, tx1.Holder, asset2, amounts[1],
				ledger.OrderID("o2"),
				ledger.OrderItemID("001"),
			)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx2) {
				return
			}

			assert.Greater(t, tx2.TX(), uint64(0))

			tx3, err := l.Add(ctx, tx1.Holder, asset1, amounts[2],
				ledger.Account(newAccount(randomName(), asset1)),
				ledger.OrderID("o2"),
				ledger.OrderItemID("001"),
			)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx3) {
				return
			}

			assert.Greater(t, tx3.TX(), uint64(0))

			txs := map[uint64]*ledger.Transaction{}

			err = l.Transactions(ctx, tx1.Holder, types.AllAssets, types.AllAccounts, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
				txs[tx.TX()] = tx
				return true, nil
			})

			if !assert.NoError(t, err) {
				return
			}

			assert.Len(t, txs, 3)

			assert.Equal(t, tx1.ID, txs[tx1.TX()].ID)
			assert.Equal(t, tx2.ID, txs[tx2.TX()].ID)
			assert.Equal(t, tx3.ID, txs[tx3.TX()].ID)

			txs = map[uint64]*ledger.Transaction{}

			err = l.Transactions(ctx, tx1.Holder, tx1.Asset, types.AllAccounts, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
				txs[tx.TX()] = tx
				return true, nil
			})

			if !assert.NoError(t, err) {
				return
			}

			assert.Len(t, txs, 2)

			assert.Equal(t, tx1.ID, txs[tx1.TX()].ID)
			assert.Equal(t, tx3.ID, txs[tx3.TX()].ID)

			txs = map[uint64]*ledger.Transaction{}

			err = l.Transactions(ctx, tx1.Holder, tx1.Asset, tx1.Account, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
				txs[tx.TX()] = tx
				return true, nil
			})

			if !assert.NoError(t, err) {
				return
			}

			assert.Len(t, txs, 1)

			assert.Equal(t, tx1.ID, txs[tx1.TX()].ID)
		})
	}
}

func Test_Balance(t *testing.T) {
	ctx := context.Background()
	client, err := newClient(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	for _, f := range formats {
		t.Run(t.Name()+"_"+f.String(), func(t *testing.T) {

			asset1 := types.Asset("BL1" + "_" + f.String())
			asset2 := types.Asset("BL2" + "_" + f.String())

			testAssets := map[types.Asset]string{
				asset1: asset1.String(),
				asset2: asset2.String(),
			}

			l := ledger.New(client,
				ledger.SupportedAssets(testAssets),
				ledger.MultiAccounts(),
				ledger.Format(f),
			)

			amounts := randFloats(3)

			tx1, err := l.Add(ctx, randomName(), asset1, amounts[0],
				ledger.OrderID("o1"),
				ledger.OrderItemID("001"),
			)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx1) {
				return
			}

			assert.Greater(t, tx1.TX(), uint64(0))

			tx2, err := l.Add(ctx, tx1.Holder, asset2, amounts[0],
				ledger.OrderID("o2"),
				ledger.OrderItemID("001"),
			)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx2) {
				return
			}

			assert.Greater(t, tx2.TX(), uint64(0))

			tx3, err := l.Add(ctx, tx1.Holder, asset1, amounts[0],
				ledger.Account(newAccount(randomName(), asset1)),
				ledger.OrderID("o2"),
				ledger.OrderItemID("001"),
			)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, tx3) {
				return
			}

			assert.Greater(t, tx3.TX(), uint64(0))

			balance, err := l.Balance(ctx, tx1.Holder, types.AllAssets, types.AllAccounts, types.Created)

			if !assert.NoError(t, err) {
				return
			}

			if !assert.NotNil(t, balance) {
				return
			}

			if assert.Len(t, balance, 2) && assert.Contains(t, balance, asset1) && assert.Contains(t, balance, asset2) {
				acc1 := balance[asset1]
				acc2 := balance[asset2]

				sum := zero.Add(tx1.Amount).Add(tx3.Amount)

				assert.Equal(t, sum.String(), acc1.Sum.String())
				assert.Equal(t, uint(2), acc1.Count)
				assert.Equal(t, tx2.Amount.String(), acc2.Sum.String())
				assert.Equal(t, uint(1), acc2.Count)

				if assert.Len(t, acc1.Accounts, 2) && assert.Contains(t, acc1.Accounts, tx1.Account) && assert.Contains(t, acc1.Accounts, tx3.Account) {
					assert.Equal(t, tx1.Amount.String(), acc1.Accounts[tx1.Account].Sum.String())
					assert.Equal(t, uint(1), acc1.Accounts[tx1.Account].Count)
					assert.Equal(t, tx3.Amount.String(), acc1.Accounts[tx3.Account].Sum.String())
					assert.Equal(t, uint(1), acc1.Accounts[tx3.Account].Count)
				}

				if assert.Len(t, acc2.Accounts, 1) && assert.Contains(t, acc2.Accounts, tx2.Account) {
					assert.Equal(t, tx2.Amount.String(), acc2.Accounts[tx2.Account].Sum.String())
					assert.Equal(t, uint(1), acc2.Accounts[tx2.Account].Count)
				}
			}
		})
	}

}

func randomName() string {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	return nameGenerator.Generate()
}

func randomAsset(assets types.Assets) types.Asset {
	if len(assets) == 0 {
		for k, v := range cfg.Assets {
			assets[k] = v
		}

	}

	keys := maps.Keys(assets)
	index := rand.Intn(len(keys) - 1)
	asset := keys[index]

	delete(assets, keys[index])

	return asset
}

func randFloats(n int, digits ...int) []decimal.Decimal {

	d1 := int(1 + uint8(rand.Float64()*3))
	d2 := 8

	if len(digits) == 1 {
		d1 := int(1 + uint8(rand.Float64()*10))
		d2 = digits[0] - d1
	} else {
		if len(digits) > 0 {
			d2 = digits[0]
		}

		if len(digits) > 1 {
			d1 = digits[1]
		}
	}

	results := make([]decimal.Decimal, n)

	for i1 := range results {
		number := make([]byte, d1+d2+1)

		for i2 := range number {
			if i2 == d1+1 {
				number[i2] = '.'
			} else {
				number[i2] = 48 + uint8(rand.Float64()*10)
			}
		}

		tmp := string(number)
		value, err := decimal.NewFromString(tmp)
		if err == nil {
			results[i1] = value
		}
	}

	return results
}

func newClient(ctx context.Context) (*client.Client, error) {
	client, err := client.New(ctx, cfg.ClientOptions.Username, cfg.ClientOptions.Password, cfg.ClientOptions.Database,
		client.ClientOptions(cfg.ClientOptions),
		client.Limit(5),
	)
	if err != nil {
		return nil, fmt.Errorf("immudb client error: %v", err)
	}

	return client, nil
}

func add(ctx context.Context, t *testing.T, l *ledger.Ledger, holder string, asset types.Asset, amount decimal.Decimal, options ...ledger.TransactionOption) (*ledger.Transaction, bool) {
	tx, err := l.Add(ctx, holder, asset, amount, options...)
	if !assert.NoError(t, err) {
		return nil, false
	}

	if !assert.NotNil(t, tx) {
		return nil, false
	}

	return tx, true
}

func remove(ctx context.Context, t *testing.T, l *ledger.Ledger, holder string, asset types.Asset, amount decimal.Decimal, options ...ledger.TransactionOption) (*ledger.Transaction, error) {
	tx, err := l.Remove(ctx, holder, asset, amount, options...)
	if !assert.NoError(t, err) {
		return nil, err
	}

	if assert.NotNil(t, tx) {
		assert.Greater(t, tx.TX(), uint64(0))
	}

	return tx, nil
}

func check(t *testing.T, tx *ledger.Transaction, tx2 *ledger.Transaction) bool {

	if tx.Status == types.Created {
		if !assert.Nil(t, tx.Modified) {
			return false
		}
	} else {
		if !assert.NotNil(t, tx.Modified) {
			return false
		}
	}

	if tx2.Status == types.Created {
		if !assert.Nil(t, tx2.Modified) {
			return false
		}
	} else {
		if !assert.NotNil(t, tx2.Modified) {
			return false
		}
	}

	if tx.Modified != nil && tx2.Modified != nil {
		if assert.Equal(t, tx.Modified.Format(time.RFC3339), tx2.Modified.Format(time.RFC3339)) {
			tx.Modified = tx2.Modified
		} else {
			return false
		}
	} else if !assert.Nil(t, tx.Modified) && !assert.Nil(t, tx2.Modified) {
		return false
	}

	if assert.False(t, tx.Created.IsZero()) || assert.False(t, tx2.Created.IsZero()) {
		return false
	}

	if assert.Equal(t, tx.Created.Format(time.RFC3339), tx2.Created.Format(time.RFC3339)) {
		tx.Created = tx2.Created
	} else {
		return false
	}

	if assert.Equal(t, tx.Amount.String(), tx2.Amount.String()) {
		tx.Amount = tx2.Amount
		if assert.Equal(t, tx, tx2) {
			return true
		}
	}

	return false
}

func newAccount(holder string, asset types.Asset) types.Account {
	hash := crc64.New(crc64.MakeTable(crc64.ECMA))
	hash.Write([]byte(holder))
	hash.Write([]byte(asset))
	account := hex.EncodeToString(hash.Sum([]byte{0})[1:])

	chk := types.Account(account + "00").Checksum()

	return types.Account(fmt.Sprintf("%v%02d", account, chk))
}
