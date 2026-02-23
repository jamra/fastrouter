package fastrouter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDynamicRoutes(t *testing.T) {
	// Create router with dynamic routes
	rb := NewRouterBuilder()

	// Add routes in lexicographic order
	routes := []struct {
		method string
		path   string
		name   string
	}{
		{"GET", "/", "root"},
		{"GET", "/api/health", "health"},
		{"GET", "/api/users/:id", "user"},
		{"GET", "/api/users/:id/posts", "user-posts"},
		{"GET", "/api/users/:id/posts/:postId", "user-post"},
		{"GET", "/page/*", "page-wildcard"},
		{"GET", "/static/*", "static-files"},
	}

	// Helper to create test handlers
	testHandler := func(name string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(name))
		})
	}

	// Add all routes
	for _, route := range routes {
		err := rb.AddRoute(route.method, route.path, testHandler(route.name))
		if err != nil {
			t.Fatalf("Error adding route %s %s: %v", route.method, route.path, err)
		}
	}

	// Build router
	router, err := rb.Build()
	if err != nil {
		t.Fatalf("Error building router: %v", err)
	}

	// Test cases for dynamic routes
	testCases := []struct {
		name           string
		method         string
		path           string
		expectFound    bool
		expectedResp   string
		expectedParams map[string]string
	}{
		// Static routes
		{
			name:         "Root route",
			method:       "GET",
			path:         "/",
			expectFound:  true,
			expectedResp: "root",
			expectedParams: map[string]string{},
		},
		{
			name:         "Health check",
			method:       "GET",
			path:         "/api/health",
			expectFound:  true,
			expectedResp: "health",
			expectedParams: map[string]string{},
		},
		
		// Parameter routes
		{
			name:         "User by ID",
			method:       "GET",
			path:         "/api/users/123",
			expectFound:  true,
			expectedResp: "user",
			expectedParams: map[string]string{"id": "123"},
		},
		{
			name:         "User posts",
			method:       "GET",
			path:         "/api/users/abc/posts",
			expectFound:  true,
			expectedResp: "user-posts",
			expectedParams: map[string]string{"id": "abc"},
		},
		{
			name:         "Specific user post",
			method:       "GET",
			path:         "/api/users/123/posts/456",
			expectFound:  true,
			expectedResp: "user-post",
			expectedParams: map[string]string{"id": "123", "postId": "456"},
		},
		
		// Wildcard routes - your main request!
		{
			name:         "Page wildcard - your example",
			method:       "GET",
			path:         "/page/me/something",
			expectFound:  true,
			expectedResp: "page-wildcard",
			expectedParams: map[string]string{"*": "me/something"},
		},
		{
			name:         "Page wildcard - admin dashboard",
			method:       "GET",
			path:         "/page/admin/dashboard/settings",
			expectFound:  true,
			expectedResp: "page-wildcard",
			expectedParams: map[string]string{"*": "admin/dashboard/settings"},
		},
		{
			name:         "Page wildcard - single segment",
			method:       "GET",
			path:         "/page/home",
			expectFound:  true,
			expectedResp: "page-wildcard",
			expectedParams: map[string]string{"*": "home"},
		},
		{
			name:         "Static files wildcard",
			method:       "GET",
			path:         "/static/css/style.css",
			expectFound:  true,
			expectedResp: "static-files",
			expectedParams: map[string]string{"*": "css/style.css"},
		},
		{
			name:         "Static files - deep path",
			method:       "GET",
			path:         "/static/images/icons/favicon.ico",
			expectFound:  true,
			expectedResp: "static-files",
			expectedParams: map[string]string{"*": "images/icons/favicon.ico"},
		},
		
		// Non-matching routes
		{
			name:         "Non-existent route",
			method:       "GET",
			path:         "/nonexistent",
			expectFound:  false,
			expectedResp: "",
			expectedParams: nil,
		},
		{
			name:         "Wrong method",
			method:       "POST",
			path:         "/api/health",
			expectFound:  false,
			expectedResp: "",
			expectedParams: nil,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler, params := router.Match(tc.method, tc.path)

			if tc.expectFound {
				if handler == nil {
					t.Errorf("Expected handler for %s %s, got nil", tc.method, tc.path)
					return
				}

				// Test handler response
				w := httptest.NewRecorder()
				r := httptest.NewRequest(tc.method, tc.path, nil)
				handler.ServeHTTP(w, r)

				if w.Body.String() != tc.expectedResp {
					t.Errorf("Expected response '%s' for %s %s, got '%s'",
						tc.expectedResp, tc.method, tc.path, w.Body.String())
				}

				// Check parameters
				for key, expected := range tc.expectedParams {
					if actual, exists := params[key]; !exists || actual != expected {
						t.Errorf("Expected param %s=%s for %s %s, got %s=%s",
							key, expected, tc.method, tc.path, key, actual)
					}
				}

				// Verify no unexpected parameters
				for key := range params {
					if _, expected := tc.expectedParams[key]; !expected {
						t.Errorf("Unexpected param %s=%s for %s %s",
							key, params[key], tc.method, tc.path)
					}
				}
			} else {
				if handler != nil {
					t.Errorf("Expected no handler for %s %s, got one", tc.method, tc.path)
				}
			}
		})
	}
}