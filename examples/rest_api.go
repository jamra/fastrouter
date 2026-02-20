package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jamra/fastrouter"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var users = map[int]User{
	1: {ID: 1, Name: "Alice"},
	2: {ID: 2, Name: "Bob"},
}

func main() {
	builder := fastrouter.NewRouterBuilder()
	
	// Routes in lexicographic order (FST requirement)
	builder.AddRoute("GET", "/", http.HandlerFunc(homeHandler))
	builder.AddRoute("GET", "/api/users", http.HandlerFunc(listUsersHandler))
	builder.AddRoute("GET", "/api/users/:id", http.HandlerFunc(getUserHandler))
	builder.AddRoute("POST", "/api/users", http.HandlerFunc(createUserHandler))
	
	router, err := builder.Build()
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("REST API Router built with %d routes\n", router.RouteCount())
	fmt.Println("Server starting on :8080")
	fmt.Println("Try:")
	fmt.Println("  GET  http://localhost:8080/")
	fmt.Println("  GET  http://localhost:8080/api/users")
	fmt.Println("  GET  http://localhost:8080/api/users/1")
	fmt.Println("  POST http://localhost:8080/api/users")
	
	log.Fatal(http.ListenAndServe(":8080", router))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<h1>FastRouter REST API</h1>
<p>FST-inspired HTTP routing demonstration</p>
<ul>
  <li><a href="/api/users">GET /api/users</a></li>
  <li><a href="/api/users/1">GET /api/users/1</a></li>
</ul>
	`)
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userList := make([]User, 0, len(users))
	for _, user := range users {
		userList = append(userList, user)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": userList,
		"count": len(userList),
	})
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	params := fastrouter.GetPathParams(r)
	idStr := params["id"]
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", 400)
		return
	}
	
	user, exists := users[id]
	if !exists {
		http.Error(w, "User not found", 404)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	
	maxID := 0
	for id := range users {
		if id > maxID {
			maxID = id
		}
	}
	user.ID = maxID + 1
	users[user.ID] = user
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(user)
}
