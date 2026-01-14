package main

import (
	"net/http"
	"testing"

	"github.com/skaveesh/ledger-lite/internal/domain"
)

func TestBudgetsGetEmpty(t *testing.T) {
	_, router := newTestServer()

	rr := performRequest(t, router, http.MethodGet, "/budgets", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /budgets status = %d, want %d", rr.Code, http.StatusOK)
	}

	got := decodeJSONResponse[[]domain.Budget](t, rr)
	if len(got) != 0 {
		t.Fatalf("GET /budgets len = %d, want 0", len(got))
	}
}

func TestBudgetsPostThenGet(t *testing.T) {
	_, router := newTestServer()

	createResp := performRequest(t, router, http.MethodPost, "/budgets", map[string]any{
		"categoryID":       1,
		"month":            3,
		"year":             2026,
		"amountLimitCents": 150000,
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("POST /budgets status = %d, want %d", createResp.Code, http.StatusCreated)
	}
	created := decodeJSONResponse[domain.Budget](t, createResp)
	if created.ID == 0 {
		t.Fatal("POST /budgets returned ID 0")
	}

	listResp := performRequest(t, router, http.MethodGet, "/budgets", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("GET /budgets status = %d, want %d", listResp.Code, http.StatusOK)
	}
	list := decodeJSONResponse[[]domain.Budget](t, listResp)
	if len(list) != 1 || list[0].AmountLimitCents != 150000 {
		t.Fatalf("GET /budgets response = %+v, want one budget amount 150000", list)
	}
}

func TestBudgetsPostInvalidJSONUnknownField(t *testing.T) {
	_, router := newTestServer()

	rr := performRequest(t, router, http.MethodPost, "/budgets", map[string]any{
		"categoryID":       1,
		"month":            3,
		"year":             2026,
		"amountLimitCents": 150000,
		"extra":            "unknown",
	})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("POST /budgets with unknown field status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}
