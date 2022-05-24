package ledger_test

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"testing"
	"time"

	immudb "github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/stream"
	"github.com/ec-systems/core.ledger.tool/pkg/client"
	"github.com/ec-systems/core.ledger.tool/pkg/config"
	"github.com/ec-systems/core.ledger.tool/pkg/ledger"
	"github.com/ec-systems/core.ledger.tool/pkg/logger"
	"github.com/ec-systems/core.ledger.tool/pkg/types"
	"github.com/goombaio/namegenerator"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"golang.org/x/exp/maps"
)

const (
	CLIENT_OPTIONS_ADDRESS                  = "34.129.189.169"
	CLIENT_OPTIONS_PORT                     = 3322
	CLIENT_OPTIONS_USERNAME                 = "immudb"
	CLIENT_OPTIONS_PASSWORD                 = "VC8lHttILEVR6bSl"
	CLIENT_OPTIONS_MTLS                     = true
	CLIENT_OPTIONS_DATABASE                 = "test"
	CLIENT_OPTIONS_MTLS_OPTIONS_CERTIFICATE = "../../certs/tls.crt"
	CLIENT_OPTIONS_MTLS_OPTIONS_CLIENT_CAS  = "../../certs/ca.crt"
	CLIENT_OPTIONS_MTLS_OPTIONS_PKEY        = "../../certs/tls.key"
	CLIENT_OPTIONS_MTLS_OPTIONS_SERVERNAME  = "ledger-immudb-primary"
	CLIENT_OPTIONS_TOKEN_FILE_NAME          = "./token"
)

var (
	customer = randomName()

	// cfg = config.Config{
	// 	LogLevel:  logger.InfoLevel,
	// 	Assets:    types.DefaultAssetMap,
	// 	Statuses:  types.DefaultStatusMap,
	// 	BatchSize: 25,
	// 	ClientOptions: &immudb.Options{
	// 		Dir:                "./test_data",
	// 		Address:            CLIENT_OPTIONS_ADDRESS,
	// 		Port:               CLIENT_OPTIONS_PORT,
	// 		Username:           CLIENT_OPTIONS_USERNAME,
	// 		Password:           CLIENT_OPTIONS_PASSWORD,
	// 		Database:           CLIENT_OPTIONS_DATABASE,
	// 		MTLs:               CLIENT_OPTIONS_MTLS,
	// 		Auth:               true,
	// 		HealthCheckRetries: 5,
	// 		HeartBeatFrequency: time.Minute * 1,
	// 		StreamChunkSize:    stream.DefaultChunkSize,
	// 		MaxRecvMsgSize:     4 * 1024 * 1024,
	// 		TokenFileName:      CLIENT_OPTIONS_TOKEN_FILE_NAME,
	// 		MTLsOptions: immudb.MTLsOptions{
	// 			Certificate: CLIENT_OPTIONS_MTLS_OPTIONS_CERTIFICATE,
	// 			ClientCAs:   CLIENT_OPTIONS_MTLS_OPTIONS_CLIENT_CAS,
	// 			Pkey:        CLIENT_OPTIONS_MTLS_OPTIONS_PKEY,
	// 			Servername:  CLIENT_OPTIONS_MTLS_OPTIONS_SERVERNAME,
	// 		},
	// 		Config:      "configs/immuclient.toml",
	// 		DialOptions: []grpc.DialOption{grpc.WithInsecure()},
	// 	},
	// }

	db = "test"

	cfg = config.Config{
		LogLevel:  logger.InfoLevel,
		Assets:    types.DefaultAssetMap,
		Statuses:  types.DefaultStatusMap,
		BatchSize: 25,
		ClientOptions: &immudb.Options{
			Dir:                "./test_data",
			Address:            "localhost",
			Port:               CLIENT_OPTIONS_PORT,
			Username:           "immudb",
			Password:           "immudb",
			Database:           db,
			MTLs:               false,
			Auth:               true,
			HealthCheckRetries: 5,
			HeartBeatFrequency: time.Minute * 1,
			StreamChunkSize:    stream.DefaultChunkSize,
			MaxRecvMsgSize:     4 * 1024 * 1024,
			TokenFileName:      CLIENT_OPTIONS_TOKEN_FILE_NAME,
			Config:             "configs/immuclient.toml",
			DialOptions:        []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
		},
	}
)

func Test_Add(t *testing.T) {
	ctx := context.Background()
	client, err := new(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	assets := cfg.Assets
	l := ledger.New(client, ledger.SupportedAssets(assets), ledger.MultiAccounts(true))

	asset := randomAsset(assets)
	customer = randomName()
	order := "order1"
	item := "test2"
	amounts := randFloats(0, 10, 2)
	reference := "test"

	tx, err := l.Add(ctx, customer, asset, amounts[0],
		ledger.NewAccount(customer, asset),
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

	assert.Equal(t, customer, tx.Customer)
	assert.Equal(t, order, tx.Order)
	assert.Equal(t, item, tx.Item)
	assert.Equal(t, asset, tx.Asset)
	assert.Equal(t, amounts[0], tx.Amount)
	assert.Equal(t, reference, tx.Reference)
	assert.Equal(t, "ID:"+tx.ID, tx.Key())
	assert.Equal(t, types.Created, tx.Status)
	assert.Equal(t, ledger.Version, tx.Version)

	tx2, err := l.Get(ctx, tx.ID)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.NotNil(t, tx) {
		return
	}

	assert.False(t, tx.Date.IsZero())
	assert.False(t, tx2.Date.IsZero())

	if assert.Equal(t, tx.Date.Format(time.RFC3339), tx2.Date.Format(time.RFC3339)) {
		tx.Date = tx2.Date
	}

	assert.False(t, tx.Created.IsZero())
	assert.False(t, tx2.Created.IsZero())

	if assert.Equal(t, tx.Created.Format(time.RFC3339), tx2.Created.Format(time.RFC3339)) {
		tx.Created = tx2.Created
	}

	assert.Equal(t, *tx, *tx2)
}

func Test_Add_Account(t *testing.T) {
	ctx := context.Background()
	client, err := new(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	assets := cfg.Assets
	l := ledger.New(client, ledger.SupportedAssets(assets), ledger.MultiAccounts(true))

	asset := randomAsset(assets)
	customer = randomName()
	account, _ := types.NewAccount(customer, asset)
	order := "order1"
	item := "test2"
	amounts := randFloats(0, 10, 2)
	reference := "test"

	tx, err := l.Add(ctx, customer, asset, amounts[0],
		ledger.NewAccount(customer, asset),
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

	assert.Equal(t, customer, tx.Customer)
	assert.Equal(t, account, tx.Account)
	assert.Equal(t, order, tx.Order)
	assert.Equal(t, item, tx.Item)
	assert.Equal(t, asset, tx.Asset)
	assert.Equal(t, amounts[0], tx.Amount)
	assert.Equal(t, reference, tx.Reference)
	assert.Equal(t, "ID:"+tx.ID, tx.Key())
	assert.Equal(t, types.Created, tx.Status)
	assert.Equal(t, ledger.Version, tx.Version)

	tx2, err := l.Get(ctx, tx.ID)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.NotNil(t, tx) {
		return
	}

	assert.False(t, tx.Date.IsZero())
	assert.False(t, tx2.Date.IsZero())

	if assert.Equal(t, tx.Date.Format(time.RFC3339), tx2.Date.Format(time.RFC3339)) {
		tx.Date = tx2.Date
	}

	assert.False(t, tx.Created.IsZero())
	assert.False(t, tx2.Created.IsZero())

	if assert.Equal(t, tx.Created.Format(time.RFC3339), tx2.Created.Format(time.RFC3339)) {
		tx.Created = tx2.Created
	}

	assert.Equal(t, *tx, *tx2)
}

func Test_Remove(t *testing.T) {
	ctx := context.Background()
	client, err := new(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	assets := cfg.Assets
	l := ledger.New(client,
		ledger.SupportedAssets(assets),
		ledger.Overdraw(true),
	)

	asset := randomAsset(assets)
	customer = randomName()
	account, _ := types.NewAccount(customer, asset)
	order := "order1"
	item := "test2"
	amount := 1.23456
	reference := float64(99)

	tx, err := l.Remove(ctx, customer, asset, amount,
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

	assert.Equal(t, customer, tx.Customer)
	assert.Equal(t, account, tx.Account)
	assert.Equal(t, order, tx.Order)
	assert.Equal(t, item, tx.Item)
	assert.Equal(t, asset, tx.Asset)
	assert.Equal(t, -amount, tx.Amount)
	assert.Equal(t, reference, tx.Reference)
	assert.Equal(t, "ID:"+tx.ID, tx.Key())
	assert.Equal(t, types.Created, tx.Status)
	assert.Equal(t, ledger.Version, tx.Version)

	tx2, err := l.Get(ctx, tx.ID)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.NotNil(t, tx) {
		return
	}

	assert.False(t, tx.Date.IsZero())
	assert.False(t, tx2.Date.IsZero())

	if assert.Equal(t, tx.Date.Format(time.RFC3339), tx2.Date.Format(time.RFC3339)) {
		tx.Date = tx2.Date
	}

	assert.False(t, tx.Created.IsZero())
	assert.False(t, tx2.Created.IsZero())

	if assert.Equal(t, tx.Created.Format(time.RFC3339), tx2.Created.Format(time.RFC3339)) {
		tx.Created = tx2.Created
	}

	assert.Equal(t, *tx, *tx2)
}

func Test_Status(t *testing.T) {
	ctx := context.Background()
	client, err := new(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	assets := cfg.Assets
	l := ledger.New(client, ledger.SupportedAssets(assets), ledger.MultiAccounts(true))

	amounts := randFloats(0, 100, 1)

	asset := types.BNB
	customer = randomName()
	order := "my_order"
	item := "1"
	reference := struct {
		ID   int
		Name string
	}{12, "test"}

	tx1, err := l.Add(ctx, customer, asset, amounts[0],
		ledger.NewAccount(customer, asset),
		ledger.OrderID(order),
		ledger.OrderItemID(item),
		ledger.Reference(reference),
	)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.NotNil(t, tx1) {
		return
	}

	assert.Greater(t, tx1.TX(), uint64(0))

	time.Sleep(2 * time.Second)

	tx2, err := l.Status(ctx, tx1.ID, types.Settled)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.NotNil(t, tx2) {
		return
	}

	history := []*ledger.Transaction{}
	err = l.History(ctx, tx1.ID, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
		history = append(history, tx)
		return true, nil
	})
	if !assert.NoError(t, err) {
		return
	}

	txs := []*ledger.Transaction{}
	err = l.Transactions(ctx, tx2.Customer, types.AllAssets, types.AllAccounts, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
		txs = append(txs, tx)
		return true, nil
	})
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, txs, 1)
	assert.Equal(t, uint64(6), txs[0].TX())

}

func Test_History(t *testing.T) {
	ctx := context.Background()
	client, err := new(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	assets := cfg.Assets
	l := ledger.New(client, ledger.SupportedAssets(assets), ledger.MultiAccounts(true))

	amounts := randFloats(0, 100, 1)

	asset := types.BNB
	customer = randomName()
	order := "my_order"
	item := "1"
	reference := struct {
		ID   int
		Name string
	}{12, "test"}

	tx1, err := l.Add(ctx, customer, asset, amounts[0],
		ledger.NewAccount(customer, asset),
		ledger.OrderID(order),
		ledger.OrderItemID(item),
		ledger.Reference(reference),
	)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.NotNil(t, tx1) {
		return
	}

	assert.Greater(t, tx1.TX(), uint64(0))

	for i := 0; i < 10; i++ {
		tx2, err := l.Status(ctx, tx1.ID, types.Settled)
		if !assert.NoError(t, err) {
			return
		}

		if !assert.NotNil(t, tx2) {
			return
		}

		tx3, err := l.Status(ctx, tx1.ID, types.Finished)
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

}

func Test_Remove_MultipleAccounts(t *testing.T) {
	ctx := context.Background()
	client, err := new(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	assets := cfg.Assets
	l := ledger.New(client,
		ledger.SupportedAssets(assets),
		ledger.Overdraw(false),
		ledger.MultiAccounts(true),
	)

	asset := assets["XRP"]
	tx1, ok := add(ctx, t, l, randomName(), asset, 1)

	if !ok {
		return
	}

	_, ok = add(ctx, t, l, tx1.Customer, asset, 2, ledger.NewAccount("dummy", tx1.Asset))

	if !ok {
		return
	}

	_, err = l.Remove(ctx, tx1.Customer, tx1.Asset, 2)

	assert.NoError(t, err)

	_, err = l.Remove(ctx, tx1.Customer, tx1.Asset, 1)

	assert.NoError(t, err)

	b, err := l.Balance(ctx, tx1.Customer, types.AllAssets, types.AllAccounts, types.Created)

	assert.NoError(t, err)

	if assert.Contains(t, b, tx1.Asset) {
		assert.Equal(t, 0.0, b[tx1.Asset].Sum)
		assert.Equal(t, uint(4), b[tx1.Asset].Count)
	}

}

func Test_Overdraw(t *testing.T) {
	ctx := context.Background()
	client, err := new(ctx)
	if !assert.NoError(t, err) {
		return
	}

	amounts := randFloats(0, 2, 1)

	defer client.Close(ctx)
	assets := cfg.Assets
	l := ledger.New(client,
		ledger.SupportedAssets(assets),
		ledger.Overdraw(false),
	)

	asset := assets["XRP"]
	tx1, ok := add(ctx, t, l, randomName(), asset, amounts[0])

	if !ok {
		return
	}

	_, err = remove(ctx, t, l, tx1.Customer, tx1.Asset, amounts[0])

	if err != nil {
		return
	}

	_, err = l.Remove(ctx, tx1.Customer, tx1.Asset, amounts[0])

	assert.EqualError(t, err, fmt.Sprintf("balance too low to remove XRP %f for customer %v", amounts[0], tx1.Customer))
}

func Test_Overdraw_MultipleAccounts(t *testing.T) {
	ctx := context.Background()
	client, err := new(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	assets := cfg.Assets
	l := ledger.New(client,
		ledger.SupportedAssets(assets),
		ledger.Overdraw(false),
		ledger.MultiAccounts(),
	)

	asset := assets["XRP"]
	tx1, ok := add(ctx, t, l, randomName(), asset, 1)

	if !ok {
		return
	}

	_, ok = add(ctx, t, l, tx1.Customer, asset, 2, ledger.NewAccount(randomName(), tx1.Asset))

	if !ok {
		return
	}

	_, err = l.Remove(ctx, tx1.Customer, tx1.Asset, 3)

	assert.EqualError(t, err, "no account found with enough balance to remove XRP 3.000000 for customer "+tx1.Customer)

	_, err = l.Remove(ctx, tx1.Customer, tx1.Asset, 2)

	assert.NoError(t, err)

	_, err = l.Remove(ctx, tx1.Customer, tx1.Asset, 1)

	assert.NoError(t, err)
}

func Test_Accounts(t *testing.T) {
	ctx := context.Background()
	client, err := new(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	assets := cfg.Assets
	l := ledger.New(client, ledger.SupportedAssets(assets))

	asset := assets["XRP"]
	tx, err := l.Add(ctx, randomName(), asset, 1,
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

	asset = assets["BTC"]
	tx, err = l.Add(ctx, tx.Customer, asset, 1,
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

	accounts, err := l.Accounts(ctx, tx.Customer, tx.Asset)

	if !assert.NoError(t, err) {
		return
	}

	if !assert.NotNil(t, accounts) {
		return
	}

	assert.Len(t, accounts, 1)
	assert.Contains(t, accounts, account)

	accounts, err = l.Accounts(ctx, tx.Customer, types.AllAssets)

	if !assert.NoError(t, err) {
		return
	}

	if !assert.NotNil(t, accounts) {
		return
	}

	assert.Len(t, accounts, 2)
	assert.Contains(t, accounts, account)
	assert.Contains(t, accounts, tx.Account)
}

func Test_AccountInfo(t *testing.T) {
	ctx := context.Background()
	client, err := new(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	assets := cfg.Assets
	l := ledger.New(client, ledger.SupportedAssets(assets))

	asset := assets["XRP"]
	tx, err := l.Add(ctx, randomName(), asset, 1)
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
	assert.Equal(t, tx.Customer, info.Customer)
	assert.Equal(t, tx.Asset, info.Asset)

}

func Test_Transactions(t *testing.T) {
	ctx := context.Background()
	client, err := new(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)
	assets := cfg.Assets
	l := ledger.New(client,
		ledger.SupportedAssets(assets),
		ledger.MultiAccounts(),
	)

	asset1 := assets["XRP"]
	tx1, err := l.Add(ctx, randomName(), asset1, 0.2456,
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

	asset2 := assets["BTC"]
	tx2, err := l.Add(ctx, tx1.Customer, asset2, 1.5,
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

	tx3, err := l.Add(ctx, tx1.Customer, asset1, 0.0023,
		ledger.NewAccount(randomName(), asset1),
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

	err = l.Transactions(ctx, tx1.Customer, types.AllAssets, types.AllAccounts, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
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

	err = l.Transactions(ctx, tx1.Customer, tx1.Asset, types.AllAccounts, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
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

	err = l.Transactions(ctx, tx1.Customer, tx1.Asset, tx1.Account, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
		txs[tx.TX()] = tx
		return true, nil
	})

	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, txs, 1)

	assert.Equal(t, tx1.ID, txs[tx1.TX()].ID)
}

func Test_Balance(t *testing.T) {
	ctx := context.Background()
	client, err := new(ctx)
	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)
	assets := cfg.Assets
	l := ledger.New(client,
		ledger.SupportedAssets(assets),
		ledger.MultiAccounts(),
	)

	asset1 := assets["XRP"]
	tx1, err := l.Add(ctx, randomName(), asset1, 0.2456,
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

	asset2 := assets["BTC"]
	tx2, err := l.Add(ctx, tx1.Customer, asset2, 1.5,
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

	tx3, err := l.Add(ctx, tx1.Customer, asset1, 0.0023,
		ledger.NewAccount(randomName(), asset1),
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

	balance, err := l.Balance(ctx, tx1.Customer, types.AllAssets, types.AllAccounts, types.Created)

	if !assert.NoError(t, err) {
		return
	}

	if !assert.NotNil(t, balance) {
		return
	}

	if assert.Len(t, balance, 2) && assert.Contains(t, balance, asset1) && assert.Contains(t, balance, asset2) {
		acc1 := balance[asset1]
		acc2 := balance[asset2]

		assert.Equal(t, tx1.Amount+tx3.Amount, acc1.Sum)
		assert.Equal(t, uint(2), acc1.Count)
		assert.Equal(t, tx2.Amount, acc2.Sum)
		assert.Equal(t, uint(1), acc2.Count)

		if assert.Len(t, acc1.Accounts, 2) && assert.Contains(t, acc1.Accounts, tx1.Account) && assert.Contains(t, acc1.Accounts, tx3.Account) {
			assert.Equal(t, tx1.Amount, acc1.Accounts[tx1.Account].Sum)
			assert.Equal(t, uint(1), acc1.Accounts[tx1.Account].Count)
			assert.Equal(t, tx3.Amount, acc1.Accounts[tx3.Account].Sum)
			assert.Equal(t, uint(1), acc1.Accounts[tx3.Account].Count)
		}

		if assert.Len(t, acc2.Accounts, 1) && assert.Contains(t, acc2.Accounts, tx2.Account) {
			assert.Equal(t, tx2.Amount, acc2.Accounts[tx2.Account].Sum)
			assert.Equal(t, uint(1), acc2.Accounts[tx2.Account].Count)
		}
	}

}

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UTC().UnixNano())

	ctx := context.Background()
	var cl *client.Client
	var err error

	cl, err = client.New(ctx, cfg.ClientOptions.Username, cfg.ClientOptions.Password, "defaultdb",
		client.ClientOptions(cfg.ClientOptions),
		client.Limit(5),
	)

	if err != nil {
		log.Fatal(err)
	}

	exists, err := cl.DatabaseExist(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	if exists {
		log.Printf("Delete test database: %v", db)

		err := cl.UnloadDatabase(ctx, db)
		if err != nil {
			log.Fatal(err)
		}

		err = cl.DeleteDatabase(ctx, db)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Create test database: %v", db)

	err = cl.CreateDatabase(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	cl.Close(ctx)

	os.Exit(code)
}

func randomName() string {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	return nameGenerator.Generate()
}

func randomAsset(assets types.Assets) types.Asset {
	keys := maps.Keys(assets)
	index := rand.Intn(len(keys) - 1)
	return assets[keys[index]]
}

func randFloats(min, max float64, n int) []float64 {
	res := make([]float64, n)
	for i := range res {
		res[i] = math.Round((min+rand.Float64()*(max-min))*1000000000) / 1000000000
	}
	return res
}

func new(ctx context.Context) (*client.Client, error) {
	client, err := client.New(ctx, cfg.ClientOptions.Username, cfg.ClientOptions.Password, cfg.ClientOptions.Database,
		client.ClientOptions(cfg.ClientOptions),
		client.Limit(5),
	)
	if err != nil {
		return nil, fmt.Errorf("immudb client error: %v", err)
	}

	return client, nil
}

func add(ctx context.Context, t *testing.T, l *ledger.Ledger, customer string, asset types.Asset, amount float64, options ...ledger.TransactionOption) (*ledger.Transaction, bool) {
	tx, err := l.Add(ctx, customer, asset, amount, options...)
	if !assert.NoError(t, err) {
		return nil, false
	}

	if !assert.NotNil(t, tx) {
		return nil, false
	}

	return tx, true
}

func remove(ctx context.Context, t *testing.T, l *ledger.Ledger, customer string, asset types.Asset, amount float64, options ...ledger.TransactionOption) (*ledger.Transaction, error) {
	tx, err := l.Remove(ctx, customer, asset, amount, options...)
	if !assert.NoError(t, err) {
		return nil, err
	}

	if assert.NotNil(t, tx) {
		assert.Greater(t, tx.TX(), uint64(0))
	}

	return tx, nil
}

type TestLedger_CreateTx_Error_Test struct {
	name          string
	ledgerOptions []ledger.LedgerOption
	exec          func(context.Context, *testing.T, *ledger.Ledger, TestLedger_CreateTx_Error_Test) bool
	customer      string
	asset         types.Asset
	amount        float64
	options       []ledger.TransactionOption
	err           string
}

func TestLedger_CreateTx_Error(t *testing.T) {
	customer := randomName()

	tests := []TestLedger_CreateTx_Error_Test{
		{
			name:     "Empty customer",
			customer: "",
			asset:    types.XRP,
			amount:   1,
			options:  nil,
			err:      "accounts: customer is mandatory",
		},
		{
			name:     "Empty asset",
			customer: randomName(),
			asset:    types.AllAssets,
			amount:   1,
			options:  nil,
			err:      "invalid asset ''",
		},
		{
			name:     "Unknown asset",
			customer: randomName(),
			asset:    types.Asset("ABC"),
			amount:   1,
			options:  nil,
			err:      "invalid asset 'ABC'",
		},
		{
			name:     "Amount = 0",
			customer: customer,
			asset:    types.XRP,
			amount:   0,
			options:  nil,
			err:      fmt.Sprintf("transaction for customer %v with 0 XRP", customer),
		},
		{
			name:     "Unknown Account",
			customer: customer,
			asset:    types.XRP,
			amount:   1,
			options: []ledger.TransactionOption{
				ledger.Account("1234567890"),
			},
			err: "account 1234567890 not found",
		},
		{
			name:     "Invalid Account",
			customer: customer,
			asset:    types.XRP,
			amount:   1,
			ledgerOptions: []ledger.LedgerOption{
				ledger.MultiAccounts(),
			},
			options: []ledger.TransactionOption{
				ledger.Account("1234567890"),
			},
			err: "invalid checksum for account 1234567890",
		},
		{
			name:     "Inconsistent Customer Account Combination",
			customer: customer,
			asset:    types.XRP,
			amount:   1,
			err:      "invalid checksum for account 1234567890",
			exec: func(ctx context.Context, t *testing.T, l *ledger.Ledger, tt TestLedger_CreateTx_Error_Test) bool {
				tx1, err := l.Add(ctx, tt.customer+"_other", tt.asset, 10)
				if !assert.NoError(t, err) {
					return false
				}

				tx2, err := l.Add(ctx, tt.customer, tt.asset, 10)
				if !assert.NoError(t, err) {
					return false
				}

				_, err = l.CreateTx(ctx, tx2.Customer, tx2.Asset, tt.amount, ledger.Account(tx1.Account))
				assert.EqualError(t, err, fmt.Sprintf("invalid customer %v for account %v (%v)", tx2.Customer, tx1.Account, tx1.Customer))

				return true
			},
		},
		{
			name:     "Inconsistent Asset Account Combination",
			customer: customer,
			asset:    types.XRP,
			amount:   1,
			err:      "invalid checksum for account 1234567890",
			exec: func(ctx context.Context, t *testing.T, l *ledger.Ledger, tt TestLedger_CreateTx_Error_Test) bool {
				tx1, err := l.Add(ctx, tt.customer, types.Bitcoin, 10)
				if !assert.NoError(t, err) {
					return false
				}

				tx2, err := l.Add(ctx, tt.customer, tt.asset, 10)
				if !assert.NoError(t, err) {
					return false
				}

				_, err = l.CreateTx(ctx, tx2.Customer, tx2.Asset, tt.amount, ledger.Account(tx1.Account))
				assert.EqualError(t, err, fmt.Sprintf("invalid asset %v for account %v (%v)", tx2.Asset, tx1.Account, tx1.Asset))

				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client, err := new(ctx)
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
				_, err := l.CreateTx(ctx, tt.customer, tt.asset, tt.amount, tt.options...)
				assert.EqualError(t, err, tt.err)
			}

		})
	}
}
