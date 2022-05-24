package cmd

import (
	"github.com/ec-systems/core.ledger.tool/pkg/config"
	"github.com/ec-systems/core.ledger.tool/pkg/types"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// addVersionCmd creates and adds the version command to Root
func addAssetsCmd(root *RootCommand) {
	var versionCmd = &cobra.Command{
		Use:   "assets",
		Short: "Supported assets",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Configuration()

			table := tablewriter.NewWriter(cmd.OutOrStderr())
			table.SetHeader([]string{"Asset", "Symbol"})

			assets := cfg.Assets

			if assets == nil || len(assets) == 0 {
				assets = types.DefaultAssetNames
			}

			for k, v := range assets {
				table.Append([]string{k, string(v)})
			}

			table.Render()

		},
	}

	root.AddCommand(versionCmd)
}
