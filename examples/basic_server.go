package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jamra/fastrouter"
)

func main() {
	builder := fastrouter.NewRouterBuilder()
	
	builder.AddRoute("GET", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to FastRouter Example!")
	}))
	builder.AddRoute("GET", "/api/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"users": [{"id":1,"name":"Alice"}]}`)
	}))
	
	router, err := builder.Build()
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("Server starting on :8080")
	fmt.Println("Try: http://localhost:8080/")
	fmt.Println("Try: http://localhost:8080/api/users")
	log.Fatal(http.ListenAndServe(":8080", router))
}
