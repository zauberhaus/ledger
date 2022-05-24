//go:generate go run github.com/ec-systems/core.ledger.tool/pkg/generator/config/
package config

import (
	"github.com/ec-systems/core.ledger.tool/pkg/logger"
	"github.com/ec-systems/core.ledger.tool/pkg/types"
	"gopkg.in/yaml.v3"

	immudb "github.com/codenotary/immudb/pkg/client"
)

var (
	config Config = Config{
		ClientOptions: immudb.DefaultOptions(),
	}
)

func Configuration() *Config {
	return &config
}

func Set(cfg *Config) {
	config = *cfg
}

type Config struct {
	LogLevel      logger.Level
	ClientOptions *immudb.Options

	Assets   types.Assets
	Statuses types.Statuses

	BatchSize int `default:"25"`
}

func (c *Config) String() string {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err.Error()
	} else {
		return string(data)
	}
}
