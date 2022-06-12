package cmd

import (
	"fmt"
	"strings"

	"github.com/ec-systems/core.ledger.server/pkg/client"
	"github.com/ec-systems/core.ledger.server/pkg/config"
	"github.com/ec-systems/core.ledger.server/pkg/ledger"
	"github.com/ec-systems/core.ledger.server/pkg/types"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func addAssetsCmd(root *RootCommand) {
	cmd := &cobra.Command{
		Use:   "assets [asset prefix]",
		Short: "Show assets",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Configuration()

			client, err := client.New(cmd.Context(), cfg.ClientOptions.Username, cfg.ClientOptions.Password, cfg.ClientOptions.Database,
				client.ClientOptions(cfg.ClientOptions),
				client.Limit(25),
			)
			if err != nil {
				return fmt.Errorf("database client error: %v", err)
			}

			defer client.Close(cmd.Context())

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.SupportedStatuses(cfg.Statuses),
			)

			showSupported, err := cmd.Flags().GetBool("supported")
			if err != nil {
				return err
			}

			if showSupported {
				prefix := ""
				if len(args) > 0 {
					prefix = strings.ToUpper(args[0])
				}

				table := tablewriter.NewWriter(cmd.OutOrStderr())
				table.SetHeader([]string{"Symbol", "Asset"})
				table.SetAlignment(tablewriter.ALIGN_LEFT)

				assets := cfg.Assets

				if prefix == "" {
					for k, v := range assets {
						table.Append([]string{k.String(), v})
					}
				} else {
					for k, v := range assets {
						if strings.HasPrefix(k.String(), prefix) {
							table.Append([]string{k.String(), v})
						}
					}
				}

				table.Render()
			} else {
				showBalance, err := cmd.Flags().GetBool("balance")
				if err != nil {
					return err
				}

				if showBalance {
					asset := types.AllAssets
					if len(args) > 0 {
						asset, err = cfg.Assets.Parse(args[0])
						if err != nil {
							return err
						}
					}

					balances, err := l.AssetBalance(cmd.Context(), asset)
					if err != nil {
						return fmt.Errorf("failed to load asset balances: %v", err)
					}

					table := tablewriter.NewWriter(cmd.OutOrStderr())
					table.SetHeader([]string{"Symbol", "Asset", "Balance"})

					for k, v := range balances {
						table.Append([]string{k.String(), cfg.Assets.Name(k), v.String()})
					}

					table.Render()
				} else {
					table := tablewriter.NewWriter(cmd.OutOrStderr())
					table.SetHeader([]string{"Symbol", "Asset"})

					assets, err := l.Assets(cmd.Context())
					if err != nil {
						return err
					}

					for _, asset := range assets {
						table.Append([]string{asset.String(), cfg.Assets.Name(asset)})
					}

					table.Render()
				}
			}

			return nil
		},
	}

	cmd.Flags().Bool("supported", false, "Show supported assets")
	cmd.Flags().Bool("balance", false, "Show balance of used assets")

	root.AddCommand(cmd)
}
