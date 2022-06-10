package cmd

import (
	"os"
	"strconv"
	"strings"

	"github.com/ec-systems/core.ledger.service/pkg/client"
	"github.com/ec-systems/core.ledger.service/pkg/config"
	"github.com/olekukonko/tablewriter"

	"fmt"

	"github.com/spf13/cobra"

	"github.com/go-playground/validator/v10"
)

func addKeysCmd(root *RootCommand) {

	cmd := &cobra.Command{
		Use:           "keys <immudb tx id>",
		Short:         "Show keys of a immudb transaction",
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

			txID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			client, err := client.New(cmd.Context(), cfg.ClientOptions.Username, cfg.ClientOptions.Password, cfg.ClientOptions.Database,
				client.ClientOptions(cfg.ClientOptions),
				client.Limit(25),
			)
			if err != nil {
				return fmt.Errorf("immudb client error: %v", err)
			}

			defer client.Close(cmd.Context())

			tx, err := client.GetTx(cmd.Context(), txID)
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Type", "ID", "Hash"})

			for i, e := range tx.Entries {
				if e.Key[0] != 0 {
					if i == 0 {
						table.Append([]string{"Key", fmt.Sprintf("%v", i), string(e.Key)})
					} else {
						table.Append([]string{"Ref", fmt.Sprintf("%v", i), string(e.Key)})
					}
				} else {
					l := uint8(e.Key[7])
					key := string(e.Key[8 : 8+l])
					table.Append([]string{"Set", fmt.Sprintf("%v", i), key})
				}
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
