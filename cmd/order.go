package cmd

import (
	"context"
	"os"
	"strings"

	"github.com/ec-systems/core.ledger.service/pkg/client"
	"github.com/ec-systems/core.ledger.service/pkg/config"
	"github.com/ec-systems/core.ledger.service/pkg/ledger"
	"github.com/olekukonko/tablewriter"

	"fmt"

	"github.com/spf13/cobra"

	"github.com/go-playground/validator/v10"
)

func addOrdersCmd(root *RootCommand) {

	cmd := &cobra.Command{
		Use:           "orders <holder> [order id] [item id]",
		Short:         "Show orders",
		Args:          cobra.RangeArgs(1, 3),
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

			holder := args[0]

			order := ""
			if len(args) > 1 {
				order = args[1]
			}

			item := ""
			if len(args) > 1 {
				order = args[1]
			}

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

			if len(args) == 1 {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"TX", "Date", "Order"})

				err = l.Orders(cmd.Context(), holder, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
					row := []string{
						fmt.Sprintf("%v", tx.TX()),
						tx.Created.Format(ledger.TimeFormat),
						tx.Order,
					}

					table.Append(row)
					return true, nil
				})
				if err != nil {
					return err
				}
				table.Render()

			} else {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"TX", "Date", "Order", "Item", "Asset", "Status", "Amount"})

				err = l.OrderItems(cmd.Context(), holder, order, item, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
					row := []string{
						fmt.Sprintf("%v", tx.TX()),
						tx.Created.Format(ledger.TimeFormat),
						tx.Order,
					}

					row = append(row, tx.Item)
					row = append(row, tx.Asset.String())
					row = append(row, tx.Status.String(l.SupportedStatus()))
					row = append(row, tx.Amount.String())

					table.Append(row)
					return true, nil
				})
				if err != nil {
					return err
				}
				table.Render()
			}

			return nil
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	root.AddCommand(cmd)
}
