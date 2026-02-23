# FastRouter

A high-performance HTTP router for Go with support for static routes, parameters, and wildcards.

## Features

- ✅ **Static routes**: `/api/health`
- ✅ **Parameter routes**: `/users/:id`, `/users/:id/posts/:postId`  
- ✅ **Wildcard routes**: `/page/*`, `/static/*`
- ✅ Fast route matching with trie-based lookup
- ✅ Lexicographic route ordering for optimal performance
- ✅ Path parameter extraction
- ✅ Memory-efficient with parameter pooling

## Quick Start

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/youruser/fastrouter"
)

func main() {
    rb := fastrouter.NewRouterBuilder()

    // Static route
    rb.AddRoute("GET", "/api/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "OK")
    }))

    // Parameter route
    rb.AddRoute("GET", "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract parameters with custom middleware (see DYNAMIC_ROUTES.md)
        fmt.Fprintln(w, "User route")
    }))

    // Wildcard route - matches /page/anything/nested
    rb.AddRoute("GET", "/page/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Wildcard route")
    }))

    router, _ := rb.Build()
    http.ListenAndServe(":8080", router)
}
```

## Dynamic Routes

FastRouter supports powerful dynamic routing patterns:

- **Parameters**: `/users/:id` matches `/users/123` 
- **Wildcards**: `/page/*` matches `/page/me/something`

See [DYNAMIC_ROUTES.md](DYNAMIC_ROUTES.md) for complete documentation and examples.

## Examples

Run the dynamic routes example:
```bash
go run example_server.go
```

Visit `http://localhost:8080/page/me/something` to see wildcard routing in action!

## Testing

```bash
go test -v
```

## Performance

See [PERFORMANCE_ASSESSMENT.md](PERFORMANCE_ASSESSMENT.md) for detailed benchmarks.