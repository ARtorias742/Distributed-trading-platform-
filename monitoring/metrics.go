package monitoring

import "github.com/prometheus/client_golang/prometheus"

var (
    OrderLatency = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "order_processing_latency_seconds",
            Buckets: prometheus.LinearBuckets(0.01, 0.05, 20),
        },
    )
)

func InitMetrics() {
    prometheus.MustRegister(OrderLatency)
}