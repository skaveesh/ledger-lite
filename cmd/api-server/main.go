package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/skaveesh/ledger-lite/internal/domain"
	"github.com/skaveesh/ledger-lite/internal/store"
)

type api struct {
	store store.Store
}

type apiError struct {
	Error string `json:"error"`
}

type uiData struct {
	Title             string
	Transactions      []domain.Transaction
	Categories        []domain.Category
	Budgets           []domain.Budget
	FilterDate        string
	FilterCategoryID  int64
	UIError           string
	SummaryMonth      int
	SummaryYear       int
	SummaryTotalCents int64
	SummaryRows       []summaryRow
}

type summaryRow struct {
	CategoryID int64
	TotalCents int64
}

func newAPI() *api {
	return &api{store: store.NewMemoryStore()}
}

func (a *api) withErrorHandling(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				writeJSON(w, http.StatusInternalServerError, apiError{Error: "internal server error"})
			}
		}()
		next(w, r)
	}
}

func (a *api) router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.withErrorHandling(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("LedgerLite"))
	}))
	mux.HandleFunc("/health", a.withErrorHandling(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	mux.HandleFunc("/categories", a.withErrorHandling(a.handleCategories))
	mux.HandleFunc("/transactions", a.withErrorHandling(a.handleTransactions))
	mux.HandleFunc("/transactions/", a.withErrorHandling(a.handleTransactionByID))
	mux.HandleFunc("/budgets", a.withErrorHandling(a.handleBudgets))
	mux.HandleFunc("/budgets/", a.withErrorHandling(a.handleBudgetByID))
	mux.HandleFunc("/ui", a.withErrorHandling(a.handleUIHome))
	mux.HandleFunc("/ui/transactions", a.withErrorHandling(a.handleUITransactions))
	mux.HandleFunc("/ui/transactions/add", a.withErrorHandling(a.handleUIAddTransaction))
	mux.HandleFunc("/ui/categories", a.withErrorHandling(a.handleUICategories))
	mux.HandleFunc("/ui/budgets", a.withErrorHandling(a.handleUIBudgets))
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
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		if err := validateCategoryInput(req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		created, err := a.store.CreateCategory(domain.Category{Name: strings.TrimSpace(req.Name)})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
	}
}

func (a *api) handleTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, a.store.ListTransactions())
	case http.MethodPost:
		var req domain.Transaction
		if err := decodeJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		if err := validateTransactionInput(req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		created, err := a.store.CreateTransaction(domain.Transaction{
			CategoryID:  req.CategoryID,
			AmountCents: req.AmountCents,
			Description: req.Description,
			Date:        req.Date.UTC().Truncate(time.Second),
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
	}
}

func (a *api) handleTransactionByID(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r.URL.Path, "/transactions/")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid transaction id"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		item, ok := a.store.GetTransaction(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, apiError{Error: "transaction not found"})
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodDelete:
		if !a.store.DeleteTransaction(id) {
			writeJSON(w, http.StatusNotFound, apiError{Error: "transaction not found"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
	}
}

func (a *api) handleBudgets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, a.store.ListBudgets())
	case http.MethodPost:
		var req domain.Budget
		if err := decodeJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		if err := validateBudgetInput(req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		created, err := a.store.CreateBudget(domain.Budget{
			CategoryID:       req.CategoryID,
			Month:            req.Month,
			Year:             req.Year,
			AmountLimitCents: req.AmountLimitCents,
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
	}
}

func (a *api) handleBudgetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
		return
	}

	id, err := parsePathID(r.URL.Path, "/budgets/")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid budget id"})
		return
	}

	var req domain.Budget
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}
	if err := validateBudgetInput(req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}

	updated, ok, err := a.store.UpdateBudget(id, domain.Budget{
		CategoryID:       req.CategoryID,
		Month:            req.Month,
		Year:             req.Year,
		AmountLimitCents: req.AmountLimitCents,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, apiError{Error: "budget not found"})
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (a *api) handleUIHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
		return
	}
	_, _ = w.Write([]byte("LedgerLite UI"))
}

func renderTemplate(w http.ResponseWriter, bodyTemplate string, data uiData) {
	tmpl, err := template.ParseFiles(
		filepath.FromSlash("cmd/api-server/templates/base.html"),
		filepath.FromSlash(bodyTemplate),
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to load template"})
		return
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to render template"})
		return
	}
}

func (a *api) handleUITransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
		return
	}

	filterDate := strings.TrimSpace(r.URL.Query().Get("date"))
	filterCategoryID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("category_id")), 10, 64)
	transactions := a.store.ListTransactions()
	filtered := make([]domain.Transaction, 0, len(transactions))
	for _, tr := range transactions {
		if filterDate != "" && tr.Date.Format("2006-01-02") != filterDate {
			continue
		}
		if filterCategoryID > 0 && tr.CategoryID != filterCategoryID {
			continue
		}
		filtered = append(filtered, tr)
	}

	renderTemplate(w, filepath.FromSlash("cmd/api-server/templates/transactions.html"), uiData{
		Title:            "Transactions",
		Transactions:     filtered,
		Categories:       a.store.ListCategories(),
		FilterDate:       filterDate,
		FilterCategoryID: filterCategoryID,
	})
}

func (a *api) handleUIAddTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
		return
	}

	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid form data"})
		return
	}

	categoryID, err := strconv.ParseInt(r.FormValue("category_id"), 10, 64)
	if err != nil || categoryID <= 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid category_id"})
		return
	}
	amountCents, err := strconv.ParseInt(r.FormValue("amount_cents"), 10, 64)
	if err != nil || amountCents == 0 {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid amount_cents"})
		return
	}

	date, err := time.Parse("2006-01-02", r.FormValue("date"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "invalid date"})
		return
	}

	_, err = a.store.CreateTransaction(domain.Transaction{
		CategoryID:  categoryID,
		AmountCents: amountCents,
		Description: strings.TrimSpace(r.FormValue("description")),
		Date:        date,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}

	http.Redirect(w, r, "/ui/transactions", http.StatusSeeOther)
}

func (a *api) handleUICategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
		return
	}

	renderTemplate(w, filepath.FromSlash("cmd/api-server/templates/categories.html"), uiData{
		Title:      "Categories",
		Categories: a.store.ListCategories(),
	})
}

func (a *api) handleUIBudgets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
		return
	}

	renderTemplate(w, filepath.FromSlash("cmd/api-server/templates/budgets.html"), uiData{
		Title:   "Budgets",
		Budgets: a.store.ListBudgets(),
	})
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
