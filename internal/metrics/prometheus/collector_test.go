package prometheus

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollector_URLMetrics(t *testing.T) {
	// Create a new registry to avoid conflicts with global metrics
	registry := prometheus.NewRegistry()
	
	// Create a collector with custom registry (for testing)
	collector := &Collector{
		urlsCreatedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "test",
			Name:      "urls_created_total",
			Help:      "Test counter",
		}),
		urlsRetrievedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "test",
			Name:      "urls_retrieved_total", 
			Help:      "Test counter",
		}),
		cacheHitsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "test",
			Name:      "cache_hits_total",
			Help:      "Test counter",
		}),
		cacheMissesTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "test",
			Name:      "cache_misses_total",
			Help:      "Test counter",
		}),
	}

	// Register metrics
	registry.MustRegister(collector.urlsCreatedTotal)
	registry.MustRegister(collector.urlsRetrievedTotal)
	registry.MustRegister(collector.cacheHitsTotal)
	registry.MustRegister(collector.cacheMissesTotal)

	// Test URL creation metrics
	collector.IncrementURLsCreated()
	collector.IncrementURLsCreated()
	
	value := testutil.ToFloat64(collector.urlsCreatedTotal)
	assert.Equal(t, float64(2), value)

	// Test URL retrieval metrics
	collector.IncrementURLsRetrieved()
	value = testutil.ToFloat64(collector.urlsRetrievedTotal)
	assert.Equal(t, float64(1), value)

	// Test cache metrics
	collector.IncrementCacheHits()
	collector.IncrementCacheHits()
	collector.IncrementCacheMisses()
	
	hits := testutil.ToFloat64(collector.cacheHitsTotal)
	misses := testutil.ToFloat64(collector.cacheMissesTotal)
	
	assert.Equal(t, float64(2), hits)
	assert.Equal(t, float64(1), misses)
}

func TestCollector_NewCollector(t *testing.T) {
	// Test that NewCollector creates a collector without panicking
	collector := NewCollector("test_namespace")
	require.NotNil(t, collector)
	
	// Test that metrics can be called without error
	collector.IncrementURLsCreated()
	collector.IncrementCacheHits()
	collector.RecordHTTPRequest("GET", "/api/urls", "200", 50*time.Millisecond)
	collector.RecordUptime(time.Now().Add(-1 * time.Hour))
}

func TestCollector_HTTPMetrics(t *testing.T) {
	collector := NewCollector("test_http")
	
	// Test HTTP request recording
	duration := 100 * time.Millisecond
	
	collector.RecordHTTPRequest("POST", "/api/urls", "201", duration)
	collector.IncrementHTTPRequests("GET", "/api/urls", "200")
	
	// These should not panic and should update the underlying metrics
	// In a real test environment, we would check the actual metric values
	// but that would require more complex setup with custom registries
}

func TestCollector_CacheMetrics(t *testing.T) {
	collector := NewCollector("test_cache")
	
	// Test cache operation metrics
	collector.RecordCacheOperationDuration("get", 5*time.Millisecond)
	collector.RecordCacheOperationDuration("set", 3*time.Millisecond)
	collector.RecordCacheSize(150)
	collector.IncrementCacheEvictions()
	
	// These should not panic
}

func TestCollector_DatabaseMetrics(t *testing.T) {
	collector := NewCollector("test_db")
	
	// Test database metrics
	collector.IncrementDBQueries("select")
	collector.IncrementDBQueries("insert")
	collector.IncrementDBErrors("select", "timeout")
	collector.RecordDBQueryDuration("insert", 10*time.Millisecond)
	collector.RecordDBConnectionsActive(5)
	
	// These should not panic
}

func TestCollector_SyncMetrics(t *testing.T) {
	collector := NewCollector("test_sync")
	
	// Test cache sync metrics
	collector.IncrementCacheSyncOperations()
	collector.IncrementCacheSyncErrors()
	collector.RecordCacheSyncDuration(200 * time.Millisecond)
	collector.RecordCacheSyncEntriesProcessed(25)
	
	// These should not panic
}