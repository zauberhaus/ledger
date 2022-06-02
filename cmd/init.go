package cmd

import (
	"log"
	"strings"

	"github.com/ec-systems/core.ledger.service/pkg/client"
	"github.com/ec-systems/core.ledger.service/pkg/config"

	"fmt"

	"github.com/spf13/cobra"

	"github.com/go-playground/validator/v10"
)

func addInitCmd(root *RootCommand) {

	cmd := &cobra.Command{
		Use:           "init",
		Short:         "Creates the database if not exists",
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

			database := cfg.ClientOptions.Database

			cl, err := client.New(cmd.Context(), cfg.ClientOptions.Username, cfg.ClientOptions.Password, "defaultdb",
				client.ClientOptions(cfg.ClientOptions),
				client.Limit(5),
			)

			if err != nil {
				log.Fatal(err)
			}

			exists, err := cl.DatabaseExist(cmd.Context(), database)
			if err != nil {
				log.Fatal(err)
			}

			if !exists {

				log.Printf("Create test database: %v", database)

				err = cl.CreateDatabase(cmd.Context(), database)
				if err != nil {
					log.Fatal(err)
				}

			}
			return nil
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.Flags().StringP("account", "a", "", "Account id (optional)")

	root.AddCommand(cmd)
}
