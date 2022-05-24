package main_test

import (
	"testing"

	"github.com/ec-systems/core.ledger.tool/cmd"
	generator "github.com/ec-systems/core.ledger.tool/pkg/generator/config"
	"github.com/stretchr/testify/assert"
)

type BindingTestConfig struct {
	Integer int    `env:"INTEGER" default:"123456"`
	String  string `env:"STRING" default:"test"`
	Bool    bool   `env:"BOOL" default:"true"`
}

func TestGroupBindings(t *testing.T) {

	rootCmd := cmd.GetRootCmd(&cmd.Version{})

	b := rootCmd.EnvBindings()
	assert.Len(t, b, 3)

	r, err := generator.GroupBindings(b)
	assert.NoError(t, err)
	assert.Len(t, r, 3)

	assert.Contains(t, r[0], "bool")
	assert.Len(t, r[0], 1)
	assert.Equal(t, r[0]["bool"], "BOOL")

	assert.Contains(t, r[1], "integer")
	assert.Len(t, r[1], 1)
	assert.Equal(t, r[1]["integer"], "INTEGER")

	assert.Contains(t, r[2], "string")
	assert.Len(t, r[2], 1)
	assert.Equal(t, r[2]["string"], "STRING")

}
