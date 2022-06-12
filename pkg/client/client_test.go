package client_test

import (
	"context"
	"math/rand"
	"strings"
	"testing"

	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/ec-systems/core.ledger.server/pkg/client"
	"github.com/stretchr/testify/assert"
)

func Test_Get(t *testing.T) {
	ctx := context.Background()

	client, err := client.New(ctx, cfg.Username, cfg.Password, cfg.Database,
		client.ClientOptions(cfg),
		client.Limit(5),
	)

	if !assert.NoError(t, err) {
		return
	}

	defer client.Close(ctx)

	key := randomName()
	setName1 := randomName()
	setName2 := randomName()
	refKey1 := randomName()
	refKey2 := randomName()
	score1 := rand.Float64()
	score2 := rand.Float64()

	kv := &schema.Op_Kv{
		Kv: &schema.KeyValue{
			Key:   []byte(key),
			Value: []byte("791846h1j4f1jk4fhj"),
		},
	}

	set1 := &schema.Op_ZAdd{
		ZAdd: &schema.ZAddRequest{
			Key:      kv.Kv.Key,
			Set:      []byte(setName1),
			Score:    score1,
			BoundRef: false,
		},
	}

	set2 := &schema.Op_ZAdd{
		ZAdd: &schema.ZAddRequest{
			Key:      kv.Kv.Key,
			Set:      []byte(setName2),
			Score:    score2,
			BoundRef: false,
		},
	}

	ref1 := &schema.Op_Ref{
		Ref: &schema.ReferenceRequest{
			ReferencedKey: kv.Kv.Key,
			Key:           []byte(refKey1),
			BoundRef:      false,
		},
	}

	ref2 := &schema.Op_Ref{
		Ref: &schema.ReferenceRequest{
			ReferencedKey: kv.Kv.Key,
			Key:           []byte(refKey2),
			BoundRef:      false,
		},
	}

	txID, err := client.Exec(ctx, kv, set1, set2, ref1, ref2)
	if !assert.NoError(t, err) {
		return
	}

	tx, err := client.GetTx(ctx, txID)
	if !assert.NoError(t, err) {
		return
	}

	refs := []string{}
	sets := []string{}

	for _, e := range tx.Entries {
		if e.Key[0] != 0 {
			refs = append(refs, string(e.Key))
		} else {
			tmp := e.Key[8:]
			parts := strings.Split(string(tmp), "?")
			if len(parts) > 1 {
				sets = append(sets, parts[0])
			}
		}
	}

	assert.Len(t, refs, 3)
	assert.Contains(t, refs, key)
	assert.Contains(t, refs, refKey1)
	assert.Contains(t, refs, refKey2)

	assert.Len(t, sets, 2)
	assert.Contains(t, sets, setName1)
	assert.Contains(t, sets, setName2)
}
