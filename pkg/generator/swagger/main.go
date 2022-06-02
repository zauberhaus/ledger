package main

import (
	"github.com/ec-systems/core.ledger.service/pkg/logger"
	"github.com/swaggo/swag"
	"github.com/swaggo/swag/gen"
)

func main() {

	logger.Info("Create swagger documentation")

	strategy := swag.CamelCase

	gen.New().Build(&gen.Config{
		MainAPIFile:         "pkg/service/ledger.go",
		SearchDir:           "../../.",
		Excludes:            "",
		PropNamingStrategy:  strategy,
		OutputDir:           "../../docs/",
		OutputTypes:         []string{"go", "json", "yaml"},
		ParseVendor:         false,
		ParseDependency:     false,
		MarkdownFilesDir:    "",
		ParseInternal:       false,
		GeneratedTime:       false,
		CodeExampleFilesDir: "",
		ParseDepth:          100,
		InstanceName:        "",
		OverridesFile:       ".swaggo",
	})
}
