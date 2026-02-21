package fastrouter

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"strings"
)

// Route represents a single route with method, path, and handler
type Route struct {
	Method  string
	Path    string
	Handler http.Handler
}

// RouterBuilder is used to collect routes before building the final router
type RouterBuilder struct {
	routes []Route
	built  bool
}

// Router represents the built, immutable router with fast lookup
type Router struct {
	root *node
}

// node represents a node in our FST-like trie structure
type node struct {
	segment   string                   // path segment for this node
	methods   map[string]http.Handler  // handlers for different HTTP methods at this exact path
	children  map[string]*node         // child nodes
	paramName string                   // if this is a parameter segment, the parameter name
	isParam   bool                     // true if this node represents a path parameter
	isWild    bool                     // true if this is a wildcard node
	wildChild *node                    // wildcard child node
}

// NewRouterBuilder creates a new router builder
func NewRouterBuilder() *RouterBuilder {
	return &RouterBuilder{
		routes: make([]Route, 0),
		built:  false,
	}
}

// AddRoute adds a route to the builder. Routes must be added in lexicographic order
// of their paths for optimal performance.
func (rb *RouterBuilder) AddRoute(method, path string, handler http.Handler) error {
	if rb.built {
		return fmt.Errorf("cannot add routes to a built router")
	}

	// Validate that routes are being added in lexicographic order
	if len(rb.routes) > 0 {
		lastPath := rb.routes[len(rb.routes)-1].Path
		if path < lastPath {
			return fmt.Errorf("routes must be added in lexicographic order: '%s' comes before '%s'", path, lastPath)
		}
	}

	route := Route{
		Method:  strings.ToUpper(method),
		Path:    path,
		Handler: handler,
	}

	rb.routes = append(rb.routes, route)
	return nil
}

// Build constructs the final immutable router from the collected routes
func (rb *RouterBuilder) Build() (*Router, error) {
	if rb.built {
		return nil, fmt.Errorf("router already built")
	}

	rb.built = true

	// Sort routes by path to ensure lexicographic order
	sort.Slice(rb.routes, func(i, j int) bool {
		return rb.routes[i].Path < rb.routes[j].Path
	})

	router := &Router{
		root: &node{
			segment:  "",
			methods:  make(map[string]http.Handler),
			children: make(map[string]*node),
		},
	}

	// Build the trie structure
	for _, route := range rb.routes {
		router.addRoute(route)
	}

	return router, nil
}

// addRoute adds a single route to the router's trie structure
func (r *Router) addRoute(route Route) {
	path := route.Path
	if path == "" || path[0] != '/' {
		path = "/" + path
	}

	segments := strings.Split(path, "/")[1:] // Skip empty first element
	if len(segments) == 1 && segments[0] == "" {
		segments = []string{} // Handle root path "/"
	}

	current := r.root
	for i, segment := range segments {
		isLast := i == len(segments)-1

		// Handle path parameters
		if strings.HasPrefix(segment, ":") {
			paramName := segment[1:]
			// Look for existing parameter child
			found := false
			for _, child := range current.children {
				if child.isParam && child.paramName == paramName {
					current = child
					found = true
					break
				}
			}
			if !found {
				newNode := &node{
					segment:   segment,
					methods:   make(map[string]http.Handler),
					children:  make(map[string]*node),
					paramName: paramName,
					isParam:   true,
				}
				current.children[segment] = newNode
				current = newNode
			}
		} else if segment == "*" {
			// Handle wildcard
			if current.wildChild == nil {
				current.wildChild = &node{
					segment:  "*",
					methods:  make(map[string]http.Handler),
					children: make(map[string]*node),
					isWild:   true,
				}
			}
			current = current.wildChild
		} else {
			// Regular segment
			if child, exists := current.children[segment]; exists {
				current = child
			} else {
				newNode := &node{
					segment:  segment,
					methods:  make(map[string]http.Handler),
					children: make(map[string]*node),
				}
				current.children[segment] = newNode
				current = newNode
			}
		}

		// If this is the last segment, add the handler
		if isLast {
			current.methods[route.Method] = route.Handler
		}
	}

	// Handle root path
	if len(segments) == 0 {
		r.root.methods[route.Method] = route.Handler
	}
}

// ServeHTTP implements http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler, params := r.Match(req.Method, req.URL.Path)
	if handler != nil {
		// Store parameters in request context if needed
		if len(params) > 0 {
			// For now, we'll just call the handler directly
			// In a full implementation, you'd want to store params in context
		}
		handler.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}

// PathParams represents extracted path parameters
type PathParams map[string]string

// Match finds a handler for the given method and path
func (r *Router) Match(method, path string) (http.Handler, PathParams) {
	if path == "" || path[0] != '/' {
		path = "/" + path
	}

	segments := strings.Split(path, "/")[1:] // Skip empty first element
	if len(segments) == 1 && segments[0] == "" {
		segments = []string{} // Handle root path "/"
	}

	params := make(PathParams)
	handler := r.matchNode(r.root, segments, method, params)
	return handler, params
}

// matchNode recursively matches path segments against the trie
func (r *Router) matchNode(n *node, segments []string, method string, params PathParams) http.Handler {
	// If we've consumed all segments, check if this node has a handler for the method
	if len(segments) == 0 {
		if handler, exists := n.methods[method]; exists {
			return handler
		}
		return nil
	}

	segment := segments[0]
	remaining := segments[1:]

	// Try exact match first
	if child, exists := n.children[segment]; exists {
		if handler := r.matchNode(child, remaining, method, params); handler != nil {
			return handler
		}
	}

	// Try parameter match
	for _, child := range n.children {
		if child.isParam {
			params[child.paramName] = segment
			if handler := r.matchNode(child, remaining, method, params); handler != nil {
				return handler
			}
			delete(params, child.paramName) // backtrack
		}
	}

	// Try wildcard match
	if n.wildChild != nil {
		// Wildcard matches everything remaining
		params["*"] = strings.Join(segments, "/")
		if handler, exists := n.wildChild.methods[method]; exists {
			return handler
		}
	}

	return nil
}

// Stats returns statistics about the router structure
func (r *Router) Stats() map[string]interface{} {
	nodeCount := 0
	routeCount := 0
	maxDepth := 0

	var countNodes func(*node, int)
	countNodes = func(n *node, depth int) {
		nodeCount++
		if depth > maxDepth {
			maxDepth = depth
		}
		routeCount += len(n.methods)
		
		for _, child := range n.children {
			countNodes(child, depth+1)
		}
		if n.wildChild != nil {
			countNodes(n.wildChild, depth+1)
		}
	}

	countNodes(r.root, 0)

	return map[string]interface{}{
		"nodes":     nodeCount,
		"routes":    routeCount,
		"max_depth": maxDepth,
	}
}

// PathParamsKey is the context key for path parameters
type contextKey struct{}

var pathParamsKey = &contextKey{}


// GetPathParams extracts path parameters from the request context
func GetPathParams(r *http.Request) PathParams {
	if params, ok := r.Context().Value(pathParamsKey).(PathParams); ok {
		return params
	}
	return make(PathParams)
}

// RouteCount returns the total number of routes
func (r *Router) RouteCount() int {
	stats := r.Stats()
	return stats["routes"].(int)
}

// NodeCount returns the total number of nodes in the trie
func (r *Router) NodeCount() int {
	stats := r.Stats()
	return stats["nodes"].(int)
}

// MatchOptimized - optimized version that avoids string allocations
func (r *Router) MatchOptimized(method, path string) (http.Handler, PathParams) {
	if path == "" || path[0] != '/' {
		path = "/" + path
	}

	params := make(PathParams)
	handler := r.matchPathOptimized(r.root, path, 1, method, params)
	return handler, params
}

// matchPathOptimized matches a path without creating string slices
func (r *Router) matchPathOptimized(n *node, path string, start int, method string, params PathParams) http.Handler {
	// If we've consumed the entire path, check for handler
	if start >= len(path) {
		if handler, exists := n.methods[method]; exists {
			return handler
		}
		return nil
	}
	
	// Find the end of current segment
	end := start
	for end < len(path) && path[end] != '/' {
		end++
	}
	
	segment := path[start:end]
	
	// Calculate next segment start position
	nextStart := end + 1
	if nextStart > len(path) {
		nextStart = len(path)
	}
	
	// Try exact match first (most common case)
	if child, exists := n.children[segment]; exists {
		if handler := r.matchPathOptimized(child, path, nextStart, method, params); handler != nil {
			return handler
		}
	}
	
	// Try parameter matches
	for _, child := range n.children {
		if child.isParam {
			params[child.paramName] = segment
			if handler := r.matchPathOptimized(child, path, nextStart, method, params); handler != nil {
				return handler
			}
			delete(params, child.paramName) // backtrack
		}
	}
	
	// Try wildcard match (matches rest of path)
	if n.wildChild != nil {
		if start < len(path) {
			params["*"] = path[start:]
		}
		if handler, exists := n.wildChild.methods[method]; exists {
			return handler
		}
	}
	
	return nil
}


// Pool for PathParams to avoid allocations
var paramsPool = sync.Pool{
	New: func() interface{} {
		return make(PathParams)
	},
}

// MatchOptimized2 - optimized version with parameter pooling
func (r *Router) MatchOptimized2(method, path string) (http.Handler, PathParams) {
	if path == "" || path[0] != '/' {
		path = "/" + path
	}

	// Check if this route could have parameters
	hasParams := false
	for i := 1; i < len(path); i++ {
		if path[i] == ':' || path[i] == '*' {
			hasParams = true
			break
		}
	}
	
	var params PathParams
	var handler http.Handler
	
	if hasParams {
		params = paramsPool.Get().(PathParams)
		// Clear any existing params
		for k := range params {
			delete(params, k)
		}
		handler = r.matchPathOptimized(r.root, path, 1, method, params)
		
		// Return empty params to pool if no matches found
		if handler == nil {
			paramsPool.Put(params)
			params = nil
		}
	} else {
		// For static routes, don't allocate params at all
		handler = r.matchPathOptimizedStatic(r.root, path, 1, method)
	}
	
	return handler, params
}

// matchPathOptimizedStatic - for static routes without parameters
func (r *Router) matchPathOptimizedStatic(n *node, path string, start int, method string) http.Handler {
	// If we've consumed the entire path, check for handler
	if start >= len(path) {
		if handler, exists := n.methods[method]; exists {
			return handler
		}
		return nil
	}
	
	// Find the end of current segment
	end := start
	for end < len(path) && path[end] != '/' {
		end++
	}
	
	segment := path[start:end]
	nextStart := end + 1
	if nextStart > len(path) {
		nextStart = len(path)
	}
	
	// Only try exact matches for static routes
	if child, exists := n.children[segment]; exists && !child.isParam {
		return r.matchPathOptimizedStatic(child, path, nextStart, method)
	}
	
	return nil
}

// ReleaseParams returns parameter map to the pool for reuse
// Call this after processing a request with parameters to optimize memory usage
func ReleaseParams(params PathParams) {
	if params != nil && len(params) > 0 {
		// Clear the map before returning to pool
		for k := range params {
			delete(params, k)
		}
		paramsPool.Put(params)
	}
}

// FastMatch is an alias for the optimized matching function
// This is the recommended method for high-performance applications
func (r *Router) FastMatch(method, path string) (http.Handler, PathParams) {
	return r.MatchOptimized2(method, path)
}
