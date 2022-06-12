package cmd

import (
	"strings"

	"github.com/ec-systems/core.ledger.server/docs"
	"github.com/ec-systems/core.ledger.server/pkg/client"
	"github.com/ec-systems/core.ledger.server/pkg/config"
	"github.com/ec-systems/core.ledger.server/pkg/ledger"
	"github.com/ec-systems/core.ledger.server/pkg/logger"
	"github.com/ec-systems/core.ledger.server/pkg/service"

	"fmt"

	"github.com/spf13/cobra"

	"github.com/go-playground/validator/v10"
)

func addServiceCmd(root *RootCommand) {

	cmd := &cobra.Command{
		Use:           "service",
		Short:         "Starts ledger web service",
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

			if cfg.Service.MTls != nil {
				docs.SwaggerInfo.Schemes = []string{"https"}
			} else {
				docs.SwaggerInfo.Schemes = []string{"http"}
			}

			if cfg.Service.Servername != "" {
				docs.SwaggerInfo.Host = fmt.Sprintf("%v:%d", cfg.Service.Servername, cfg.Service.Port)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Configuration()

			logger.Infof("Start ledger service %v", root.GetVersion().GitVersion)
			logger.Infof("Configuration\n%v", cfg)

			client, err := client.New(cmd.Context(), cfg.ClientOptions.Username, cfg.ClientOptions.Password, cfg.ClientOptions.Database,
				client.ClientOptions(cfg.ClientOptions),
				client.Limit(25),
			)
			if err != nil {
				return fmt.Errorf("database client error: %v", err)
			}

			defer client.Close(cmd.Context())

			l := ledger.New(client,
				ledger.SupportedAssets(cfg.Assets),
				ledger.SupportedStatuses(cfg.Statuses),
			)

			svc, err := service.NewLedgerService(cmd.Context(), l, &cfg.Service)
			if err != nil {
				return fmt.Errorf("service error: %v", err)
			}

			return svc.Start()
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cfg := config.Configuration()

	cmd.Flags().StringP("interface", "i", cfg.Service.Device, "Network device for listener")
	root.bindFlags(cmd.Flags(), "Service.Device", "interface")

	cmd.Flags().String("servername", cfg.Service.Servername, "Published server name")
	root.bindFlags(cmd.Flags(), "Service.Servername", "servername")

	cmd.Flags().IntP("port", "p", cfg.Service.Port, "Service port")
	root.bindFlags(cmd.Flags(), "Service.Port", "port")

	cmd.Flags().IntP("metrics", "M", cfg.Service.Metrics, "Metrics port")
	root.bindFlags(cmd.Flags(), "Service.Metrics", "metrics")

	cmd.Flags().Bool("production", cfg.Service.Production, "Service port")
	root.bindFlags(cmd.Flags(), "Service.Production", "production")

	cmd.Flags().Bool("access-log", cfg.Service.AccessLogger, "Enabled access logger")
	root.bindFlags(cmd.Flags(), "Service.AccessLogger", "access-log")

	root.AddCommand(cmd)
}
