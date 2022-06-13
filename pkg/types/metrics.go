package types

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shopspring/decimal"
)

type MetricsCollector interface {
	prometheus.Collector

	Add(Asset, decimal.Decimal)
}
