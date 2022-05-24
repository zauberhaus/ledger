package cmd

import (
	"context"
	"os"
	"strings"

	"github.com/ec-systems/core.ledger.tool/pkg/client"
	"github.com/ec-systems/core.ledger.tool/pkg/config"
	"github.com/ec-systems/core.ledger.tool/pkg/ledger"
	"github.com/ec-systems/core.ledger.tool/pkg/types"
	"github.com/olekukonko/tablewriter"

	"fmt"

	"github.com/spf13/cobra"

	"github.com/go-playground/validator/v10"
)

func addTxCmd(root *RootCommand) {
	//var cmd *cobra.Command

	cmd := &cobra.Command{
		Use:           "tx",
		Short:         "List all transactions of a customer",
		Args:          cobra.RangeArgs(0, 3),
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
			customer := ""

			client, err := client.New(cmd.Context(), cfg.ClientOptions.Username, cfg.ClientOptions.Password, cfg.ClientOptions.Database,
				client.ClientOptions(cfg.ClientOptions),
				client.Limit(25),
			)
			if err != nil {
				return fmt.Errorf("immudb client error: %v", err)
			}

			defer client.Close(cmd.Context())

			statuses := cfg.Statuses
			assets := cfg.Assets
			l := ledger.New(client,
				ledger.SupportedAssets(assets),
				ledger.SupportedStatuses(statuses),
			)

			if len(args) > 0 {
				customer = args[0]
			}

			asset := types.AllAssets
			if len(args) > 1 {
				asset, err = assets.Parse(args[1])
				if err != nil {
					return err
				}
			}

			account := types.AllAccounts
			if len(args) > 2 {
				account = types.Account(args[2])
				if !account.Check() {
					return fmt.Errorf("invalid checksum for account '%v'", account)
				}
			}

			status := types.AllStatuses
			if len(args) > 3 {
				status, err = statuses.Parse(args[3])
				if err != nil {
					return err
				}
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"TX", "ID", "Date", "Updated", "Customer", "Account", "Order", "Item", "Asset", "Status", "Amount", "Ref"})

			err = l.Transactions(cmd.Context(), customer, asset, account, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
				if tx.Status > status {
					table.Append(tx.Row(false))
				}
				return true, nil
			})

			if err != nil {
				return err
			}

			table.Render()

			_ = l
			_ = asset
			_ = customer
			_ = status

			return nil
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	root.AddCommand(cmd)
}
