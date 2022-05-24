package cmd

import (
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

func addAccountsCmd(root *RootCommand) {
	cmd := &cobra.Command{
		Use:           "accounts",
		Short:         "List all accounts of a customer",
		Args:          cobra.RangeArgs(1, 2),
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

			l := ledger.New(client, ledger.SupportedAssets(cfg.Assets))

			customer := args[0]

			var asset types.Asset
			if len(args) > 1 {
				asset, err = l.ParseAsset(args[1])
				if err != nil {
					return err
				}
			}

			balances, err := l.Balance(cmd.Context(), customer, asset, types.AllAccounts, types.AllStatuses)
			if err != nil {
				return err
			}

			if len(balances) == 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), "No accounts for customer %v not found\n", customer)
				return nil
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Customer", "Account", "Asset", "TX Count", "Balance"})

			for asset, balance := range balances {
				for account, acc := range balance.Accounts {
					table.Append([]string{customer, account.String(), asset.String(), fmt.Sprintf("%v", acc.Count), fmt.Sprintf("%.8f", acc.Sum)})
				}
			}

			table.Render()

			return nil
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	//cfg := config.Configuration()

	//root.bindCmdFlag(cmd.Flags(), "Worker.TaskQueueName", "queue")

	root.AddCommand(cmd)
}
