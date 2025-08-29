package prometheus

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	port := "9090"
	endpoint := "/metrics"

	server := NewServer(port, endpoint)

	assert.NotNil(t, server)
	assert.Equal(t, port, server.port)
	assert.Equal(t, endpoint, server.endpoint)
	assert.NotNil(t, server.server)
	assert.Equal(t, ":"+port, server.server.Addr)
	assert.Equal(t, 5*time.Second, server.server.ReadTimeout)
	assert.Equal(t, 10*time.Second, server.server.WriteTimeout)
	assert.Equal(t, 15*time.Second, server.server.IdleTimeout)
}

func TestServer_Getters(t *testing.T) {
	port := "9091"
	endpoint := "/prometheus"

	server := NewServer(port, endpoint)

	assert.Equal(t, port, server.Port())
	assert.Equal(t, endpoint, server.Endpoint())
	assert.Equal(t, "http://localhost:9091/prometheus", server.URL())
}

func TestServer_StartAndShutdown(t *testing.T) {
	port := findFreePort(t)
	endpoint := "/metrics"

	server := NewServer(port, endpoint)

	// Start server in goroutine
	startErr := make(chan error, 1)
	go func() {
		startErr <- server.Start()
	}()

	// Wait for server to start
	waitForServer(t, port, 2*time.Second)

	// Test that server is responding
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/health", port))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	assert.NoError(t, err)

	// Verify server stopped
	select {
	case err := <-startErr:
		assert.Equal(t, http.ErrServerClosed, err)
	case <-time.After(time.Second):
		t.Fatal("Server did not stop within timeout")
	}
}

func TestServer_HealthEndpoint(t *testing.T) {
	port := findFreePort(t)
	endpoint := "/metrics"

	server := NewServer(port, endpoint)

	// Start server
	go func() {
		server.Start()
	}()

	// Wait for server to start
	waitForServer(t, port, 2*time.Second)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Test health endpoint
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/health", port))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "OK", string(body))
}

func TestServer_MetricsEndpoint(t *testing.T) {
	port := findFreePort(t)
	endpoint := "/metrics"

	server := NewServer(port, endpoint)

	// Start server
	go func() {
		server.Start()
	}()

	// Wait for server to start
	waitForServer(t, port, 2*time.Second)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Test metrics endpoint
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s%s", port, endpoint))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Should contain prometheus metrics format
	buf := make([]byte, 1000)
	n, err := resp.Body.Read(buf)
	require.NoError(t, err)
	content := string(buf[:n])

	// Prometheus metrics typically contain # HELP and # TYPE comments
	assert.True(t, strings.Contains(content, "# HELP") || strings.Contains(content, "# TYPE"),
		"Response should contain Prometheus metrics format")
}

func TestServer_CustomEndpoint(t *testing.T) {
	port := findFreePort(t)
	customEndpoint := "/custom-metrics"

	server := NewServer(port, customEndpoint)

	// Start server
	go func() {
		server.Start()
	}()

	// Wait for server to start
	waitForServer(t, port, 2*time.Second)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Test custom endpoint
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s%s", port, customEndpoint))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Default /metrics should not exist
	resp, err = http.Get(fmt.Sprintf("http://localhost:%s/metrics", port))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestServer_ShutdownTimeout(t *testing.T) {
	port := findFreePort(t)
	endpoint := "/metrics"

	server := NewServer(port, endpoint)

	// Start server
	go func() {
		server.Start()
	}()

	// Wait for server to start
	waitForServer(t, port, 2*time.Second)

	// Test shutdown with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		assert.Contains(t, err.Error(), "context deadline exceeded")
	}
	// Note: shutdown might succeed even with short timeout if server shuts down quickly

	// Shutdown properly
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	server.Shutdown(ctx2)
}

func TestServer_InvalidPort(t *testing.T) {
	// Test with invalid port (should still create server, but fail on Start)
	server := NewServer("invalid-port", "/metrics")

	err := server.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown port")
}

func TestServer_PortAlreadyInUse(t *testing.T) {
	port := findFreePort(t)
	endpoint := "/metrics"

	// Start first server
	server1 := NewServer(port, endpoint)
	go func() {
		server1.Start()
	}()

	// Wait for first server to start
	waitForServer(t, port, 2*time.Second)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		server1.Shutdown(ctx)
	}()

	// Try to start second server on same port
	server2 := NewServer(port, endpoint)
	err := server2.Start()
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "bind")
}

func TestServer_MultipleEndpoints(t *testing.T) {
	port := findFreePort(t)
	endpoint := "/metrics"

	server := NewServer(port, endpoint)

	// Start server
	go func() {
		server.Start()
	}()

	// Wait for server to start
	waitForServer(t, port, 2*time.Second)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Test both endpoints
	endpoints := []struct {
		path           string
		expectedStatus int
	}{
		{"/health", http.StatusOK},
		{endpoint, http.StatusOK},
		{"/nonexistent", http.StatusNotFound},
	}

	for _, ep := range endpoints {
		t.Run("endpoint_"+ep.path, func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%s%s", port, ep.path))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, ep.expectedStatus, resp.StatusCode)
		})
	}
}

func TestServer_HTTPMethods(t *testing.T) {
	port := findFreePort(t)
	endpoint := "/metrics"

	server := NewServer(port, endpoint)

	// Start server
	go func() {
		server.Start()
	}()

	// Wait for server to start
	waitForServer(t, port, 2*time.Second)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Test different HTTP methods on health endpoint
	client := &http.Client{Timeout: time.Second}

	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD"}
	for _, method := range methods {
		t.Run("method_"+method, func(t *testing.T) {
			req, err := http.NewRequest(method, fmt.Sprintf("http://localhost:%s/health", port), nil)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Health endpoint should respond to all methods (default mux behavior)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestServer_Timeouts(t *testing.T) {
	port := findFreePort(t)
	endpoint := "/metrics"

	server := NewServer(port, endpoint)

	// Verify timeout configuration
	assert.Equal(t, 5*time.Second, server.server.ReadTimeout)
	assert.Equal(t, 10*time.Second, server.server.WriteTimeout)
	assert.Equal(t, 15*time.Second, server.server.IdleTimeout)
}

// Helper functions

// findFreePort finds an available port for testing
func findFreePort(t *testing.T) string {
	t.Helper()
	
	// Use a simple approach: try a range of ports
	for port := 19090; port < 19200; port++ {
		portStr := fmt.Sprintf("%d", port)
		
		// Try to bind to the port
		server := NewServer(portStr, "/test")
		
		// Quick check if port is free by trying to start and immediately shutdown
		go func() {
			server.Start()
		}()
		
		time.Sleep(10 * time.Millisecond)
		
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		err := server.Shutdown(ctx)
		cancel()
		
		if err == nil || err == http.ErrServerClosed {
			return portStr
		}
	}
	
	t.Fatal("Could not find free port for testing")
	return ""
}

// waitForServer waits for a server to be ready to accept connections
func waitForServer(t *testing.T, port string, timeout time.Duration) {
	t.Helper()
	
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := http.Get(fmt.Sprintf("http://localhost:%s/health", port))
		if err == nil {
			conn.Body.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	
	t.Fatalf("Server on port %s did not start within %v", port, timeout)
}