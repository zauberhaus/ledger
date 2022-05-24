package config_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ec-systems/core.ledger.tool/pkg/config"
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type TestConfig struct {
	Interval config.Duration
}

func TestDuration_JSON(t *testing.T) {
	wanted := &TestConfig{
		Interval: config.Duration(78 * time.Minute),
	}

	actual := &TestConfig{}

	data, err := json.Marshal(wanted)
	assert.NoError(t, err)

	err = json.Unmarshal(data, actual)
	assert.NoError(t, err)

	assert.Equal(t, wanted, actual)
}

func TestDuration_YAML(t *testing.T) {
	wanted := &TestConfig{
		Interval: config.Duration(78 * time.Minute),
	}

	actual := &TestConfig{}

	data, err := yaml.Marshal(wanted)
	assert.NoError(t, err)

	err = yaml.Unmarshal(data, actual)
	assert.NoError(t, err)

	assert.Equal(t, wanted, actual)
}

func TestDuration_TOML(t *testing.T) {
	wanted := &TestConfig{
		Interval: config.Duration(78 * time.Minute),
	}

	actual := &TestConfig{}

	data, err := toml.Marshal(wanted)
	assert.NoError(t, err)

	err = toml.Unmarshal(data, actual)
	assert.NoError(t, err)

	assert.Equal(t, wanted, actual)
}
