package cmd

import (
	"strconv"
	"strings"

	"github.com/ec-systems/core.ledger.tool/pkg/client"
	"github.com/ec-systems/core.ledger.tool/pkg/config"
	"github.com/ec-systems/core.ledger.tool/pkg/ledger"
	"github.com/ec-systems/core.ledger.tool/pkg/logger"
	"github.com/ec-systems/core.ledger.tool/pkg/types"

	"fmt"

	"github.com/spf13/cobra"

	"github.com/go-playground/validator/v10"
)

func addRemoveCmd(root *RootCommand) {
	var cmd *cobra.Command

	cmd = &cobra.Command{
		Use:           "remove <customer> <asset> <amount> [order] [order item]",
		Short:         "Remove assets from the ledger",
		Args:          cobra.RangeArgs(3, 5),
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

			customer := args[0]

			amount, err := strconv.ParseFloat(args[2], 64)
			if err != nil {
				return fmt.Errorf("invalid amount format: %v", err)
			}

			order := ""
			if len(args) > 3 {
				order = args[3]
			}

			item := ""
			if len(args) > 4 {
				item = args[4]
			}

			accountID, err := cmd.Flags().GetString("account")
			if err != nil {
				return err
			}

			account := types.Account(accountID)
			if !account.Empty() && !account.Check() {
				return fmt.Errorf("invalid checksum for account %v", account)
			}

			client, err := client.New(cmd.Context(), cfg.ClientOptions.Username, cfg.ClientOptions.Password, cfg.ClientOptions.Database,
				client.ClientOptions(cfg.ClientOptions),
				client.Limit(25),
			)

			if err != nil {
				return fmt.Errorf("immudb client error: %v", err)
			}

			defer client.Close(cmd.Context())

			assets := cfg.Assets
			l := ledger.New(client, ledger.SupportedAssets(assets))

			asset, err := l.ParseAsset(args[1])
			if err != nil {
				return err
			}

			id, err := l.Remove(cmd.Context(), customer, asset, amount,
				ledger.Account(account),
				ledger.OrderID(order),
				ledger.OrderItemID(item),
			)

			if err != nil {
				return err
			}

			logger.Infof("Transaction created: %v", id)

			return nil
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.Flags().StringP("account", "a", "", "Customer account id (optional)")
	cmd.Flags().StringP("order", "o", "", "Order id (optional)")
	cmd.Flags().StringP("order-item", "i", "", "Order item id (optional)")

	root.AddCommand(cmd)
}
