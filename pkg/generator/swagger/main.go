package main

import (
	"os"
	"path"

	"github.com/ec-systems/core.ledger.service/pkg/logger"
	"github.com/swaggo/swag"
	"github.com/swaggo/swag/gen"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		logger.Fatalf("Getwd: %v", err)
	}

	logger.Info("Create swagger documentation")

	searchDir := path.Join(wd, "..", "..")
	outputDir := path.Join(wd, "..", "..", "docs")
	mainAPIFile := "pkg/service/ledger.go"

	logger.Infof("SearchDir: %v", searchDir)
	logger.Infof("OutputDir: %v", outputDir)
	logger.Infof("MainAPIFile: %v", mainAPIFile)

	strategy := swag.CamelCase

	err = gen.New().Build(&gen.Config{
		MainAPIFile:         mainAPIFile,
		SearchDir:           searchDir,
		Excludes:            "",
		PropNamingStrategy:  strategy,
		OutputDir:           outputDir,
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

	if err != nil {
		logger.Fatal(err)
	}
}
