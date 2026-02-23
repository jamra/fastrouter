# Dynamic Routes in FastRouter

FastRouter supports dynamic routes through **parameters** and **wildcards**, enabling you to handle complex routing patterns efficiently.

## Route Types

### 1. Static Routes
```go
rb.AddRoute("GET", "/api/health", handler)
```
Matches exactly `/api/health`

### 2. Parameter Routes  
```go
rb.AddRoute("GET", "/api/users/:id", handler)
rb.AddRoute("GET", "/api/users/:id/posts/:postId", handler)
```
- `:id` captures a single path segment
- Route `/api/users/:id` matches `/api/users/123`, `/api/users/john`, etc.
- Captured values are available in `PathParams`

### 3. Wildcard Routes (Your Request!)
```go
rb.AddRoute("GET", "/page/*", handler)
```
- `*` captures everything after the prefix
- Route `/page/*` matches:
  - `/page/me/something` → captures `me/something`
  - `/page/admin/dashboard/settings` → captures `admin/dashboard/settings`
  - `/page/home` → captures `home`

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    fastrouter "github.com/youruser/fastrouter"
)

// Setup for parameter passing
type contextKey struct{}
var pathParamsKey = &contextKey{}

func withPathParams(handler http.Handler, params fastrouter.PathParams) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := context.WithValue(r.Context(), pathParamsKey, params)
        handler.ServeHTTP(w, r.WithContext(ctx))
    })
}

func getPathParams(r *http.Request) fastrouter.PathParams {
    if params, ok := r.Context().Value(pathParamsKey).(fastrouter.PathParams); ok {
        return params
    }
    return make(fastrouter.PathParams)
}

type EnhancedRouter struct {
    *fastrouter.Router
}

func (er *EnhancedRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    handler, params := er.Router.Match(r.Method, r.URL.Path)
    if handler != nil {
        withPathParams(handler, params).ServeHTTP(w, r)
    } else {
        http.NotFound(w, r)
    }
}

func main() {
    rb := fastrouter.NewRouterBuilder()

    // Parameter route
    rb.AddRoute("GET", "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        params := getPathParams(r)
        userID := params["id"]
        fmt.Fprintf(w, "User ID: %s", userID)
    }))

    // Wildcard route - your exact example!
    rb.AddRoute("GET", "/page/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        params := getPathParams(r)
        subPath := params["*"]
        fmt.Fprintf(w, "Wildcard matched! Sub-path: %s", subPath)
    }))

    router, _ := rb.Build()
    enhancedRouter := &EnhancedRouter{Router: router}

    http.ListenAndServe(":8080", enhancedRouter)
}
```

## Route Matching Examples

| Route Pattern | Request Path | Match? | Parameters |
|---------------|--------------|---------|------------|
| `/page/*` | `/page/me/something` | ✅ | `*: "me/something"` |
| `/page/*` | `/page/admin/dashboard` | ✅ | `*: "admin/dashboard"` |
| `/users/:id` | `/users/123` | ✅ | `id: "123"` |
| `/users/:id/posts/:postId` | `/users/123/posts/456` | ✅ | `id: "123"`, `postId: "456"` |
| `/api/health` | `/api/health` | ✅ | (none) |
| `/page/*` | `/other/path` | ❌ | - |

## Important Notes

### Route Order
Routes must be added in **lexicographic order** for optimal performance:
```go
rb.AddRoute("GET", "/api/health", handler)    // ✅ First
rb.AddRoute("GET", "/api/users/:id", handler) // ✅ After /api/health 
rb.AddRoute("GET", "/page/*", handler)        // ✅ After /api/users/:id
```

### Priority
1. **Static routes** have highest priority
2. **Parameter routes** have medium priority  
3. **Wildcard routes** have lowest priority

### Performance
- Use `router.Match()` for full compatibility
- `router.FastMatch()` has a bug with dynamic routes in the current version
- Parameters are extracted efficiently during matching

## Your Specific Use Case

Your request for `/page/*` to match `/page/me/something` works perfectly:

```go
rb.AddRoute("GET", "/page/*", handler)
```

This will:
- ✅ Match `/page/me/something`
- ✅ Match `/page/admin/dashboard`  
- ✅ Match `/page/any/nested/path`
- ✅ Capture everything after `/page/` in the `*` parameter

## Testing

Run the included tests to verify dynamic routing:
```bash
go test -v -run TestDynamicRoutes
```

## Live Demo

Run the example server:
```bash
go run example_server.go
```

Then visit `http://localhost:8080/page/me/something` to see your wildcard route in action!