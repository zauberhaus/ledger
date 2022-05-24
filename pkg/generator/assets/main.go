package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"path"
	"strings"

	"github.com/ec-systems/core.ledger.tool/pkg/logger"
	"gopkg.in/yaml.v3"
)

var (
	filename  = "assets.go"
	doNotEdit = "//Code generated by assets generator. DO NOT EDIT.\n"
)

func main() {
	yamlFile, err := ioutil.ReadFile(path.Join("..", "..", "assets.yaml"))
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	var assets map[string]string

	err = yaml.Unmarshal(yamlFile, &assets)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	var sb strings.Builder

	sb.WriteString(doNotEdit)
	sb.WriteString("package types\n\n")

	sb.WriteString("const (\n")
	for k, v := range assets {
		sb.WriteString(fmt.Sprintf("\t%v Asset = \"%v\"\n", k, v))
	}
	sb.WriteString(")\n\n")

	sb.WriteString("var DefaultAssetNames = Assets{\n")
	for k, v := range assets {
		sb.WriteString(fmt.Sprintf("\t\"%v\": \"%v\",\n", k, v))
	}

	sb.WriteString("}\n\n")

	sb.WriteString("var DefaultAssetMap = Assets{\n")
	for k, v := range assets {
		sb.WriteString(fmt.Sprintf("\t\"%v\": %v,\n", v, k))
	}

	sb.WriteString("}\n")

	logger.Infof("Write assets file: %v", filename)
	err = ioutil.WriteFile(filename, []byte(sb.String()), fs.ModePerm)
	if err != nil {
		logger.Fatalf("Error write %v: %v", filename, err)
	}

}