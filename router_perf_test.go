package fastrouter

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// Simple handler for testing
type testHandler struct {
	name string
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(h.name))
}

func TestCurrentRouterPerformance(t *testing.T) {
	fmt.Println("=== Current Router Performance Analysis ===")
	
	// Build router with test routes
	rb := NewRouterBuilder()
	
	// Add routes in lexicographic order (as required by the current implementation)
	routes := []struct{ path, name string }{
		{"/", "home"},
		{"/admin", "admin_home"},
		{"/admin/users", "admin_users"},
		{"/api", "api_home"},
		{"/api/users", "api_users"},
		{"/api/users/:id", "api_user"},
		{"/api/users/:id/posts", "api_posts"},
		{"/users", "users"},
		{"/users/:id", "user"},
		{"/users/:id/settings", "user_settings"},
	}
	
	for _, route := range routes {
		handler := &testHandler{name: route.name}
		err := rb.AddRoute("GET", route.path, handler)
		if err != nil {
			t.Fatalf("Failed to add route %s: %v", route.path, err)
		}
	}
	
	router, err := rb.Build()
	if err != nil {
		t.Fatalf("Failed to build router: %v", err)
	}
	
	// Test paths that exercise different route types
	testPaths := []string{
		"/",                    // root
		"/users",              // static
		"/users/123",          // single param
		"/users/123/settings", // nested param
		"/api/users/456",      // nested with param
		"/api/users/456/posts", // deep nested with param
		"/admin/users",        // static nested
		"/nonexistent",        // 404 case
	}
	
	iterations := 100000
	
	fmt.Printf("Testing with %d iterations across %d different paths\n", iterations, len(testPaths))
	
	start := time.Now()
	var matches, misses int
	
	for i := 0; i < iterations; i++ {
		path := testPaths[i%len(testPaths)]
		handler, _ := router.Match("GET", path)
		if handler != nil {
			matches++
		} else {
			misses++
		}
	}
	
	duration := time.Since(start)
	nsPerOp := duration.Nanoseconds() / int64(iterations)
	opsPerSec := float64(iterations) / duration.Seconds()
	
	fmt.Printf("Results:\n")
	fmt.Printf("  Total time: %v\n", duration)
	fmt.Printf("  Average per operation: %d ns\n", nsPerOp)
	fmt.Printf("  Operations per second: %.0f\n", opsPerSec)
	fmt.Printf("  Successful matches: %d\n", matches)
	fmt.Printf("  Misses (404s): %d\n", misses)
	
	// Print router statistics
	fmt.Printf("\nRouter Statistics:\n")
	stats := router.Stats()
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

func BenchmarkCurrentRouter(b *testing.B) {
	// Build router
	rb := NewRouterBuilder()
	
	routes := []string{
		"/",
		"/api",
		"/api/users", 
		"/api/users/:id",
		"/api/users/:id/posts",
		"/users",
		"/users/:id",
		"/users/:id/settings",
	}
	
	for _, route := range routes {
		rb.AddRoute("GET", route, &testHandler{name: "test"})
	}
	
	router, _ := rb.Build()
	
	testPaths := []string{
		"/",
		"/users",
		"/users/123",
		"/users/123/settings", 
		"/api/users/456",
		"/api/users/456/posts",
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		path := testPaths[i%len(testPaths)]
		router.Match("GET", path)
	}
}
