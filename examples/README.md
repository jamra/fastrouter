# FastRouter Examples

Real HTTP server examples demonstrating FastRouter's FST-inspired design.

## 🚀 Quick Start

### Basic Server
```bash
cd examples
go run basic_server.go
# Visit: http://localhost:8080/
```

### REST API Server  
```bash
cd examples
go run rest_api.go
# Visit: http://localhost:8080/api/users
```

## 🏗️ FST Design Principles

1. **Lexicographic Order Required**: Routes must be added alphabetically
2. **Immutable After Build**: No runtime route changes
3. **O(Path Length) Performance**: Speed independent of route count
4. **Memory Efficient**: Structure sharing for common prefixes

## 📋 Example Features

- ✅ Path parameters (`:id`)
- ✅ Wildcard matching (`*`)
- ✅ Multiple HTTP methods
- ✅ JSON APIs
- ✅ Parameter extraction
- ✅ Performance metrics

## 💡 Usage

```go
// 1. Create builder
builder := fastrouter.NewRouterBuilder()

// 2. Add routes in lexicographic order  
builder.AddRoute("GET", "/api/users", handler)
builder.AddRoute("GET", "/api/users/:id", handler)

// 3. Build immutable router
router, _ := builder.Build()

// 4. Extract path params in handlers
func handler(w http.ResponseWriter, r *http.Request) {
    params := fastrouter.GetPathParams(r)
    userID := params["id"]
}

// 5. Start server
http.ListenAndServe(":8080", router)
```

**Inspired by:** [Finite State Transducers](https://burntsushi.net/transducers/)
