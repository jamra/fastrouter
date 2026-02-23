package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/example/fastrouter"
)

// PathParamsKey is a context key for storing path parameters
type contextKey struct{}

var pathParamsKey = &contextKey{}

// Middleware to store path parameters in request context
func withPathParams(handler http.Handler, params fastrouter.PathParams) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), pathParamsKey, params)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetPathParams extracts path parameters from request context
func getPathParams(r *http.Request) fastrouter.PathParams {
	if params, ok := r.Context().Value(pathParamsKey).(fastrouter.PathParams); ok {
		return params
	}
	return make(fastrouter.PathParams)
}

// Enhanced router that handles path parameters
type EnhancedRouter struct {
	*fastrouter.Router
}

func (er *EnhancedRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, params := er.Router.Match(r.Method, r.URL.Path)
	if handler != nil {
		// Wrap handler with path parameters
		withPathParams(handler, params).ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}

func main() {
	// Create router builder
	rb := fastrouter.NewRouterBuilder()

	// Root route with navigation
	rb.AddRoute("GET", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Dynamic FastRouter Demo</title></head>
<body>
<h1>🚀 Dynamic FastRouter Demo</h1>
<p>Your wildcard route <code>/page/*</code> is working!</p>

<h2>Test Routes:</h2>
<ul>
<li><a href="/api/users/123">/api/users/123</a> - Parameter route</li>
<li><a href="/api/users/123/posts">/api/users/123/posts</a> - Nested parameters</li>
<li><a href="/api/users/123/posts/456">/api/users/123/posts/456</a> - Multiple parameters</li>
<li><a href="/page/me/something"><strong>/page/me/something</strong></a> - <strong>Your wildcard example!</strong></li>
<li><a href="/page/admin/dashboard">/page/admin/dashboard</a> - Another wildcard</li>
<li><a href="/static/css/style.css">/static/css/style.css</a> - File serving wildcard</li>
</ul>

<h2>Router Features:</h2>
<ul>
<li>✅ Static routes</li>
<li>✅ Parameter routes with <code>:param</code></li>
<li>✅ Wildcard routes with <code>*</code></li>
<li>✅ Fast trie-based matching</li>
<li>✅ Path parameter extraction</li>
</ul>
</body>
</html>
		`)
	}))

	// User routes with parameters
	rb.AddRoute("GET", "/api/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := getPathParams(r)
		userID := params["id"]
		fmt.Fprintf(w, "✅ Parameter route matched!\n\nUser ID: %s\nFull path: %s\n\nParameters: %v", 
			userID, r.URL.Path, params)
	}))

	rb.AddRoute("GET", "/api/users/:id/posts", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := getPathParams(r)
		userID := params["id"]
		fmt.Fprintf(w, "✅ Nested parameter route matched!\n\nPosts for user: %s\nFull path: %s\n\nParameters: %v", 
			userID, r.URL.Path, params)
	}))

	rb.AddRoute("GET", "/api/users/:id/posts/:postId", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := getPathParams(r)
		userID := params["id"]
		postID := params["postId"]
		fmt.Fprintf(w, "✅ Multiple parameters matched!\n\nUser: %s\nPost: %s\nFull path: %s\n\nParameters: %v", 
			userID, postID, r.URL.Path, params)
	}))

	// Page wildcard route - your main request!
	rb.AddRoute("GET", "/page/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := getPathParams(r)
		subPath := params["*"]
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Wildcard Route Success!</title></head>
<body>
<h1>🎉 Wildcard Route Matched Successfully!</h1>
<p><strong>Your example works perfectly!</strong></p>

<div style="background: #f0f8ff; padding: 20px; border-radius: 5px; margin: 20px 0;">
<h2>Route Details:</h2>
<p><strong>Route pattern:</strong> <code>/page/*</code></p>
<p><strong>Request path:</strong> <code>%s</code></p>
<p><strong>Captured wildcard:</strong> <code>%s</code></p>
</div>

<h2>How it works:</h2>
<ul>
<li>The <code>/page/*</code> pattern matches any path starting with <code>/page/</code></li>
<li>Everything after <code>/page/</code> is captured in the <code>*</code> parameter</li>
<li>Works with any depth: <code>/page/a/b/c/d/e</code></li>
</ul>

<p><a href="/">← Back to examples</a></p>
</body>
</html>
		`, r.URL.Path, subPath)
	}))

	// File serving wildcard
	rb.AddRoute("GET", "/static/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := getPathParams(r)
		filePath := params["*"]
		fmt.Fprintf(w, "✅ Static file wildcard matched!\n\nWould serve file: %s\nFull path: %s\n\n(In a real app, you'd serve the actual file here)", 
			filePath, r.URL.Path)
	}))

	// Build the router
	router, err := rb.Build()
	if err != nil {
		log.Fatal("Failed to build router:", err)
	}

	// Wrap in enhanced router for parameter handling
	enhancedRouter := &EnhancedRouter{Router: router}

	fmt.Println("🚀 Dynamic FastRouter server starting on http://localhost:8080")
	fmt.Println("")
	fmt.Println("✨ Your wildcard route /page/* is working!")
	fmt.Println("   Try: http://localhost:8080/page/me/something")
	fmt.Println("")
	fmt.Println("📊 Router built successfully:")
	stats := router.Stats()
	fmt.Printf("   - %d nodes in trie structure\n", stats["nodes"])
	fmt.Printf("   - %d total routes\n", stats["routes"])
	fmt.Printf("   - %d max depth\n", stats["max_depth"])
	fmt.Println("")
	fmt.Println("Open http://localhost:8080 in your browser to test!")

	log.Fatal(http.ListenAndServe(":8080", enhancedRouter))
}