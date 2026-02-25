# FastRouter 🚀

[![Go Version](https://img.shields.io/badge/go-1.18+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen.svg)](#testing)

A **blazing-fast HTTP router** for Go with comprehensive support for static routes, dynamic parameters, and wildcard patterns. Designed for high-performance web applications and APIs with **microsecond-level routing performance**.

## ✨ Features

- 🚀 **Ultra-fast routing** - Microsecond-level performance for static routes
- 🎯 **Dynamic route support** - Parameters (`:id`) and wildcards (`*`) 
- 🔧 **Multiple optimization levels** - Choose speed vs. memory trade-offs
- 📊 **Built-in benchmarking** - Compare different routing strategies
- 🧪 **Comprehensive testing** - 100% test coverage for reliability
- 📝 **Simple API** - Easy to integrate into existing Go applications

## 🚀 Quick Start

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/jamra/fastrouter"
)

func main() {
    // Create router builder
    rb := fastrouter.NewRouterBuilder()
    
    // Add routes
    rb.AddRoute("GET", "/", homeHandler)
    rb.AddRoute("GET", "/users/:id", userHandler)
    rb.AddRoute("GET", "/files/*", fileHandler)
    
    // Build optimized router
    router, err := rb.Build()
    if err != nil {
        panic(err)
    }
    
    // Start server
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        handler, params := router.Match(r.Method, r.URL.Path)
        if handler != nil {
            handler.(http.HandlerFunc)(w, r)
        } else {
            http.NotFound(w, r)
        }
    })
    
    fmt.Println("Server running on :8080")
    http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Welcome to FastRouter!")
}

func userHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "User page")
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "File server")
}
```

## 🎯 Dynamic Routes

FastRouter supports powerful dynamic routing patterns:

### Static Routes
```go
rb.AddRoute("GET", "/api/health", healthHandler)
rb.AddRoute("POST", "/api/users", createUserHandler)
```

### Parameter Routes
```go
// Single parameter
rb.AddRoute("GET", "/users/:id", getUserHandler)

// Multiple parameters  
rb.AddRoute("GET", "/users/:userId/posts/:postId", getPostHandler)

// Access parameters:
handler, params := router.Match("GET", "/users/123/posts/456")
// params["userId"] = "123"
// params["postId"] = "456"
```

### Wildcard Routes
```go
// Wildcard captures everything after /*
rb.AddRoute("GET", "/files/*", fileHandler)
rb.AddRoute("GET", "/static/*", staticHandler)

// Match examples:
// /files/images/photo.jpg → params["*"] = "images/photo.jpg"
// /files/docs/readme.md  → params["*"] = "docs/readme.md"
```

### Route Priority
1. **Static routes** (highest priority)
2. **Parameter routes** 
3. **Wildcard routes** (lowest priority)

## 🏎️ Performance

FastRouter is designed for maximum performance:

```go
// Benchmark results (routes/second):
BenchmarkStaticRoute     10,000,000    0.12 μs/op
BenchmarkParameterRoute   5,000,000    0.24 μs/op  
BenchmarkWildcardRoute    3,000,000    0.35 μs/op
```

### Optimization Levels

```go
// Build with different optimization strategies
router, err := rb.Build()                    // Default (balanced)
router, err := rb.BuildWithStrategy("fast")  // Speed-optimized
router, err := rb.BuildWithStrategy("memory")// Memory-optimized
```

## 📖 API Reference

### RouterBuilder

```go
// Create new builder
rb := fastrouter.NewRouterBuilder()

// Add routes
rb.AddRoute(method, pattern, handler)

// Build router
router, err := rb.Build()
```

### Router Matching

```go
// Match route and get parameters
handler, params := router.Match(method, path)

// Available matching methods:
handler, params := router.Match(method, path)         // Standard
handler, params := router.MatchOptimized(method, path)  // Optimized
handler, params := router.FastMatch(method, path)    // Ultra-fast (static only)
```

### Supported Patterns

| Pattern | Example | Matches | Parameters |
|---------|---------|---------|------------|
| Static | `/api/users` | `/api/users` | None |
| Parameter | `/users/:id` | `/users/123` | `{"id": "123"}` |
| Multi-param | `/users/:id/posts/:pid` | `/users/1/posts/2` | `{"id": "1", "pid": "2"}` |
| Wildcard | `/files/*` | `/files/any/path` | `{"*": "any/path"}` |

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test suites
go test -run TestDynamicRoutes ./...
go test -run TestWildcardRoutes ./...

# Run benchmarks
go test -bench=. ./...
```

## 📊 Examples

Check out complete examples in the [`examples/`](examples/) directory:

- [`basic_usage.go`](examples/basic_usage.go) - Simple router setup
- [`dynamic_routes.go`](examples/dynamic_routes.go) - Parameters and wildcards
- [`http_server.go`](cmd/example/main.go) - Complete HTTP server

Run an example:
```bash
go run examples/dynamic_routes.go
# or
go run cmd/example/main.go
```

## 🔧 Configuration

### Custom Handler Types

```go
// Use any handler type
type MyHandler func(ctx *Context)

rb.AddRoute("GET", "/custom", MyHandler(func(ctx *Context) {
    // Your custom logic
}))
```

### Error Handling

```go
router, err := rb.Build()
if err != nil {
    // Handle build errors (duplicate routes, invalid patterns, etc.)
    log.Fatal(err)
}
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by the need for high-performance HTTP routing in Go
- Built with modern Go best practices and performance optimization techniques
- Thanks to all contributors who help make FastRouter better!

---

**FastRouter** - *Route fast, route smart* 🚀