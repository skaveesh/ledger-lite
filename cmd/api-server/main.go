package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
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

type contextKey string

const requestIDContextKey contextKey = "request_id"

var requestIDCounter uint64
var totalRequests uint64
var totalServerErrors uint64

func nextRequestID() string {
	id := atomic.AddUint64(&requestIDCounter, 1)
	return fmt.Sprintf("req-%d", id)
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if requestID == "" {
			requestID = nextRequestID()
		}
		w.Header().Set("X-Request-ID", requestID)
		r = r.WithContext(context.WithValue(r.Context(), requestIDContextKey, requestID))
		next.ServeHTTP(w, r)
	})
}

func requestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDContextKey).(string)
	return id
}

func requestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		atomic.AddUint64(&totalRequests, 1)
		if rec.status >= 500 {
			atomic.AddUint64(&totalServerErrors, 1)
		}
		log.Printf("request_id=%s %s %s status=%d duration_ms=%d", requestIDFromContext(r.Context()), r.Method, r.URL.Path, rec.status, time.Since(start).Milliseconds())
	})
}

func applyMiddleware(h http.Handler, m ...func(http.Handler) http.Handler) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
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
	mux.HandleFunc("/ui/transactions/delete", a.withErrorHandling(a.handleUIDeleteTransaction))
	mux.HandleFunc("/ui/categories", a.withErrorHandling(a.handleUICategories))
	mux.HandleFunc("/ui/budgets", a.withErrorHandling(a.handleUIBudgets))
	mux.HandleFunc("/ui/summary", a.withErrorHandling(a.handleUIMonthlySummary))
	mux.HandleFunc("/metrics", a.withErrorHandling(a.handleMetrics))
	return applyMiddleware(mux, requestIDMiddleware, requestLoggingMiddleware)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(status int) {
	sr.status = status
	sr.ResponseWriter.WriteHeader(status)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decodeJSON(r *http.Request, out any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	if dec.More() {
		return errors.New("invalid JSON body: multiple JSON values are not allowed")
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

func parsePositiveIntDefault(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func paginateTransactions(items []domain.Transaction, page int, pageSize int) []domain.Transaction {
	if len(items) == 0 {
		return items
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []domain.Transaction{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func sortTransactions(items []domain.Transaction, sortBy string, order string) {
	desc := strings.EqualFold(order, "desc")
	switch sortBy {
	case "amount":
		sort.Slice(items, func(i, j int) bool {
			if desc {
				return items[i].AmountCents > items[j].AmountCents
			}
			return items[i].AmountCents < items[j].AmountCents
		})
	case "date":
		sort.Slice(items, func(i, j int) bool {
			if desc {
				return items[i].Date.After(items[j].Date)
			}
			return items[i].Date.Before(items[j].Date)
		})
	}
}

func (a *api) handleTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		transactions := a.store.ListTransactions()
		sortBy := strings.TrimSpace(r.URL.Query().Get("sort_by"))
		order := strings.TrimSpace(r.URL.Query().Get("order"))
		sortTransactions(transactions, sortBy, order)
		page := parsePositiveIntDefault(r.URL.Query().Get("page"), 1)
		pageSize := parsePositiveIntDefault(r.URL.Query().Get("page_size"), 20)
		writeJSON(w, http.StatusOK, paginateTransactions(transactions, page, pageSize))
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
	renderTemplateStatus(w, http.StatusOK, bodyTemplate, data)
}

func renderTemplateStatus(w http.ResponseWriter, status int, bodyTemplate string, data uiData) {
	tmpl, err := template.ParseFiles(
		filepath.FromSlash("cmd/api-server/templates/base.html"),
		filepath.FromSlash(bodyTemplate),
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to load template"})
		return
	}
	w.WriteHeader(status)
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "failed to render template"})
		return
	}
}

func (a *api) uiTransactionsData(filterDate string, filterCategoryID int64, uiError string) uiData {
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
	return uiData{
		Title:            "Transactions",
		Transactions:     filtered,
		Categories:       a.store.ListCategories(),
		FilterDate:       filterDate,
		FilterCategoryID: filterCategoryID,
		UIError:          uiError,
	}
}

func (a *api) handleUITransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
		return
	}

	filterDate := strings.TrimSpace(r.URL.Query().Get("date"))
	filterCategoryID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("category_id")), 10, 64)

	renderTemplate(w, filepath.FromSlash("cmd/api-server/templates/transactions.html"), a.uiTransactionsData(filterDate, filterCategoryID, ""))
}

func (a *api) handleUIAddTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		renderTemplateStatus(w, http.StatusMethodNotAllowed, filepath.FromSlash("cmd/api-server/templates/transactions.html"), a.uiTransactionsData("", 0, "method not allowed"))
		return
	}

	if err := r.ParseForm(); err != nil {
		renderTemplateStatus(w, http.StatusBadRequest, filepath.FromSlash("cmd/api-server/templates/transactions.html"), a.uiTransactionsData("", 0, "invalid form data"))
		return
	}

	categoryID, err := strconv.ParseInt(r.FormValue("category_id"), 10, 64)
	if err != nil || categoryID <= 0 {
		renderTemplateStatus(w, http.StatusBadRequest, filepath.FromSlash("cmd/api-server/templates/transactions.html"), a.uiTransactionsData("", 0, "invalid category id"))
		return
	}
	amountCents, err := strconv.ParseInt(r.FormValue("amount_cents"), 10, 64)
	if err != nil || amountCents == 0 {
		renderTemplateStatus(w, http.StatusBadRequest, filepath.FromSlash("cmd/api-server/templates/transactions.html"), a.uiTransactionsData("", 0, "invalid amount"))
		return
	}

	date, err := time.Parse("2006-01-02", r.FormValue("date"))
	if err != nil {
		renderTemplateStatus(w, http.StatusBadRequest, filepath.FromSlash("cmd/api-server/templates/transactions.html"), a.uiTransactionsData("", 0, "invalid date"))
		return
	}

	_, err = a.store.CreateTransaction(domain.Transaction{
		CategoryID:  categoryID,
		AmountCents: amountCents,
		Description: strings.TrimSpace(r.FormValue("description")),
		Date:        date,
	})
	if err != nil {
		renderTemplateStatus(w, http.StatusBadRequest, filepath.FromSlash("cmd/api-server/templates/transactions.html"), a.uiTransactionsData("", 0, err.Error()))
		return
	}

	http.Redirect(w, r, "/ui/transactions", http.StatusSeeOther)
}

func (a *api) handleUIDeleteTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		renderTemplateStatus(w, http.StatusMethodNotAllowed, filepath.FromSlash("cmd/api-server/templates/transactions.html"), a.uiTransactionsData("", 0, "method not allowed"))
		return
	}
	if err := r.ParseForm(); err != nil {
		renderTemplateStatus(w, http.StatusBadRequest, filepath.FromSlash("cmd/api-server/templates/transactions.html"), a.uiTransactionsData("", 0, "invalid form data"))
		return
	}

	id, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("id")), 10, 64)
	if err != nil || id <= 0 {
		renderTemplateStatus(w, http.StatusBadRequest, filepath.FromSlash("cmd/api-server/templates/transactions.html"), a.uiTransactionsData("", 0, "invalid transaction id"))
		return
	}
	if !a.store.DeleteTransaction(id) {
		renderTemplateStatus(w, http.StatusNotFound, filepath.FromSlash("cmd/api-server/templates/transactions.html"), a.uiTransactionsData("", 0, "transaction not found"))
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

func (a *api) handleUIMonthlySummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
		return
	}

	now := time.Now()
	month := int(now.Month())
	year := now.Year()
	if m, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("month"))); err == nil && m >= 1 && m <= 12 {
		month = m
	}
	if y, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("year"))); err == nil && y > 0 {
		year = y
	}

	totals := map[int64]int64{}
	var total int64
	for _, tr := range a.store.ListTransactions() {
		if int(tr.Date.Month()) == month && tr.Date.Year() == year {
			totals[tr.CategoryID] += tr.AmountCents
			total += tr.AmountCents
		}
	}

	rows := make([]summaryRow, 0, len(totals))
	for categoryID, cents := range totals {
		rows = append(rows, summaryRow{CategoryID: categoryID, TotalCents: cents})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].CategoryID < rows[j].CategoryID })

	renderTemplate(w, filepath.FromSlash("cmd/api-server/templates/summary.html"), uiData{
		Title:             "Monthly Summary",
		SummaryMonth:      month,
		SummaryYear:       year,
		SummaryTotalCents: total,
		SummaryRows:       rows,
	})
}

func (a *api) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "requests_total %d\n", atomic.LoadUint64(&totalRequests))
	_, _ = fmt.Fprintf(w, "server_errors_total %d\n", atomic.LoadUint64(&totalServerErrors))
}

func main() {
	app := newAPI()
	server := &http.Server{
		Addr:    ":8080",
		Handler: app.router(),
	}

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-stopCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown error: %v", err)
		}
	}()

	log.Println("api server listening on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
