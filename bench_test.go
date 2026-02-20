package fastrouter

import (
	"net/http"
	"testing"
)

var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func BenchmarkSimpleMatch(b *testing.B) {
	rb := NewRouterBuilder()
	rb.AddRoute("GET", "/", handler)
	rb.AddRoute("GET", "/api/users", handler)
	rb.AddRoute("GET", "/api/users/:id", handler)
	rb.AddRoute("GET", "/api/posts", handler)
	
	router, _ := rb.Build()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.Match("GET", "/api/users/123")
	}
}

func BenchmarkRouterBuild(b *testing.B) {
	routes := []struct{ method, path string }{
		{"GET", "/"},
		{"GET", "/api/users"},
		{"GET", "/api/users/:id"},
		{"GET", "/api/posts"},
		{"GET", "/api/posts/:id"},
		{"GET", "/health"},
		{"GET", "/metrics"},
		{"GET", "/status"},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb := NewRouterBuilder()
		for _, route := range routes {
			rb.AddRoute(route.method, route.path, handler)
		}
		rb.Build()
	}
}
