package cmd

import (
	"os"
	"strings"

	"github.com/ec-systems/core.ledger.service/pkg/client"
	"github.com/ec-systems/core.ledger.service/pkg/config"
	"github.com/ec-systems/core.ledger.service/pkg/ledger"
	"github.com/ec-systems/core.ledger.service/pkg/types"
	"github.com/olekukonko/tablewriter"

	"fmt"

	"github.com/spf13/cobra"

	"github.com/go-playground/validator/v10"
)

func addAccountsCmd(root *RootCommand) {
	cmd := &cobra.Command{
		Use:           "accounts <holder id> [asset]",
		Short:         "List all accounts of a holder",
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

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.SupportedStatuses(cfg.Statuses),
			)

			holder := args[0]

			var asset types.Asset
			if len(args) > 1 {
				asset, err = cfg.Assets.Parse(args[0])
				if err != nil {
					return err
				}
			}

			balances, err := l.Balance(cmd.Context(), holder, asset, types.AllAccounts, types.AllStatuses)
			if err != nil {
				return err
			}

			if len(balances) == 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), "No accounts for holder %v not found\n", holder)
				return nil
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Holder", "Account", "Asset", "TX Count", "Balance"})

			for asset, balance := range balances {
				for account, acc := range balance.Accounts {
					table.Append([]string{holder, account.String(), asset.String(), fmt.Sprintf("%v", acc.Count), acc.Sum.String()})
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
