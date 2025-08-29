package metrics

import (
	"time"
)

// Collector defines the interface for collecting application metrics
type Collector interface {
	// URL creation metrics
	IncrementURLsCreated()
	IncrementURLsCreatedWithError(errorType string)
	
	// URL retrieval metrics
	IncrementURLsRetrieved()
	IncrementURLsRetrievedFromCache()
	IncrementURLsRetrievedFromDB()
	IncrementURLsRetrievedWithError(errorType string)
	
	// URL deletion metrics
	IncrementURLsDeleted()
	IncrementURLsDeletedWithError(errorType string)
	
	// HTTP request metrics
	RecordHTTPRequest(method, endpoint, status string, duration time.Duration)
	IncrementHTTPRequests(method, endpoint, status string)
	
	// Cache metrics
	IncrementCacheHits()
	IncrementCacheMisses()
	IncrementCacheEvictions()
	RecordCacheOperationDuration(operation string, duration time.Duration)
	RecordCacheSize(size int)
	
	// Database metrics
	IncrementDBQueries(queryType string)
	IncrementDBErrors(queryType, errorType string)
	RecordDBQueryDuration(queryType string, duration time.Duration)
	RecordDBConnectionsActive(count int)
	
	// Background sync metrics
	IncrementCacheSyncOperations()
	IncrementCacheSyncErrors()
	RecordCacheSyncDuration(duration time.Duration)
	RecordCacheSyncEntriesProcessed(count int)
	
	// General application metrics
	RecordUptime(startTime time.Time)
	RecordMemoryUsage(bytes uint64)
	RecordActiveConnections(count int)
}

// NoOpCollector is a metrics collector that does nothing (useful for testing or when metrics are disabled)
type NoOpCollector struct{}

func (n *NoOpCollector) IncrementURLsCreated()                                           {}
func (n *NoOpCollector) IncrementURLsCreatedWithError(errorType string)                 {}
func (n *NoOpCollector) IncrementURLsRetrieved()                                        {}
func (n *NoOpCollector) IncrementURLsRetrievedFromCache()                               {}
func (n *NoOpCollector) IncrementURLsRetrievedFromDB()                                  {}
func (n *NoOpCollector) IncrementURLsRetrievedWithError(errorType string)               {}
func (n *NoOpCollector) IncrementURLsDeleted()                                          {}
func (n *NoOpCollector) IncrementURLsDeletedWithError(errorType string)                 {}
func (n *NoOpCollector) RecordHTTPRequest(method, endpoint, status string, duration time.Duration) {}
func (n *NoOpCollector) IncrementHTTPRequests(method, endpoint, status string)          {}
func (n *NoOpCollector) IncrementCacheHits()                                            {}
func (n *NoOpCollector) IncrementCacheMisses()                                          {}
func (n *NoOpCollector) IncrementCacheEvictions()                                       {}
func (n *NoOpCollector) RecordCacheOperationDuration(operation string, duration time.Duration) {}
func (n *NoOpCollector) RecordCacheSize(size int)                                       {}
func (n *NoOpCollector) IncrementDBQueries(queryType string)                            {}
func (n *NoOpCollector) IncrementDBErrors(queryType, errorType string)                  {}
func (n *NoOpCollector) RecordDBQueryDuration(queryType string, duration time.Duration) {}
func (n *NoOpCollector) RecordDBConnectionsActive(count int)                            {}
func (n *NoOpCollector) IncrementCacheSyncOperations()                                  {}
func (n *NoOpCollector) IncrementCacheSyncErrors()                                      {}
func (n *NoOpCollector) RecordCacheSyncDuration(duration time.Duration)                {}
func (n *NoOpCollector) RecordCacheSyncEntriesProcessed(count int)                      {}
func (n *NoOpCollector) RecordUptime(startTime time.Time)                              {}
func (n *NoOpCollector) RecordMemoryUsage(bytes uint64)                                {}
func (n *NoOpCollector) RecordActiveConnections(count int)                             {}

// Ensure NoOpCollector implements Collector interface
var _ Collector = (*NoOpCollector)(nil)