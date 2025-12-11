package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/skaveesh/ledger-lite/internal/domain"
)

func TestBudgetsIDPut(t *testing.T) {
	_, router := newTestServer()

	createResp := performRequest(t, router, http.MethodPost, "/budgets", map[string]any{
		"categoryID":       1,
		"month":            3,
		"year":             2026,
		"amountLimitCents": 100000,
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("POST /budgets status = %d, want %d", createResp.Code, http.StatusCreated)
	}
	created := decodeJSON[domain.Budget](t, createResp)

	putPath := fmt.Sprintf("/budgets/%d", created.ID)
	putResp := performRequest(t, router, http.MethodPut, putPath, map[string]any{
		"categoryID":       1,
		"month":            4,
		"year":             2026,
		"amountLimitCents": 125000,
	})
	if putResp.Code != http.StatusOK {
		t.Fatalf("PUT %s status = %d, want %d", putPath, putResp.Code, http.StatusOK)
	}
	updated := decodeJSON[domain.Budget](t, putResp)
	if updated.Month != 4 || updated.AmountLimitCents != 125000 {
		t.Fatalf("PUT %s response = %+v, want updated month/amount", putPath, updated)
	}
}
