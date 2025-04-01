package monitoring

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kv_store_requests_total",
			Help: "Total number of KV store requests",
		},
		[]string{"operation", "status"},
	)
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kv_store_request_duration_seconds",
			Help:    "Duration of KV store requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

func init() {
	prometheus.MustRegister(RequestsTotal)
	prometheus.MustRegister(RequestDuration)
}

func StartMetricsServer(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Fatal(http.ListenAndServe(addr, nil))
	}()
}
