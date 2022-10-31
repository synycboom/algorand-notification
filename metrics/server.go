package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus metric names broken out for reuse.
const (
	ActiveConnectionsName   = "active_connections"
	ActiveSubscriptionsName = "active_subscription"
)

// RegisterServerMetrics registers metrics related to the server
func RegisterServerMetrics() {
	prometheus.Register(ActiveConnections)
	prometheus.Register(ActiveSubscriptions)
}

var (
	ActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: "server",
			Name:      ActiveConnectionsName,
			Help:      "Total active connections",
		},
	)

	ActiveSubscriptions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "server",
			Name:      ActiveSubscriptionsName,
			Help:      "Total active subscriptions by name",
		},
		[]string{"name"},
	)
)
