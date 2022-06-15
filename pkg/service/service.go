package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/ec-systems/core.ledger.server/pkg/config"
	"github.com/ec-systems/core.ledger.server/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-chi/chi/v5"
)

type MTlsService struct {
	dev         string
	port        int
	metricsPort int
	mtls        *config.MTLsOptions

	done chan bool

	server  *http.Server
	mserver *http.Server

	router  *chi.Mux
	metrics *chi.Mux
}

func NewMTlsService(options ...ServiceOption) *MTlsService {
	svc := &MTlsService{
		router:  chi.NewRouter(),
		metrics: chi.NewRouter(),

		done: make(chan bool, 1),
	}

	for _, o := range options {
		if o != nil {
			o.Set(svc)
		}
	}

	return svc
}

func (l *MTlsService) Start() chan error {
	var wg sync.WaitGroup

	addr := fmt.Sprintf("%v:%v", l.dev, l.port)
	errChan := make(chan error, 1)

	wg.Add(1)
	go func() {
		if l.mtls != nil {
			tlsConfig, err := l.getTLSConfig(l.mtls.ClientCAs)
			if err != nil {
				errChan <- err
				wg.Done()
				return
			}

			server := &http.Server{
				Addr:      addr,
				TLSConfig: tlsConfig,
				Handler:   l.router,
			}

			logger.Infof("Ledger TLS service listen on %v:%v", l.dev, l.port)
			errChan <- server.ListenAndServeTLS(l.mtls.Certificate, l.mtls.Pkey)
			close(errChan)
		} else {
			l.server = &http.Server{
				Addr:    addr,
				Handler: l.router,
			}

			logger.Infof("Ledger service listen on %v:%v", l.dev, l.port)
			if err := l.server.ListenAndServe(); err != http.ErrServerClosed {
				errChan <- fmt.Errorf("failed to start service listener: %v", err)
			} else {
				logger.Info("Service listener stopped")
			}

		}

		wg.Done()
	}()

	if l.metricsPort > 0 {
		wg.Add(1)
		go func() {
			addr := fmt.Sprintf("%v:%v", l.dev, l.metricsPort)

			l.metrics.Handle("/metrics", promhttp.Handler())

			l.mserver = &http.Server{
				Addr:    addr,
				Handler: l.metrics,
			}

			logger.Infof("Metrics service listen on %v:%v", l.dev, l.metricsPort)
			if err := l.mserver.ListenAndServe(); err != http.ErrServerClosed {
				errChan <- fmt.Errorf("failed to start metrics listener: %v", err)
			} else {
				logger.Info("Metrics listener stopped")
			}

			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	return errChan
}

func (l *MTlsService) Done() chan bool {
	return l.done
}

func (l *MTlsService) Stop(ctx context.Context) error {
	var wg sync.WaitGroup

	var err [2]error

	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if l.mserver != nil {
		go func() {
			wg.Add(1)
			err[0] = l.mserver.Shutdown(timeout)
			wg.Done()
		}()
	}

	if l.server != nil {
		go func() {
			wg.Add(1)
			err[1] = l.server.Shutdown(timeout)
			wg.Done()
		}()
	}

	wg.Wait()

	close(l.done)

	if err[1] != nil {
		return err[1]
	}

	return err[0]
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
