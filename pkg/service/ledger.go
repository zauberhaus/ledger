package service

//go:generate go run github.com/ec-systems/core.ledger.server/pkg/generator/swagger/

import (
	"context"

	"github.com/ec-systems/core.ledger.server/pkg/config"
	"github.com/ec-systems/core.ledger.server/pkg/ledger"
	"github.com/ec-systems/core.ledger.server/pkg/logger"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/ec-systems/core.ledger.server/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

type LedgerService struct {
	svc *MTlsService
	cfg *config.ServiceConfig
}

// @title Core Ledger
// @description This is the web service of the core asset ledger.

// @contact.name Easy Crypto Core Team
// @contact.url http://easycrypto.ai
// @contact.email support@easycrypto.ai

// @BasePath /
func NewLedgerService(ctx context.Context, ledger *ledger.Ledger, cfg *config.ServiceConfig) (*LedgerService, error) {

	svc := &LedgerService{
		cfg: cfg,
	}

	var swagger ServiceOption
	if !cfg.Production {
		logger.Info("Enable swagger documentation: /swagger/index.html")
		swagger = Mount("/swagger", httpSwagger.WrapHandler)
	}

	var accessLogger ServiceOption
	if cfg.AccessLogger {
		logger.Info("Enable access logger")
		accessLogger = Use(middleware.Logger)
	}

	var redirect ServiceOption
	if swagger != nil {
		logger.Info("Enable redirect to swagger ui")
		redirect = Redirect("/", "/swagger/index.html")
	}

	svc.svc = NewMTlsService(
		Metrics(cfg.Metrics, "ledger"),
		accessLogger,
		Use(middleware.Recoverer),
		Device(cfg.Device),
		Port(cfg.Port),
		MTls((*config.MTLsOptions)(cfg.MTls)),
		Mount("/accounts", NewAccountsService(ledger)),
		Mount("/assets", NewAssetsService(ledger)),
		Mount("/info", NewInfoService(ledger)),
		Method("GET", NewHealthService(ledger)),
		MetricsMethod("GET", NewHealthService(ledger)),
		swagger,
		redirect,
	)

	return svc, nil
}

func (l *LedgerService) Start() error {
	if l.cfg.Metrics > 0 {
		go func() {
			l.svc.StartMetrics()
		}()
	}

	return l.svc.Start()
}
