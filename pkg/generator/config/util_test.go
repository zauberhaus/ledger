package main_test

import (
	"testing"

	"github.com/ec-systems/core.ledger.server/cmd"
	generator "github.com/ec-systems/core.ledger.server/pkg/generator/config"
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
	assert.Len(t, b, 40)

	r, err := generator.GroupBindings(b)
	assert.NoError(t, err)
	assert.Len(t, r, 9)

}
