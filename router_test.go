package fastrouter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterBuilderOrder(t *testing.T) {
	rb := NewRouterBuilder()

	// Test that routes must be added in lexicographic order
	err := rb.AddRoute("GET", "/users", nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = rb.AddRoute("GET", "/posts", nil) // This should fail as "posts" < "users"
	if err == nil {
		t.Errorf("Expected error for out-of-order route addition")
	}
}

func TestRouterBuilderCorrectOrder(t *testing.T) {
	rb := NewRouterBuilder()

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/api"},
		{"GET", "/api/posts"},
		{"GET", "/api/users"},
		{"POST", "/api/users"},
		{"GET", "/api/users/:id"},
		{"GET", "/files/*"},
		{"GET", "/users"},
	}

	for _, route := range routes {
		err := rb.AddRoute(route.method, route.path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		if err != nil {
			t.Errorf("Unexpected error adding route %s %s: %v", route.method, route.path, err)
		}
	}

	router, err := rb.Build()
	if err != nil {
		t.Errorf("Error building router: %v", err)
	}

	if router == nil {
		t.Error("Expected router to be non-nil")
	}
}

func TestBasicRouteMatching(t *testing.T) {
	rb := NewRouterBuilder()

	// Add routes in lexicographic order
	testHandler := func(expected string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(expected))
		})
	}

	routes := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/", "root"},
		{"GET", "/api/posts", "posts"},
		{"GET", "/api/users", "users"},
		{"POST", "/api/users", "create-user"},
		{"GET", "/users/profile", "profile"},
	}

	for _, route := range routes {
		err := rb.AddRoute(route.method, route.path, testHandler(route.expected))
		if err != nil {
			t.Fatalf("Error adding route: %v", err)
		}
	}

	router, err := rb.Build()
	if err != nil {
		t.Fatalf("Error building router: %v", err)
	}

	// Test exact matches
	testCases := []struct {
		method   string
		path     string
		expected string
		found    bool
	}{
		{"GET", "/", "root", true},
		{"GET", "/api/posts", "posts", true},
		{"GET", "/api/users", "users", true},
		{"POST", "/api/users", "create-user", true},
		{"GET", "/users/profile", "profile", true},
		{"DELETE", "/api/users", "", false}, // Wrong method
		{"GET", "/api/nonexistent", "", false}, // Nonexistent path
		{"GET", "/api", "", false}, // Partial path
	}

	for _, tc := range testCases {
		handler, _ := router.Match(tc.method, tc.path)
		if tc.found {
			if handler == nil {
				t.Errorf("Expected handler for %s %s, got nil", tc.method, tc.path)
				continue
			}
			
			// Test the handler
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tc.method, tc.path, nil)
			handler.ServeHTTP(w, r)
			
			if w.Body.String() != tc.expected {
				t.Errorf("Expected response '%s' for %s %s, got '%s'", 
					tc.expected, tc.method, tc.path, w.Body.String())
			}
		} else {
			if handler != nil {
				t.Errorf("Expected no handler for %s %s, got one", tc.method, tc.path)
			}
		}
	}
}

func TestParameterRoutes(t *testing.T) {
	rb := NewRouterBuilder()

	paramHandler := func(expectedParam string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(fmt.Sprintf("param:%s", expectedParam)))
		})
	}

	// Add routes in lexicographic order (parameters come after static routes)
	routes := []struct {
		method string
		path   string
		name   string
	}{
		{"GET", "/users", "static"},
		{"GET", "/users/:id", "id"},
		{"GET", "/users/:id/posts", "id-posts"},
		{"GET", "/users/:id/posts/:postId", "post"},
	}

	for _, route := range routes {
		err := rb.AddRoute(route.method, route.path, paramHandler(route.name))
		if err != nil {
			t.Fatalf("Error adding route %s %s: %v", route.method, route.path, err)
		}
	}

	router, err := rb.Build()
	if err != nil {
		t.Fatalf("Error building router: %v", err)
	}

	testCases := []struct {
		method       string
		path         string
		expectFound  bool
		expectedResp string
		expectedParams map[string]string
	}{
		{"GET", "/users", true, "param:static", map[string]string{}},
		{"GET", "/users/123", true, "param:id", map[string]string{"id": "123"}},
		{"GET", "/users/abc/posts", true, "param:id-posts", map[string]string{"id": "abc"}},
		{"GET", "/users/123/posts/456", true, "param:post", map[string]string{"id": "123", "postId": "456"}},
		{"GET", "/users/123/comments", false, "", nil},
	}

	for _, tc := range testCases {
		handler, params := router.Match(tc.method, tc.path)
		
		if tc.expectFound {
			if handler == nil {
				t.Errorf("Expected handler for %s %s, got nil", tc.method, tc.path)
				continue
			}

			// Check parameters
			for key, expected := range tc.expectedParams {
				if actual, exists := params[key]; !exists || actual != expected {
					t.Errorf("Expected param %s=%s for %s %s, got %s=%s", 
						key, expected, tc.method, tc.path, key, actual)
				}
			}

			// Test handler response
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tc.method, tc.path, nil)
			handler.ServeHTTP(w, r)
			
			if w.Body.String() != tc.expectedResp {
				t.Errorf("Expected response '%s' for %s %s, got '%s'", 
					tc.expectedResp, tc.method, tc.path, w.Body.String())
			}
		} else {
			if handler != nil {
				t.Errorf("Expected no handler for %s %s, got one", tc.method, tc.path)
			}
		}
	}
}

func TestWildcardRoutes(t *testing.T) {
	rb := NewRouterBuilder()

	wildcardHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("wildcard"))
	})

	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("static"))
	})

	// Add routes in lexicographic order
	err := rb.AddRoute("GET", "/files/*", wildcardHandler)
	if err != nil {
		t.Fatalf("Error adding static route: %v", err)
	}

	err = rb.AddRoute("GET", "/files/specific.txt", staticHandler)
	if err != nil {
		t.Fatalf("Error adding wildcard route: %v", err)
	}

	router, err := rb.Build()
	if err != nil {
		t.Fatalf("Error building router: %v", err)
	}

	testCases := []struct {
		path     string
		expected string
	}{
		{"/files/specific.txt", "static"},           // Exact match wins
		{"/files/other.txt", "wildcard"},       // Wildcard match
		{"/files/dir/file.txt", "wildcard"},    // Deep wildcard match
		{"/files/", "wildcard"},                // Wildcard at directory level
	}

	for _, tc := range testCases {
		handler, params := router.Match("GET", tc.path)
		if handler == nil {
			t.Errorf("Expected handler for %s, got nil", tc.path)
			continue
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", tc.path, nil)
		handler.ServeHTTP(w, r)

		if w.Body.String() != tc.expected {
			t.Errorf("Expected response '%s' for %s, got '%s'", 
				tc.expected, tc.path, w.Body.String())
		}

		// Check wildcard parameter
		if tc.expected == "wildcard" {
			if _, exists := params["*"]; !exists {
				t.Errorf("Expected wildcard parameter for %s", tc.path)
			}
		}
	}
}

func TestRouterStats(t *testing.T) {
	rb := NewRouterBuilder()

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/api/posts"},
		{"GET", "/api/users"},
		{"POST", "/api/users"},
		{"GET", "/api/users/:id"},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	for _, route := range routes {
		err := rb.AddRoute(route.method, route.path, handler)
		if err != nil {
			t.Fatalf("Error adding route: %v", err)
		}
	}

	router, err := rb.Build()
	if err != nil {
		t.Fatalf("Error building router: %v", err)
	}

	stats := router.Stats()
	
	if nodeCount, ok := stats["nodes"]; !ok || nodeCount.(int) <= 0 {
		t.Errorf("Expected positive node count, got %v", nodeCount)
	}

	if routeCount, ok := stats["routes"]; !ok || routeCount.(int) != 5 {
		t.Errorf("Expected route count of 5, got %v", routeCount)
	}

	if maxDepth, ok := stats["max_depth"]; !ok || maxDepth.(int) <= 0 {
		t.Errorf("Expected positive max depth, got %v", maxDepth)
	}

	t.Logf("Router stats: %+v", stats)
}

func TestRouterServeHTTP(t *testing.T) {
	rb := NewRouterBuilder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	err := rb.AddRoute("GET", "/test", handler)
	if err != nil {
		t.Fatalf("Error adding route: %v", err)
	}

	router, err := rb.Build()
	if err != nil {
		t.Fatalf("Error building router: %v", err)
	}

	// Test found route
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("Expected response 'OK', got '%s'", w.Body.String())
	}

	// Test not found route
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/notfound", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestCannotModifyBuiltRouter(t *testing.T) {
	rb := NewRouterBuilder()

	err := rb.AddRoute("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Error adding route: %v", err)
	}

	_, err = rb.Build()
	if err != nil {
		t.Fatalf("Error building router: %v", err)
	}

	// Should not be able to add routes after building
	err = rb.AddRoute("GET", "/another", nil)
	if err == nil {
		t.Error("Expected error when adding route to built router")
	}

	// Should not be able to build again
	_, err = rb.Build()
	if err == nil {
		t.Error("Expected error when building router twice")
	}
}

// Performance benchmarks
func BenchmarkRouter_StaticRoute(b *testing.B) {
	rb := NewRouterBuilder()
	rb.AddRoute("GET", "/api/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	router, _ := rb.Build()
	
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRouter_ParamRoute(b *testing.B) {
	rb := NewRouterBuilder()
	rb.AddRoute("GET", "/api/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	router, _ := rb.Build()
	
	req := httptest.NewRequest("GET", "/api/users/123", nil)
	w := httptest.NewRecorder()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRouter_MatchOnly(b *testing.B) {
	rb := NewRouterBuilder()
	rb.AddRoute("GET", "/api/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rb.AddRoute("GET", "/api/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rb.AddRoute("GET", "/api/users/:id/posts", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	router, _ := rb.Build()
	
	paths := []string{"/api/users", "/api/users/123", "/api/users/123/posts"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := paths[i%len(paths)]
		handler, params := router.Match("GET", path)
		_ = handler
		_ = params
	}
}

func BenchmarkRouter_MatchOnlyOptimized(b *testing.B) {
	rb := NewRouterBuilder()
	rb.AddRoute("GET", "/api/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rb.AddRoute("GET", "/api/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rb.AddRoute("GET", "/api/users/:id/posts", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	router, _ := rb.Build()
	
	paths := []string{"/api/users", "/api/users/123", "/api/users/123/posts"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := paths[i%len(paths)]
		handler, params := router.MatchOptimized("GET", path)
		_ = handler
		_ = params
	}
}

func BenchmarkRouter_MatchOnlyOptimized2(b *testing.B) {
	rb := NewRouterBuilder()
	rb.AddRoute("GET", "/api/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rb.AddRoute("GET", "/api/posts", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rb.AddRoute("GET", "/api/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rb.AddRoute("GET", "/api/users/:id/posts", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	router, _ := rb.Build()
	
	paths := []string{"/api/users", "/api/posts", "/api/users/123", "/api/users/123/posts"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := paths[i%len(paths)]
		handler, params := router.MatchOptimized2("GET", path)
		_ = handler
		if params != nil && len(params) > 0 {
			// Return params to pool when done (in real usage)
			paramsPool.Put(params)
		}
	}
}

func BenchmarkComparison_Static(b *testing.B) {
	rb := NewRouterBuilder()
	rb.AddRoute("GET", "/api/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rb.AddRoute("GET", "/api/posts", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	router, _ := rb.Build()
	
	b.Run("FastRouter-Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			handler, params := router.Match("GET", "/api/users")
			_ = handler
			_ = params
		}
	})
	
	b.Run("FastRouter-Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			handler, params := router.MatchOptimized2("GET", "/api/users")
			_ = handler
			_ = params
		}
	})
}

func BenchmarkComparison_Params(b *testing.B) {
	rb := NewRouterBuilder()
	rb.AddRoute("GET", "/api/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rb.AddRoute("GET", "/api/users/:id/posts/:post", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	router, _ := rb.Build()
	
	b.Run("FastRouter-Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			handler, params := router.Match("GET", "/api/users/123")
			_ = handler
			_ = params
		}
	})
	
	b.Run("FastRouter-Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			handler, params := router.MatchOptimized2("GET", "/api/users/123")
			_ = handler
			if params != nil {
				paramsPool.Put(params)
			}
		}
	})
}
