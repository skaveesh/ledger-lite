package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/skaveesh/ledger-lite/internal/domain"
)

func TestTransactionsGetEmpty(t *testing.T) {
	_, router := newTestServer()

	rr := performRequest(t, router, http.MethodGet, "/transactions", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /transactions status = %d, want %d", rr.Code, http.StatusOK)
	}

	got := decodeJSONResponse[[]domain.Transaction](t, rr)
	if len(got) != 0 {
		t.Fatalf("GET /transactions len = %d, want 0", len(got))
	}
}

func TestTransactionsPostThenGet(t *testing.T) {
	_, router := newTestServer()
	date := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)

	createResp := performRequest(t, router, http.MethodPost, "/transactions", map[string]any{
		"categoryID":  1,
		"amountCents": 4500,
		"description": "Groceries",
		"date":        date,
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("POST /transactions status = %d, want %d", createResp.Code, http.StatusCreated)
	}
	created := decodeJSONResponse[domain.Transaction](t, createResp)
	if created.ID == 0 {
		t.Fatal("POST /transactions returned ID 0")
	}

	listResp := performRequest(t, router, http.MethodGet, "/transactions", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("GET /transactions status = %d, want %d", listResp.Code, http.StatusOK)
	}
	list := decodeJSONResponse[[]domain.Transaction](t, listResp)
	if len(list) != 1 || list[0].AmountCents != 4500 {
		t.Fatalf("GET /transactions response = %+v, want one transaction amount 4500", list)
	}
}

func TestTransactionsPostInvalidJSONUnknownField(t *testing.T) {
	_, router := newTestServer()

	rr := performRequest(t, router, http.MethodPost, "/transactions", map[string]any{
		"categoryID":  1,
		"amountCents": 4500,
		"description": "Groceries",
		"date":        time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
		"oops":        "unknown",
	})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("POST /transactions with unknown field status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestTransactionsGetPagination(t *testing.T) {
	_, router := newTestServer()
	for i := 1; i <= 3; i++ {
		resp := performRequest(t, router, http.MethodPost, "/transactions", map[string]any{
			"categoryID":  1,
			"amountCents": i * 100,
			"description": "item",
			"date":        time.Date(2026, 3, i, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
		})
		if resp.Code != http.StatusCreated {
			t.Fatalf("POST /transactions status = %d, want %d", resp.Code, http.StatusCreated)
		}
	}

	rr := performRequest(t, router, http.MethodGet, "/transactions?page=2&page_size=2", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /transactions pagination status = %d, want %d", rr.Code, http.StatusOK)
	}
	got := decodeJSONResponse[[]domain.Transaction](t, rr)
	if len(got) != 1 {
		t.Fatalf("GET /transactions pagination len = %d, want 1", len(got))
	}
}
