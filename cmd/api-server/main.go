package main

import (
	"encoding/json"
	"log"
	"net/http"

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
	return mux
}

func (a *api) handleCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(a.store.ListCategories())
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
