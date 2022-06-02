package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/ec-systems/core.ledger.service/pkg/logger"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

func Marshal(cfg interface{}, file string) ([]byte, error) {

	ext := filepath.Ext(file)[1:]

	switch ext {
	case "yml", "yaml":
		var b bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&b)
		yamlEncoder.SetIndent(2)

		err := yamlEncoder.Encode(cfg)
		if err != nil {
			logger.Fatal(err)
		}

		return b.Bytes(), err

	case "json":
		return json.MarshalIndent(cfg, "", "  ")
	case "toml":
		return toml.Marshal(cfg)
	default:
		return nil, fmt.Errorf("unknown file extension")
	}
}
