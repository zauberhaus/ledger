package metrics

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	requestsName = "requests_total"
	latencyName  = "request_duration_milliseconds"
)

type MetricsMiddleware struct {
	requests *prometheus.CounterVec
	latency  *prometheus.HistogramVec
}

func NewMiddleware(name string, buckets ...float64) func(next http.Handler) http.Handler {
	var m MetricsMiddleware
	m.requests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        requestsName,
			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(m.requests)

	if len(buckets) == 0 {
		buckets = []float64{300, 1200, 5000}
	}
	m.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        latencyName,
		Help:        "How long it took to process the request, partitioned by status code, method and HTTP path.",
		ConstLabels: prometheus.Labels{"service": name},
		Buckets:     buckets,
	},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(m.latency)
	return m.handler
}

func (c MetricsMiddleware) handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/swagger") {
			next.ServeHTTP(w, r)
		} else {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			c.requests.WithLabelValues(http.StatusText(ww.Status()), r.Method, r.URL.Path).Inc()
			c.latency.WithLabelValues(http.StatusText(ww.Status()), r.Method, r.URL.Path).Observe(float64(time.Since(start).Nanoseconds()) / 1000000)
		}
	}
	return http.HandlerFunc(fn)
}
