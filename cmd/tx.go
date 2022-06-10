package cmd

import (
	"context"
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

const (
	updatedCol = 1
	statusCol  = 2
	refCol     = 8
	keyCol     = 16
	orderCol   = 32
	itemCol    = 64
)

func addTxCmd(root *RootCommand) {

	cmd := &cobra.Command{
		Use:           "tx",
		Short:         "List all transactions <holder id> [asset] [account id]",
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
			holder := ""

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
				holder = args[0]
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

			ref, err := cmd.Flags().GetBool("ref")
			if err != nil {
				return err
			}

			updated, err := cmd.Flags().GetBool("change")
			if err != nil {
				return err
			}

			status, err := cmd.Flags().GetBool("status")
			if err != nil {
				return err
			}

			key, err := cmd.Flags().GetBool("key")
			if err != nil {
				return err
			}

			order, err := cmd.Flags().GetBool("order")
			if err != nil {
				return err
			}

			item, err := cmd.Flags().GetBool("item")
			if err != nil {
				return err
			}

			verify, err := cmd.Flags().GetBool("verify")
			if err != nil {
				return err
			}

			verify = verify && cfg.ClientOptions.ServerSigningPubKey != ""

			columns := []string{"TX", "ID", "Date", "Holder", "Account", "Asset", "Amount"}
			colFlag := uint8(0)

			table := tablewriter.NewWriter(os.Stdout)

			if updated {
				columns = append(columns, "Updated")
				colFlag |= updatedCol
			}

			if order {
				columns = append(columns, "Order")
				colFlag |= orderCol
			}

			if item {
				columns = append(columns, "Item")
				colFlag |= itemCol
			}

			if status {
				columns = append(columns, "Status")
				colFlag |= statusCol
			}

			if key {
				columns = append(columns, "Key")
				colFlag |= keyCol
			}

			if ref {
				columns = append(columns, "Ref")
				colFlag |= refCol
			}

			table.SetHeader(columns)

			err = l.Transactions(cmd.Context(), holder, asset, account, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
				if verify {
					t, err := l.Get(cmd.Context(), tx.ID)
					if err != nil {
						return false, err
					}

					_ = t
				}

				table.Append(row(tx, colFlag, l.SupportedStatus()))
				return true, nil
			})

			if err != nil {
				return err
			}

			table.Render()

			return nil
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	root.AddCommand(cmd)

	cmd.Flags().BoolP("ref", "r", false, "Show ref column")
	cmd.Flags().BoolP("change", "c", false, "Show updated column")
	cmd.Flags().BoolP("status", "s", false, "Show status column")
	cmd.Flags().BoolP("key", "k", false, "Show key column")
	cmd.Flags().BoolP("order", "o", false, "Show order column")
	cmd.Flags().BoolP("item", "i", false, "Show item column")

	cmd.Flags().BoolP("all", "a", false, "Show all columns")

	cmd.Flags().BoolP("verify", "V", false, "Verify all transactions")
}

func row(t *ledger.Transaction, cols uint8, statuses types.Statuses) []string {
	row := []string{}

	row = append(row, fmt.Sprintf("%v", t.TX()))
	row = append(row, t.ID.String())
	row = append(row, t.Created.Format(ledger.TimeFormat))
	row = append(row, t.Holder)
	row = append(row, string(t.Account))
	row = append(row, t.Asset.String())
	row = append(row, t.Amount.String())

	if isSet(cols, updatedCol) {
		row = append(row, t.Modified.Format(ledger.TimeFormat))
	}

	if isSet(cols, orderCol) {
		row = append(row, t.Order)
	}

	if isSet(cols, itemCol) {
		row = append(row, t.Item)
	}

	if isSet(cols, statusCol) {
		row = append(row, t.Status.String(statuses))
	}

	if isSet(cols, keyCol) {
		row = append(row, t.Key())
	}

	if isSet(cols, refCol) {
		row = append(row, string(t.Reference))
	}

	return row
}

func isSet(val uint8, col uint8) bool {
	return val&col == col
}
