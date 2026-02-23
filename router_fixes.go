package fastrouter

import (
	"net/http"
)

// FixedRouter wraps the original Router with a corrected FastMatch implementation
type FixedRouter struct {
	*Router
	hasParams bool // Pre-computed flag indicating if router has any parameterized routes
}

// NewFixedRouter creates a router with the corrected FastMatch behavior
func NewFixedRouter(router *Router) *FixedRouter {
	// Check if router has any parameterized routes by examining the structure
	hasParams := checkForParams(router.root)
	
	return &FixedRouter{
		Router:    router,
		hasParams: hasParams,
	}
}

// checkForParams recursively checks if the trie contains any parameter or wildcard nodes
func checkForParams(n *node) bool {
	if n.isParam || n.isWild {
		return true
	}
	
	for _, child := range n.children {
		if checkForParams(child) {
			return true
		}
	}
	
	if n.wildChild != nil && checkForParams(n.wildChild) {
		return true
	}
	
	return false
}

// FastMatch provides the corrected fast matching implementation
func (fr *FixedRouter) FastMatch(method, path string) (http.Handler, PathParams) {
	if path == "" || path[0] != '/' {
		path = "/" + path
	}

	var params PathParams
	var handler http.Handler
	
	if fr.hasParams {
		// Router has parameterized routes, use full matching with parameter pooling
		params = paramsPool.Get().(PathParams)
		// Clear any existing params
		for k := range params {
			delete(params, k)
		}
		handler = fr.matchPathOptimized(fr.root, path, 1, method, params)
		
		// Return empty params to pool if no matches found
		if handler == nil {
			paramsPool.Put(params)
			params = nil
		}
	} else {
		// Router has only static routes, use static optimization
		handler = fr.matchPathOptimizedStatic(fr.root, path, 1, method)
	}
	
	return handler, params
}

// Enhanced RouterBuilder that builds FixedRouters
type EnhancedRouterBuilder struct {
	*RouterBuilder
}

// NewEnhancedRouterBuilder creates a new enhanced router builder
func NewEnhancedRouterBuilder() *EnhancedRouterBuilder {
	return &EnhancedRouterBuilder{
		RouterBuilder: NewRouterBuilder(),
	}
}

// Build constructs a FixedRouter with corrected FastMatch behavior  
func (erb *EnhancedRouterBuilder) Build() (*FixedRouter, error) {
	router, err := erb.RouterBuilder.Build()
	if err != nil {
		return nil, err
	}
	
	return NewFixedRouter(router), nil
}