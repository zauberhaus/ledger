//go:generate go run github.com/ec-systems/core.ledger.service/pkg/generator/config/
package config

import (
	"regexp"

	"github.com/ec-systems/core.ledger.service/pkg/logger"
	"github.com/ec-systems/core.ledger.service/pkg/types"
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
	Service       ServiceConfig

	Assets   types.Assets
	Statuses types.Statuses

	BatchSize int          `default:"25"`
	Format    types.Format `default:"json"`
}

type ServiceConfig struct {
	Device       string
	Port         int `default:"8888"`
	Production   bool
	AccessLogger bool `default:"true"`
	Metrics      int  `default:"9094"`
	Servername   string

	MTls *MTLsOptions `json:",omitempty" yaml:",omitempty"`
}

type MTLsOptions immudb.MTLsOptions

func (c *Config) String() string {

	data, err := yaml.Marshal(c)
	if err != nil {
		return err.Error()
	}

	regex := regexp.MustCompile(`password:\s{0,1}.+`)
	data = regex.ReplaceAll(data, []byte("password: *****"))
	return string(data)

}
