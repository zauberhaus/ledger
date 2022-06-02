package service

import (
	"net/http"

	"github.com/ec-systems/core.ledger.service/pkg/config"
	"github.com/ec-systems/core.ledger.service/pkg/metrics"
)

type ServiceOption interface {
	Set(*MTlsService)
}

type ServiceOptionFunc func(*MTlsService)

func (f ServiceOptionFunc) Set(c *MTlsService) {
	f(c)
}

func Device(dev string) ServiceOption {
	return ServiceOptionFunc(func(c *MTlsService) {
		c.dev = dev
	})
}

func Port(port int) ServiceOption {
	return ServiceOptionFunc(func(c *MTlsService) {
		c.port = port
	})
}

func Metrics(port int, name string, buckets ...float64) ServiceOption {
	return ServiceOptionFunc(func(c *MTlsService) {
		if port > 0 {
			c.router.Use(metrics.NewMiddleware(name, buckets...))
			c.metrics = port
		}
	})
}

func MTls(options *config.MTLsOptions) ServiceOption {
	return ServiceOptionFunc(func(c *MTlsService) {
		c.mtls = options
	})
}

func Use(middlewares ...func(http.Handler) http.Handler) ServiceOption {
	return ServiceOptionFunc(func(c *MTlsService) {
		c.router.Use(middlewares...)
	})
}

func Mount(path string, handler http.Handler) ServiceOption {
	return ServiceOptionFunc(func(c *MTlsService) {
		c.router.Mount(path, handler)
	})
}

func Get(routes map[string]http.HandlerFunc) ServiceOption {
	return ServiceOptionFunc(func(c *MTlsService) {
		for k, v := range routes {
			c.router.Get(k, v)
		}
	})
}
