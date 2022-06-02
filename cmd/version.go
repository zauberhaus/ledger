package cmd

import (
	"fmt"
	"runtime"

	"github.com/ec-systems/core.ledger.service/pkg/logger"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Version struct {
	BuildDate    string `yaml:"buildDate,omitempty" json:"buildDate,omitempty" xml:"buildDate,omitempty"`
	Compiler     string `yaml:"compiler" json:"compiler" xml:"compiler"`
	GitCommit    string `yaml:"gitCommit,omitempty" json:"gitCommit,omitempty" xml:"gitCommit,omitempty"`
	GitTreeState string `yaml:"gitTreeState,omitempty" json:"gitTreeState,omitempty" xml:"gitTreeState,omitempty"`
	GitVersion   string `yaml:"gitVersion,omitempty" json:"gitVersion,omitempty" xml:"gitVersion,omitempty"`
	GoVersion    string `yaml:"goVersion" json:"goVersion" xml:"goVersion"`
	Platform     string `yaml:"platform" json:"platform" xml:"platform"`
}

func (v *Version) String() string {
	data, _ := yaml.Marshal(v)
	return string(data)
}

// NewVersion creates a new version object
func NewVersion(buildDate string, gitCommit string, tag string, treeState string) *Version {
	return &Version{
		BuildDate:    buildDate,
		Compiler:     runtime.Compiler,
		GitCommit:    gitCommit,
		GitTreeState: treeState,
		GitVersion:   tag,
		GoVersion:    runtime.Version(),
		Platform:     fmt.Sprintf("%v/%v", runtime.GOOS, runtime.GOARCH),
	}
}

// addVersionCmd creates and adds the version command to Root
func addVersionCmd(root *RootCommand) {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the version info",
		Run: func(cmd *cobra.Command, args []string) {

			data, err := yaml.Marshal(root.GetVersion())
			if err != nil {
				logger.Error("Invalid version: %v", err)
			}

			fmt.Fprintln(cmd.OutOrStderr(), string(data))
		},
	}

	root.AddCommand(versionCmd)
}
