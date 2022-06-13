package metrics

import (
	"os"

	"github.com/ec-systems/core.ledger.server/pkg/config"
	"github.com/ec-systems/core.ledger.server/pkg/logger"
	"github.com/ec-systems/core.ledger.server/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shopspring/decimal"
)

type TxCollector struct {
	addSum      *prometheus.CounterVec
	addCount    *prometheus.CounterVec
	removeSum   *prometheus.CounterVec
	removeCount *prometheus.CounterVec

	database string
	dbServer string

	ns string
	ip string
}

func NewTxCollector(cfg *config.Config) types.MetricsCollector {

	labels := []string{
		"asset",
		"host",
		"database",
		"immudb",
	}

	ns := os.Getenv("POD_NAMESPACE")
	if ns != "" {
		labels = append(labels, "namespace")
	}

	ip := os.Getenv("POD_IP")
	if ip != "" {
		labels = append(labels, "ip")
	}

	addSum := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "core",
			Subsystem: "ledger",
			Name:      "add_sum",
			Help:      "Sum of add transactions",
		},
		labels,
	)

	addCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "core",
			Subsystem: "ledger",
			Name:      "add_counter",
			Help:      "Number of add transactions",
		},
		labels,
	)

	removeSum := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "core",
			Subsystem: "ledger",
			Name:      "remove_sum",
			Help:      "Sum of remove transactions",
		},
		labels,
	)

	removeCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "core",
			Subsystem: "ledger",
			Name:      "remove_counter",
			Help:      "Number of remove transactions",
		},
		labels,
	)

	return &TxCollector{
		addSum:      addSum,
		addCount:    addCount,
		removeSum:   removeSum,
		removeCount: removeCount,
		database:    cfg.ClientOptions.Database,
		dbServer:    cfg.ClientOptions.Address,
		ns:          ns,
		ip:          ip,
	}
}

func (c *TxCollector) Add(asset types.Asset, value decimal.Decimal) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("collector error: %v", r)
		}
	}()

	fvalue, _ := value.Float64()
	hostname, _ := os.Hostname()

	labels := []string{
		asset.String(),
		hostname,
		c.database,
		c.dbServer,
	}

	if c.ns != "" {
		labels = append(labels, c.ns)
	}

	if c.ip != "" {
		labels = append(labels, c.ip)
	}

	if value.IsNegative() {
		c.removeSum.WithLabelValues(labels...).Add(-fvalue)
		c.removeCount.WithLabelValues(labels...).Inc()
	} else {
		c.addSum.WithLabelValues(labels...).Add(fvalue)
		c.addCount.WithLabelValues(labels...).Inc()
	}
}

func (c TxCollector) Collect(channel chan<- prometheus.Metric) {
	c.addSum.Collect(channel)
	c.addCount.Collect(channel)
	c.removeSum.Collect(channel)
	c.removeCount.Collect(channel)
}

func (c TxCollector) Describe(channel chan<- *prometheus.Desc) {
	c.addSum.Describe(channel)
	c.addCount.Describe(channel)
	c.removeSum.Describe(channel)
	c.removeCount.Describe(channel)
}
