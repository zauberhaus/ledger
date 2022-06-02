package ledger_test

import (
	"context"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/ec-systems/core.ledger.service/pkg/client"
)

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

	exists, err := cl.DatabaseExist(ctx, CLIENT_OPTIONS_DATABASE)
	if err != nil {
		log.Fatal(err)
	}

	if exists {
		log.Printf("Delete test database: %v", CLIENT_OPTIONS_DATABASE)

		err := cl.UnloadDatabase(ctx, CLIENT_OPTIONS_DATABASE)
		if err != nil {
			log.Fatal(err)
		}

		err = cl.DeleteDatabase(ctx, CLIENT_OPTIONS_DATABASE)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Create test database: %v", CLIENT_OPTIONS_DATABASE)

	err = cl.CreateDatabase(ctx, CLIENT_OPTIONS_DATABASE)
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	cl.Close(ctx)

	os.Exit(code)
}
