package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	mux.HandleFunc("/budgets", a.handleBudgets)
	mux.HandleFunc("/budgets/", a.handleBudgetByID)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decodeJSON(r *http.Request, out any) error {
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		return errors.New("invalid JSON body")
	}
	return nil
}

func parsePathID(path string, prefix string) (int64, error) {
	idPart := strings.TrimPrefix(path, prefix)
	if idPart == "" {
		return 0, errors.New("missing id")
	}
	return strconv.ParseInt(idPart, 10, 64)
}

func validateCategoryInput(c domain.Category) error {
	if strings.TrimSpace(c.Name) == "" {
		return errors.New("category name is required")
	}
	return nil
}

func validateTransactionInput(t domain.Transaction) error {
	if t.Date.IsZero() {
		return errors.New("transaction date is required")
	}
	if t.AmountCents == 0 {
		return errors.New("transaction amount must be non-zero")
	}
	return nil
}

func validateBudgetInput(b domain.Budget) error {
	if b.Month < 1 || b.Month > 12 {
		return errors.New("budget month must be between 1 and 12")
	}
	if b.Year < 1 {
		return errors.New("budget year must be greater than 0")
	}
	if b.AmountLimitCents <= 0 {
		return errors.New("budget amount limit must be greater than 0")
	}
	return nil
}

func (a *api) handleCategories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, a.store.ListCategories())
	case http.MethodPost:
		var req domain.Category
		if err := decodeJSON(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := validateCategoryInput(req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		created, err := a.store.CreateCategory(domain.Category{Name: strings.TrimSpace(req.Name)})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *api) handleTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, a.store.ListTransactions())
	case http.MethodPost:
		var req domain.Transaction
		if err := decodeJSON(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := validateTransactionInput(req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		created, err := a.store.CreateTransaction(domain.Transaction{
			CategoryID:  req.CategoryID,
			AmountCents: req.AmountCents,
			Description: req.Description,
			Date:        req.Date.UTC().Truncate(time.Second),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *api) handleTransactionByID(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r.URL.Path, "/transactions/")
	if err != nil {
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		item, ok := a.store.GetTransaction(id)
		if !ok {
			http.Error(w, "transaction not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodDelete:
		if !a.store.DeleteTransaction(id) {
			http.Error(w, "transaction not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *api) handleBudgets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, a.store.ListBudgets())
	case http.MethodPost:
		var req domain.Budget
		if err := decodeJSON(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := validateBudgetInput(req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		created, err := a.store.CreateBudget(domain.Budget{
			CategoryID:       req.CategoryID,
			Month:            req.Month,
			Year:             req.Year,
			AmountLimitCents: req.AmountLimitCents,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *api) handleBudgetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := parsePathID(r.URL.Path, "/budgets/")
	if err != nil {
		http.Error(w, "invalid budget id", http.StatusBadRequest)
		return
	}

	var req domain.Budget
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateBudgetInput(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updated, ok, err := a.store.UpdateBudget(id, domain.Budget{
		CategoryID:       req.CategoryID,
		Month:            req.Month,
		Year:             req.Year,
		AmountLimitCents: req.AmountLimitCents,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !ok {
		http.Error(w, "budget not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, updated)
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
