package index_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ec-systems/core.ledger.service/pkg/ledger/index"
	"github.com/ec-systems/core.ledger.service/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestAssetIndex_Key(t *testing.T) {
	asset := types.XRP
	id, err := types.NewRandomID()

	if assert.NoError(t, err) {
		key := index.AssetTx.Key(asset, "c001", "a001", id)
		assert.False(t, strings.HasSuffix(string(key), ":"))
		assert.Equal(t, fmt.Sprintf("AT:XRP:c001:a001:%v", id.HexString()), string(key))
	}
}
