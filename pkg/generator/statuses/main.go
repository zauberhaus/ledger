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
	filename  = "./statuses.go"
	doNotEdit = "//Code generated by statuses generator. DO NOT EDIT.\n"
)

func main() {
	yamlFile, err := ioutil.ReadFile(path.Join("..", "..", "statuses.yaml"))
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	var status map[string]int

	err = yaml.Unmarshal(yamlFile, &status)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	status["Unknown"] = -1
	status["Created"] = 0

	var sb strings.Builder

	sb.WriteString(doNotEdit)
	sb.WriteString("package types\n\n")

	sb.WriteString("const (\n")

	for k, v := range status {
		sb.WriteString(fmt.Sprintf("\t%v Status = %v\n", k, v))
	}

	sb.WriteString(")\n\n")

	sb.WriteString("var DefaultStatusMap = Statuses{\n")
	for k := range status {
		sb.WriteString(fmt.Sprintf("\t\"%v\": %v,\n", k, k))
	}

	// sb.WriteString("}\n\n")

	// sb.WriteString("var reverseStatusMap = map[Status]string{\n")
	// for k := range status {
	// 	sb.WriteString(fmt.Sprintf("\t%v: \"%v\",\n", k, k))
	// }

	sb.WriteString("}\n")

	logger.Infof("Write status file: %v", filename)
	err = ioutil.WriteFile(filename, []byte(sb.String()), fs.ModePerm)
	if err != nil {
		logger.Fatalf("Error write %v: %v", filename, err)
	}

}
