package cmd

import (
	"context"
	"os"
	"strings"

	"github.com/ec-systems/core.ledger.tool/pkg/client"
	"github.com/ec-systems/core.ledger.tool/pkg/config"
	"github.com/ec-systems/core.ledger.tool/pkg/ledger"
	"github.com/olekukonko/tablewriter"

	"fmt"

	"github.com/spf13/cobra"

	"github.com/go-playground/validator/v10"
)

func addHistoryCmd(root *RootCommand) {

	cmd := &cobra.Command{
		Use:           "history",
		Short:         "Show the history of a immudb key",
		Args:          cobra.ExactArgs(1),
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

			key := args[0]

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

			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"TX", "Date", "Status"})

			err = l.History(cmd.Context(), key, func(ctx context.Context, tx *ledger.Transaction) (bool, error) {
				table.Append(tx.Change())
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
}
