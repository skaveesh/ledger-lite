package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/skaveesh/ledger-lite/internal/domain"
	"github.com/skaveesh/ledger-lite/internal/store"
)

type api struct {
	store store.Store
}

func newAPI() *api {
	return &api{store: store.NewMemoryStore()}
}

func (a *api) router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("LedgerLite"))
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/categories", a.handleCategories)
	mux.HandleFunc("/transactions", a.handleTransactions)
	mux.HandleFunc("/transactions/", a.handleTransactionByID)
	return mux
}

func (a *api) handleCategories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(a.store.ListCategories())
	case http.MethodPost:
		var req domain.Category
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		created, err := a.store.CreateCategory(domain.Category{Name: req.Name})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(created)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *api) handleTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(a.store.ListTransactions())
	case http.MethodPost:
		var req domain.Transaction
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		created, err := a.store.CreateTransaction(domain.Transaction{
			CategoryID:  req.CategoryID,
			AmountCents: req.AmountCents,
			Description: req.Description,
			Date:        req.Date,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(created)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *api) handleTransactionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idPart := strings.TrimPrefix(r.URL.Path, "/transactions/")
	id, err := strconv.ParseInt(idPart, 10, 64)
	if err != nil {
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	item, ok := a.store.GetTransaction(id)
	if !ok {
		http.Error(w, "transaction not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

func main() {
	app := newAPI()
	server := &http.Server{
		Addr:    ":8080",
		Handler: app.router(),
	}

	log.Println("api server listening on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
