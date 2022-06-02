package main

import (
	"fmt"
	"os"

	"github.com/ec-systems/core.ledger.service/cmd"
)

var (
	tag       string // git tag used to build the program
	gitCommit string // sha1 revision used to build the program
	buildTime string // when the executable was built
	treeState string // git tree state
)

func main() {
	version := cmd.NewVersion(buildTime, gitCommit, tag, treeState)

	if err := cmd.GetRootCmd(version).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
