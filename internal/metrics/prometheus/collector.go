package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	
	"github.com/joshdurbin/url-shortener/internal/metrics"
)

// Collector implements the metrics.Collector interface using Prometheus
type Collector struct {
	// URL metrics
	urlsCreatedTotal         prometheus.Counter
	urlsCreatedErrorsTotal   *prometheus.CounterVec
	urlsRetrievedTotal       prometheus.Counter
	urlsRetrievedCacheTotal  prometheus.Counter
	urlsRetrievedDBTotal     prometheus.Counter
	urlsRetrievedErrorsTotal *prometheus.CounterVec
	urlsDeletedTotal         prometheus.Counter
	urlsDeletedErrorsTotal   *prometheus.CounterVec

	// HTTP metrics
	httpRequestsTotal     *prometheus.CounterVec
	httpRequestDuration   *prometheus.HistogramVec
	httpActiveConnections prometheus.Gauge

	// Cache metrics
	cacheHitsTotal        prometheus.Counter
	cacheMissesTotal      prometheus.Counter
	cacheEvictionsTotal   prometheus.Counter
	cacheOperationDuration *prometheus.HistogramVec
	cacheSize             prometheus.Gauge

	// Database metrics
	dbQueriesTotal        *prometheus.CounterVec
	dbErrorsTotal         *prometheus.CounterVec
	dbQueryDuration       *prometheus.HistogramVec
	dbConnectionsActive   prometheus.Gauge

	// Cache sync metrics
	cacheSyncOperationsTotal     prometheus.Counter
	cacheSyncErrorsTotal         prometheus.Counter
	cacheSyncDuration           prometheus.Histogram
	cacheSyncEntriesProcessed   prometheus.Histogram

	// Application metrics
	applicationUptime      prometheus.Gauge
	memoryUsageBytes       prometheus.Gauge
}

// NewCollector creates a new Prometheus metrics collector
func NewCollector(namespace string) *Collector {
	c := &Collector{
		// URL metrics
		urlsCreatedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "urls_created_total",
			Help:      "Total number of URLs created",
		}),
		urlsCreatedErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "urls_created_errors_total",
			Help:      "Total number of URL creation errors",
		}, []string{"error_type"}),
		urlsRetrievedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "urls_retrieved_total",
			Help:      "Total number of URLs retrieved",
		}),
		urlsRetrievedCacheTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "urls_retrieved_cache_total",
			Help:      "Total number of URLs retrieved from cache",
		}),
		urlsRetrievedDBTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "urls_retrieved_db_total",
			Help:      "Total number of URLs retrieved from database",
		}),
		urlsRetrievedErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "urls_retrieved_errors_total",
			Help:      "Total number of URL retrieval errors",
		}, []string{"error_type"}),
		urlsDeletedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "urls_deleted_total",
			Help:      "Total number of URLs deleted",
		}),
		urlsDeletedErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "urls_deleted_errors_total",
			Help:      "Total number of URL deletion errors",
		}, []string{"error_type"}),

		// HTTP metrics
		httpRequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		}, []string{"method", "endpoint", "status"}),
		httpRequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "endpoint", "status"}),
		httpActiveConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "http_active_connections",
			Help:      "Number of active HTTP connections",
		}),

		// Cache metrics
		cacheHitsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "cache_hits_total",
			Help:      "Total number of cache hits",
		}),
		cacheMissesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "cache_misses_total",
			Help:      "Total number of cache misses",
		}),
		cacheEvictionsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "cache_evictions_total",
			Help:      "Total number of cache evictions",
		}),
		cacheOperationDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "cache_operation_duration_seconds",
			Help:      "Duration of cache operations in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		}, []string{"operation"}),
		cacheSize: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "cache_size_entries",
			Help:      "Number of entries in cache",
		}),

		// Database metrics
		dbQueriesTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "db_queries_total",
			Help:      "Total number of database queries",
		}, []string{"query_type"}),
		dbErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "db_errors_total",
			Help:      "Total number of database errors",
		}, []string{"query_type", "error_type"}),
		dbQueryDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "db_query_duration_seconds",
			Help:      "Duration of database queries in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		}, []string{"query_type"}),
		dbConnectionsActive: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_connections_active",
			Help:      "Number of active database connections",
		}),

		// Cache sync metrics
		cacheSyncOperationsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "cache_sync_operations_total",
			Help:      "Total number of cache sync operations",
		}),
		cacheSyncErrorsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "cache_sync_errors_total",
			Help:      "Total number of cache sync errors",
		}),
		cacheSyncDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "cache_sync_duration_seconds",
			Help:      "Duration of cache sync operations in seconds",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
		}),
		cacheSyncEntriesProcessed: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "cache_sync_entries_processed",
			Help:      "Number of entries processed during cache sync",
			Buckets:   []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
		}),

		// Application metrics
		applicationUptime: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "application_uptime_seconds",
			Help:      "Application uptime in seconds",
		}),
		memoryUsageBytes: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "memory_usage_bytes",
			Help:      "Memory usage in bytes",
		}),
	}

	return c
}

// URL creation metrics
func (c *Collector) IncrementURLsCreated() {
	c.urlsCreatedTotal.Inc()
}

func (c *Collector) IncrementURLsCreatedWithError(errorType string) {
	c.urlsCreatedErrorsTotal.WithLabelValues(errorType).Inc()
}

// URL retrieval metrics
func (c *Collector) IncrementURLsRetrieved() {
	c.urlsRetrievedTotal.Inc()
}

func (c *Collector) IncrementURLsRetrievedFromCache() {
	c.urlsRetrievedCacheTotal.Inc()
}

func (c *Collector) IncrementURLsRetrievedFromDB() {
	c.urlsRetrievedDBTotal.Inc()
}

func (c *Collector) IncrementURLsRetrievedWithError(errorType string) {
	c.urlsRetrievedErrorsTotal.WithLabelValues(errorType).Inc()
}

// URL deletion metrics
func (c *Collector) IncrementURLsDeleted() {
	c.urlsDeletedTotal.Inc()
}

func (c *Collector) IncrementURLsDeletedWithError(errorType string) {
	c.urlsDeletedErrorsTotal.WithLabelValues(errorType).Inc()
}

// HTTP request metrics
func (c *Collector) RecordHTTPRequest(method, endpoint, status string, duration time.Duration) {
	c.httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	c.httpRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration.Seconds())
}

func (c *Collector) IncrementHTTPRequests(method, endpoint, status string) {
	c.httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
}

// Cache metrics
func (c *Collector) IncrementCacheHits() {
	c.cacheHitsTotal.Inc()
}

func (c *Collector) IncrementCacheMisses() {
	c.cacheMissesTotal.Inc()
}

func (c *Collector) IncrementCacheEvictions() {
	c.cacheEvictionsTotal.Inc()
}

func (c *Collector) RecordCacheOperationDuration(operation string, duration time.Duration) {
	c.cacheOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

func (c *Collector) RecordCacheSize(size int) {
	c.cacheSize.Set(float64(size))
}

// Database metrics
func (c *Collector) IncrementDBQueries(queryType string) {
	c.dbQueriesTotal.WithLabelValues(queryType).Inc()
}

func (c *Collector) IncrementDBErrors(queryType, errorType string) {
	c.dbErrorsTotal.WithLabelValues(queryType, errorType).Inc()
}

func (c *Collector) RecordDBQueryDuration(queryType string, duration time.Duration) {
	c.dbQueryDuration.WithLabelValues(queryType).Observe(duration.Seconds())
}

func (c *Collector) RecordDBConnectionsActive(count int) {
	c.dbConnectionsActive.Set(float64(count))
}

// Background sync metrics
func (c *Collector) IncrementCacheSyncOperations() {
	c.cacheSyncOperationsTotal.Inc()
}

func (c *Collector) IncrementCacheSyncErrors() {
	c.cacheSyncErrorsTotal.Inc()
}

func (c *Collector) RecordCacheSyncDuration(duration time.Duration) {
	c.cacheSyncDuration.Observe(duration.Seconds())
}

func (c *Collector) RecordCacheSyncEntriesProcessed(count int) {
	c.cacheSyncEntriesProcessed.Observe(float64(count))
}

// General application metrics
func (c *Collector) RecordUptime(startTime time.Time) {
	c.applicationUptime.Set(time.Since(startTime).Seconds())
}

func (c *Collector) RecordMemoryUsage(bytes uint64) {
	c.memoryUsageBytes.Set(float64(bytes))
}

func (c *Collector) RecordActiveConnections(count int) {
	c.httpActiveConnections.Set(float64(count))
}

// Ensure Collector implements metrics.Collector interface
var _ metrics.Collector = (*Collector)(nil)