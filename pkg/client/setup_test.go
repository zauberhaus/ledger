package client_test

import (
	"context"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	immudb "github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/stream"
	"github.com/goombaio/namegenerator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ec-systems/core.ledger.server/pkg/client"
)

const (
	CLIENT_OPTIONS_ADDRESS                  = "localhost"
	CLIENT_OPTIONS_PORT                     = 3322
	CLIENT_OPTIONS_USERNAME                 = "immudb"
	CLIENT_OPTIONS_PASSWORD                 = "immudb"
	CLIENT_OPTIONS_MTLS                     = false
	CLIENT_OPTIONS_DATABASE                 = "testclient"
	CLIENT_OPTIONS_MTLS_OPTIONS_CERTIFICATE = "../../certs/tls.crt"
	CLIENT_OPTIONS_MTLS_OPTIONS_CLIENT_CAS  = "../../certs/ca.crt"
	CLIENT_OPTIONS_MTLS_OPTIONS_PKEY        = "../../certs/tls.key"
	CLIENT_OPTIONS_MTLS_OPTIONS_SERVERNAME  = "ledger-immudb-primary"
	CLIENT_OPTIONS_TOKEN_FILE_NAME          = "./token"
)

var (
	cfg = &immudb.Options{
		Dir:                "./testdata",
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
		Config:             "configs/immuclient.toml",
		DialOptions:        []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},

		//MTLsOptions: immudb.MTLsOptions{
		//	Certificate: CLIENT_OPTIONS_MTLS_OPTIONS_CERTIFICATE,
		//	ClientCAs:   CLIENT_OPTIONS_MTLS_OPTIONS_CLIENT_CAS,
		//	Pkey:        CLIENT_OPTIONS_MTLS_OPTIONS_PKEY,
		//	Servername:  CLIENT_OPTIONS_MTLS_OPTIONS_SERVERNAME,
		//},
	}
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UTC().UnixNano())

	ctx := context.Background()
	var cl *client.Client
	var err error

	cl, err = client.New(ctx, cfg.Username, cfg.Password, "defaultdb",
		client.ClientOptions(cfg),
		client.Limit(5),
	)

	if err != nil {
		log.Fatal(err)
	}

	exists, err := cl.DatabaseExist(ctx, cfg.Database)
	if err != nil {
		log.Fatal(err)
	}

	if exists {
		log.Printf("Delete test database: %v", cfg.Database)

		err := cl.UnloadDatabase(ctx, cfg.Database)
		if err != nil {
			log.Fatal(err)
		}

		err = cl.DeleteDatabase(ctx, cfg.Database)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Create test database: %v", cfg.Database)

	err = cl.CreateDatabase(ctx, cfg.Database)
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
