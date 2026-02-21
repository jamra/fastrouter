package fastrouter

import (
	"time"
	"fmt"
	"net/http"
	"testing"
	
	"github.com/julienschmidt/httprouter"
)

func BenchmarkHttpRouter(b *testing.B) {
	router := httprouter.New()
	
	// Add the same routes as our router
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write([]byte("home"))
	})
	router.GET("/api", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write([]byte("api"))
	})
	router.GET("/api/users", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write([]byte("users"))
	})
	router.GET("/api/users/:id", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write([]byte("user"))
	})
	router.GET("/api/users/:id/posts", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write([]byte("posts"))
	})
	router.GET("/users", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write([]byte("users"))
	})
	router.GET("/users/:id", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write([]byte("user"))
	})
	router.GET("/users/:id/settings", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write([]byte("settings"))
	})
	
	testPaths := []string{
		"/",
		"/users",
		"/users/123",
		"/users/123/settings",
		"/api/users/456", 
		"/api/users/456/posts",
	}
	
	// Create requests for testing
	requests := make([]*http.Request, len(testPaths))
	for i, path := range testPaths {
		requests[i], _ = http.NewRequest("GET", path, nil)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		req := requests[i%len(requests)]
		handler, _, _ := router.Lookup("GET", req.URL.Path)
		_ = handler // Use the handler to prevent optimization
	}
}

func TestRouterComparison(t *testing.T) {
	fmt.Println("\n=== Router Performance Comparison ===")
	
	// Test httprouter
	fmt.Println("Setting up httprouter...")
	httpRouter := httprouter.New()
	httpRouter.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	httpRouter.GET("/users", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	httpRouter.GET("/users/:id", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	httpRouter.GET("/users/:id/settings", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	httpRouter.GET("/api/users/:id", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	httpRouter.GET("/api/users/:id/posts", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	
	// Test fastrouter
	fmt.Println("Setting up fastrouter...")
	rb := NewRouterBuilder()
	rb.AddRoute("GET", "/", &testHandler{})
	rb.AddRoute("GET", "/api/users/:id", &testHandler{})
	rb.AddRoute("GET", "/api/users/:id/posts", &testHandler{})
	rb.AddRoute("GET", "/users", &testHandler{})
	rb.AddRoute("GET", "/users/:id", &testHandler{})
	rb.AddRoute("GET", "/users/:id/settings", &testHandler{})
	fastRouter, _ := rb.Build()
	
	testPaths := []string{
		"/",
		"/users",
		"/users/123", 
		"/users/123/settings",
		"/api/users/456",
		"/api/users/456/posts",
	}
	
	iterations := 1000000
	
	// Benchmark httprouter
	fmt.Printf("Benchmarking httprouter (%d iterations)...\n", iterations)
	start := time.Now()
	for i := 0; i < iterations; i++ {
		path := testPaths[i%len(testPaths)]
		httpRouter.Lookup("GET", path)
	}
	httpRouterTime := time.Since(start)
	
	// Benchmark fastrouter
	fmt.Printf("Benchmarking fastrouter (%d iterations)...\n", iterations)
	start = time.Now()
	for i := 0; i < iterations; i++ {
		path := testPaths[i%len(testPaths)]
		fastRouter.Match("GET", path)
	}
	fastRouterTime := time.Since(start)
	
	fmt.Printf("\nResults:\n")
	fmt.Printf("httprouter:  %v (%d ns/op)\n", httpRouterTime, httpRouterTime.Nanoseconds()/int64(iterations))
	fmt.Printf("fastrouter:  %v (%d ns/op)\n", fastRouterTime, fastRouterTime.Nanoseconds()/int64(iterations))
	
	if fastRouterTime < httpRouterTime {
		speedup := float64(httpRouterTime) / float64(fastRouterTime)
		fmt.Printf("fastrouter is %.2fx FASTER than httprouter! 🎉\n", speedup)
	} else {
		slowdown := float64(fastRouterTime) / float64(httpRouterTime)
		fmt.Printf("fastrouter is %.2fx slower than httprouter (room for improvement)\n", slowdown)
	}
}
