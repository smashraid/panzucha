package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// --- Users ---
	UsersCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "users_created_total",
			Help: "Total number of users created",
		},
		[]string{"role"}, // "admin", "user"
	)
	UsersUpdated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "users_updated_total",
			Help: "Total number of user updates",
		},
	)
	UsersDeleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "users_deleted_total",
			Help: "Total number of user deletions",
		},
	)

	// --- Products ---
	ProductsCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "products_created_total",
			Help: "Total number of products created",
		},
	)
	ProductsUpdated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "products_updated_total",
			Help: "Total number of product updates",
		},
	)
	ProductsDeleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "products_deleted_total",
			Help: "Total number of product deletions",
		},
	)
	// Stock decrement counter (can be used to track inventory changes)
	StockDecrements = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "stock_decrements_total",
			Help: "Total number of stock decrement operations",
		},
	)

	// --- Orders ---
	OrdersCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total number of orders created",
		},
	)
	OrdersStatusChanged = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_status_changed_total",
			Help: "Total number of order status changes",
		},
		[]string{"from_status", "to_status"},
	)

	// Gauge for current pending orders (could be updated on create/status change)
	PendingOrders = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "orders_pending_current",
			Help: "Current number of pending orders",
		},
	)
	OrdersTotalPrice = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_total_price_sum",
			Help: "Sum of total price of all orders (can be used for revenue tracking)",
		},
	)

	// --- Database operation duration histogram ---
	DBOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_operation_duration_seconds",
			Help:    "Duration of database operations",
			Buckets: prometheus.DefBuckets, // default buckets: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
		},
		[]string{"operation", "table"},
	)

	// --- HTTP request duration (already provided by otelhttp, but we can add custom) ---
	// Not needed if using otelhttp, which exposes http_server_duration_ms.
)

// Helper to increment status change counter
func IncrementOrderStatusChange(from, to string) {
	OrdersStatusChanged.WithLabelValues(from, to).Inc()
}

// Helper to update pending orders gauge (call when status changes to/from pending)
func UpdatePendingOrdersGauge(delta int) {
	PendingOrders.Add(float64(delta))
}
