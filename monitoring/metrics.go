// monitoring/metrics.go
package monitoring

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/cobrich/netcfg-backup/utils"
)

var (
	// JobsTotal - a counter for the total number of backup jobs processed.
	// Labels: host, status (success/failed)
	JobsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "netcfg_backup_jobs_total",
			Help: "Total number of backup jobs processed.",
		},
		[]string{"host", "status"},
	)

	// JobDuration - a histogram of the backup job duration in seconds.
	// Labels: host
	JobDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "netcfg_backup_job_duration_seconds",
			Help:    "Duration of backup jobs in seconds.",
			Buckets: prometheus.LinearBuckets(1, 5, 10), // 10 buckets, starting at 1s, with 5s intervals
		},
		[]string{"host"},
	)
)

// StartMetricsServer starts an HTTP server to expose Prometheus metrics.
func StartMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	
	// Run the server in a separate goroutine so it doesn't block the main app.
	go func() {
		utils.Log.Info("Starting metrics server on :9090/metrics")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			utils.Log.Errorf("Metrics server failed to start: %v", err)
		}
	}()
}