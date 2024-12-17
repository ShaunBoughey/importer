// cmd/mockapi/main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"importer/models"
)

type MockAPI struct {
	customers        map[string]int // CustomerNumber to ID
	accounts         map[string]int // AccountNumber to ID
	customerAccounts []models.CustomerAccountLinkRequest
	nextID           int
	mu               sync.Mutex
}

func NewMockAPI() *MockAPI {
	return &MockAPI{
		customers:        make(map[string]int),
		accounts:         make(map[string]int),
		customerAccounts: make([]models.CustomerAccountLinkRequest, 0),
		nextID:           1,
	}
}

func main() {
	port := flag.Int("port", 3000, "Port to run mock API on")
	flag.Parse()

	api := NewMockAPI()

	// Customer endpoints
	http.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req models.CustomerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		api.mu.Lock()
		id := api.nextID
		api.nextID++
		api.customers[req.CustomerNumber] = id
		api.mu.Unlock()

		log.Printf("Created customer %s with ID %d", req.CustomerNumber, id)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": id})
	})

	// Account endpoints
	http.HandleFunc("/accounts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req models.AccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		api.mu.Lock()
		id := api.nextID
		api.nextID++
		api.accounts[req.AccountNumber] = id
		api.mu.Unlock()

		log.Printf("Created account %s with ID %d", req.AccountNumber, id)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": id})
	})

	// Customer-Account link endpoints
	http.HandleFunc("/customer-accounts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req models.CustomerAccountLinkRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		api.mu.Lock()
		api.customerAccounts = append(api.customerAccounts, req)
		api.mu.Unlock()

		log.Printf("Created link between customer %d and account %d", req.CustomerID, req.AccountID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	// Stats endpoint
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		api.mu.Lock()
		stats := map[string]int{
			"customers": len(api.customers),
			"accounts":  len(api.accounts),
			"links":     len(api.customerAccounts),
		}
		api.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting mock API server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
