package cmd

import (
	"os"
	"strings"

	"github.com/ec-systems/core.ledger.server/pkg/client"
	"github.com/ec-systems/core.ledger.server/pkg/config"
	"github.com/ec-systems/core.ledger.server/pkg/ledger"
	"github.com/ec-systems/core.ledger.server/pkg/types"
	"github.com/olekukonko/tablewriter"

	"fmt"

	"github.com/spf13/cobra"

	"github.com/go-playground/validator/v10"
)

func addHoldersCmd(root *RootCommand) {

	cmd := &cobra.Command{
		Use:           "holders",
		Short:         "List all account holders",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Configuration()
			validate := validator.New()

			err := validate.Struct(cfg)
			switch v := err.(type) {
			case validator.ValidationErrors:
				messages := []string{}
				for _, err := range v {
					msg := fmt.Sprintf("%v is %v", err.StructNamespace(), err.ActualTag())
					messages = append(messages, msg)
				}

				return fmt.Errorf("invalid configuration: %v", strings.Join(messages, ", "))
			case *validator.InvalidValidationError:
				return fmt.Errorf("invalid configuration: %v", v)
			default:
				if err != nil {
					return err
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Configuration()

			client, err := client.New(cmd.Context(), cfg.ClientOptions.Username, cfg.ClientOptions.Password, cfg.ClientOptions.Database,
				client.ClientOptions(cfg.ClientOptions),
				client.Limit(25),
			)
			if err != nil {
				return fmt.Errorf("immudb client error: %v", err)
			}

			defer client.Close(cmd.Context())

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.SupportedStatuses(cfg.Statuses),
			)

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Holder", "Account", "Asset"})
			table.SetAutoMergeCells(true)

			err = l.Holders(cmd.Context(), func(holder string, account types.Account, asset types.Asset) (bool, error) {
				table.Append([]string{holder, account.String(), asset.String()})
				return true, nil
			})
			if err != nil {
				return err
			}

			table.SetFooterAlignment(tablewriter.ALIGN_RIGHT)
			table.Render()

			return nil
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	root.AddCommand(cmd)
}
