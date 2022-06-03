package service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	immudb "github.com/codenotary/immudb/pkg/client"
	"github.com/goombaio/namegenerator"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"

	"github.com/codenotary/immudb/pkg/stream"
	"github.com/ec-systems/core.ledger.service/pkg/client"
	"github.com/ec-systems/core.ledger.service/pkg/config"
	"github.com/ec-systems/core.ledger.service/pkg/ledger"
	"github.com/ec-systems/core.ledger.service/pkg/logger"
	"github.com/ec-systems/core.ledger.service/pkg/service"
	"github.com/ec-systems/core.ledger.service/pkg/types"
	"github.com/phayes/freeport"
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
	cfg = config.Config{
		LogLevel:  logger.InfoLevel,
		Assets:    types.DefaultAssetMap,
		Statuses:  types.DefaultStatusMap,
		BatchSize: 25,
		Format:    types.JSON,
		Service: config.ServiceConfig{
			Device: "",
			Port:   12345,
			MTls:   nil,
		},
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

	url        = ""
	metricsUrl = ""

	assets = types.Assets{}
)

func Test_Assets(t *testing.T) {
	resp, err := http.Get(url + "/info/assets")
	if !assert.NoError(t, err) {
		return
	}

	var assets service.Assets
	err = json.NewDecoder(resp.Body).Decode(&assets)
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, assets, len(cfg.Assets))

	for _, v := range assets {
		if !assert.Contains(t, cfg.Assets, types.Asset(v.Symbol)) {
			return
		}
	}
}

func Test_Status(t *testing.T) {
	resp, err := http.Get(url + "/info/statuses")
	if !assert.NoError(t, err) {
		return
	}

	var statuses []service.Status
	err = json.NewDecoder(resp.Body).Decode(&statuses)
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, statuses, len(cfg.Statuses))

	for _, v := range statuses {
		name := v.Name
		status := types.Status(v.ID)
		if !assert.Contains(t, cfg.Statuses, name) || assert.Equal(t, cfg.Statuses[name], status) {
			return
		}
	}
}

func Test_Add_Remove_Balance(t *testing.T) {
	holder := randomName()
	asset := randomAsset()
	order := randomName()
	item := randomName()
	ref := randomName()

	amount1, _ := decimal.NewFromString("1.0")

	resp, err := put("/accounts/%v/%v/%v?order=%v&item=%v&ref=%v", holder, asset, amount1, order, item, ref)
	if !assert.NoError(t, err) || !assert.Equal(t, resp.StatusCode, 200) {
		return
	}

	var tx1 ledger.Transaction
	err = json.NewDecoder(resp.Body).Decode(&tx1)
	if !assert.NoError(t, err) {
		return
	}

	assert.NotEmpty(t, tx1.ID)
	assert.NotEmpty(t, tx1.Account)

	assert.Equal(t, holder, tx1.Holder)
	assert.Equal(t, order, tx1.Order)
	assert.Equal(t, item, tx1.Item)
	assert.Equal(t, asset, tx1.Asset)
	assert.Equal(t, amount1.String(), tx1.Amount.String())
	assert.Equal(t, types.Created, tx1.Status)
	assert.Equal(t, ref, tx1.Reference)

	amount2 := amount1.Div(decimal.NewFromFloat(2))

	resp, err = del("/accounts/%v/%v/%v?order=%v&item=%v&ref=%v", holder, asset, amount2, order, item, ref)
	if !assert.NoError(t, err) || !assert.Equal(t, resp.StatusCode, 200) {
		return
	}

	var tx2 ledger.Transaction
	err = json.NewDecoder(resp.Body).Decode(&tx2)
	if !assert.NoError(t, err) {
		return
	}

	assert.NotEmpty(t, tx2.ID)
	assert.Equal(t, tx1.Account, tx2.Account)

	assert.Equal(t, holder, tx2.Holder)
	assert.Equal(t, order, tx2.Order)
	assert.Equal(t, item, tx2.Item)
	assert.Equal(t, asset, tx2.Asset)
	assert.Equal(t, amount2.Neg().String(), tx2.Amount.String())
	assert.Equal(t, types.Created, tx2.Status)
	assert.Equal(t, ref, tx2.Reference)

	resp, err = get("/accounts/%v/%v", holder, asset)
	if !assert.NoError(t, err) || !assert.Equal(t, resp.StatusCode, 200) {
		return
	}

	var balances map[types.Asset]*types.Balance
	err = json.NewDecoder(resp.Body).Decode(&balances)
	if !assert.NoError(t, err) {
		return
	}

	assert.Contains(t, balances, asset)
	assert.Equal(t, uint(2), balances[asset].Count)
	assert.Equal(t, amount1.Add(amount2.Neg()).String(), balances[asset].Sum.String())
	assert.Len(t, balances[asset].Accounts, 1)

	asset2 := randomAsset()
	amount3, _ := decimal.NewFromString("0.048724761927846319849328746")

	resp, err = put("/accounts/%v/%v/%v", holder, asset2, amount3)
	if !assert.NoError(t, err) || !assert.Equal(t, resp.StatusCode, 200) {
		return
	}

	var tx3 ledger.Transaction
	err = json.NewDecoder(resp.Body).Decode(&tx3)
	if !assert.NoError(t, err) {
		return
	}

	assert.NotEmpty(t, tx3.ID)
	assert.NotEqual(t, tx1.Account, tx3.Account)

	resp, err = get("/accounts/%v", holder)
	if !assert.NoError(t, err) || !assert.Equal(t, resp.StatusCode, 200) {
		return
	}

	var balances2 map[types.Asset]*types.Balance
	err = json.NewDecoder(resp.Body).Decode(&balances2)
	if !assert.NoError(t, err) {
		return
	}

	assert.Contains(t, balances2, asset)
	assert.Contains(t, balances2, asset2)

	resp, err = patch("/accounts/%v/%v/%v/%v/%v", tx3.Holder, tx3.Asset, tx3.Account, tx3.ID, types.Finished)
	if !assert.NoError(t, err) || !assert.Equal(t, resp.StatusCode, 200) {
		return
	}

	var tx4 ledger.Transaction
	err = json.NewDecoder(resp.Body).Decode(&tx4)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, tx3.ID, tx4.ID)
	assert.Equal(t, types.Finished, tx4.Status)

	time.Sleep(500 * time.Millisecond)

	resp, err = get("/accounts/%v/%v/%v/%v", tx3.Holder, tx3.Asset, tx3.Account, tx3.ID)
	if !assert.NoError(t, err) || !assert.Equal(t, resp.StatusCode, 200) {
		return
	}

	var history []*ledger.Transaction
	err = json.NewDecoder(resp.Body).Decode(&history)
	if !assert.NoError(t, err) {
		return
	}

	if assert.NotNil(t, history) && assert.Len(t, history, 2) {
		assert.Equal(t, types.Created, history[0].Status)
		assert.Equal(t, types.Finished, history[1].Status)
	}
}

func TestMain(m *testing.M) {
	if url == "" {

		rand.Seed(time.Now().UTC().UnixNano())
		ctx := context.Background()

		client, port, metrics, err := start(ctx, &cfg)
		if err != nil {
			log.Fatal(err)
		}

		defer client.Close(ctx)

		time.Sleep(100 * time.Millisecond)

		url = fmt.Sprintf("http://localhost:%d", port)
		metricsUrl = fmt.Sprintf("http://localhost:%d", metrics)
	}

	code := m.Run()

	os.Exit(code)
}

func start(ctx context.Context, c *config.Config) (*client.Client, int, int, error) {
	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, 0, 0, err
	}

	metrics, err := freeport.GetFreePort()
	if err != nil {
		return nil, 0, 0, err
	}

	client, err := client.New(ctx, c.ClientOptions.Username, c.ClientOptions.Password, c.ClientOptions.Database,
		client.ClientOptions(c.ClientOptions),
		client.Limit(25),
	)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("database client error: %v", err)
	}

	l := ledger.New(client,
		ledger.SupportedAssets(c.Assets),
		ledger.SupportedStatuses(c.Statuses),
	)

	scfg := cfg.Service
	scfg.Port = port
	scfg.Metrics = metrics

	svc, err := service.NewLedgerService(ctx, l, &scfg)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("service error: %v", err)
	}

	go svc.Start()

	return client, port, metrics, nil
}

func randomName() string {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	return nameGenerator.Generate()
}

func randomAsset() types.Asset {
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

func call(method string, format string, args ...interface{}) (*http.Response, error) {
	client := &http.Client{}
	url := fmt.Sprintf(url+format, args...)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)

}

func put(format string, args ...interface{}) (*http.Response, error) {
	return call("PUT", format, args...)
}

func del(format string, args ...interface{}) (*http.Response, error) {
	return call("DELETE", format, args...)
}

func get(format string, args ...interface{}) (*http.Response, error) {
	return call("GET", format, args...)
}

func patch(format string, args ...interface{}) (*http.Response, error) {
	return call("PATCH", format, args...)
}
