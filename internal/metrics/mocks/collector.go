package mocks

import (
	"time"

	"github.com/stretchr/testify/mock"
)

// Collector is a mock implementation of metrics.Collector
type Collector struct {
	mock.Mock
}

func (m *Collector) IncrementURLsCreated() {
	m.Called()
}

func (m *Collector) IncrementURLsCreatedWithError(errorType string) {
	m.Called(errorType)
}

func (m *Collector) IncrementURLsRetrieved() {
	m.Called()
}

func (m *Collector) IncrementURLsRetrievedFromCache() {
	m.Called()
}

func (m *Collector) IncrementURLsRetrievedFromDB() {
	m.Called()
}

func (m *Collector) IncrementURLsRetrievedWithError(errorType string) {
	m.Called(errorType)
}

func (m *Collector) IncrementURLsDeleted() {
	m.Called()
}

func (m *Collector) IncrementURLsDeletedWithError(errorType string) {
	m.Called(errorType)
}

func (m *Collector) RecordHTTPRequest(method, endpoint, status string, duration time.Duration) {
	m.Called(method, endpoint, status, duration)
}

func (m *Collector) IncrementHTTPRequests(method, endpoint, status string) {
	m.Called(method, endpoint, status)
}

func (m *Collector) IncrementCacheHits() {
	m.Called()
}

func (m *Collector) IncrementCacheMisses() {
	m.Called()
}

func (m *Collector) IncrementCacheEvictions() {
	m.Called()
}

func (m *Collector) RecordCacheOperationDuration(operation string, duration time.Duration) {
	m.Called(operation, duration)
}

func (m *Collector) RecordCacheSize(size int) {
	m.Called(size)
}

func (m *Collector) IncrementDBQueries(queryType string) {
	m.Called(queryType)
}

func (m *Collector) IncrementDBErrors(queryType, errorType string) {
	m.Called(queryType, errorType)
}

func (m *Collector) RecordDBQueryDuration(queryType string, duration time.Duration) {
	m.Called(queryType, duration)
}

func (m *Collector) RecordDBConnectionsActive(count int) {
	m.Called(count)
}

func (m *Collector) IncrementCacheSyncOperations() {
	m.Called()
}

func (m *Collector) IncrementCacheSyncErrors() {
	m.Called()
}

func (m *Collector) RecordCacheSyncDuration(duration time.Duration) {
	m.Called(duration)
}

func (m *Collector) RecordCacheSyncEntriesProcessed(count int) {
	m.Called(count)
}

func (m *Collector) RecordUptime(startTime time.Time) {
	m.Called(startTime)
}

func (m *Collector) RecordMemoryUsage(bytes uint64) {
	m.Called(bytes)
}

func (m *Collector) RecordActiveConnections(count int) {
	m.Called(count)
}