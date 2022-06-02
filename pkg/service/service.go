package service

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ec-systems/core.ledger.service/pkg/config"
	"github.com/ec-systems/core.ledger.service/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-chi/chi/v5"
)

type MTlsService struct {
	dev     string
	port    int
	metrics int
	mtls    *config.MTLsOptions

	router *chi.Mux
}

func NewMTlsService(options ...ServiceOption) *MTlsService {
	svc := &MTlsService{
		router: chi.NewRouter(),
	}

	for _, o := range options {
		if o != nil {
			o.Set(svc)
		}
	}

	return svc
}

func (l *MTlsService) Start() error {
	addr := fmt.Sprintf("%v:%v", l.dev, l.port)

	if l.mtls != nil {
		tlsConfig, err := l.getTLSConfig(l.mtls.ClientCAs)
		if err != nil {
			return err
		}

		server := &http.Server{
			Addr:      addr,
			TLSConfig: tlsConfig,
			Handler:   l.router,
		}

		logger.Infof("Ledger TLS service listen on %v:%v", l.dev, l.port)
		return server.ListenAndServeTLS(l.mtls.Certificate, l.mtls.Pkey)
	} else {
		server := &http.Server{
			Addr:    addr,
			Handler: l.router,
		}

		logger.Infof("Ledger service listen on %v:%v", l.dev, l.port)
		return server.ListenAndServe()
	}
}

func (l *MTlsService) StartMetrics() error {
	addr := fmt.Sprintf("%v:%v", l.dev, l.metrics)

	r := chi.NewRouter()

	r.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	logger.Infof("Metrics service listen on %v:%v", l.dev, l.metrics)
	return server.ListenAndServe()
}

func (s *MTlsService) Mount(path string, router chi.Router) {
	s.router.Mount(path, router)
}

func (s *MTlsService) getTLSConfig(ca string) (*tls.Config, error) {
	caCert, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, fmt.Errorf("failed to read mtls certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	return tlsConfig, nil
}
