package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/skaveesh/ledger-lite/internal/domain"
)

func TestTransactionsIDGetAndDelete(t *testing.T) {
	_, router := newTestServer()
	date := time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)

	createResp := performRequest(t, router, http.MethodPost, "/transactions", map[string]any{
		"categoryID":  1,
		"amountCents": 1200,
		"description": "Coffee",
		"date":        date,
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("POST /transactions status = %d, want %d", createResp.Code, http.StatusCreated)
	}
	created := decodeJSON[domain.Transaction](t, createResp)

	getPath := fmt.Sprintf("/transactions/%d", created.ID)
	getResp := performRequest(t, router, http.MethodGet, getPath, nil)
	if getResp.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, want %d", getPath, getResp.Code, http.StatusOK)
	}
	fetched := decodeJSON[domain.Transaction](t, getResp)
	if fetched.ID != created.ID {
		t.Fatalf("GET %s id = %d, want %d", getPath, fetched.ID, created.ID)
	}

	deleteResp := performRequest(t, router, http.MethodDelete, getPath, nil)
	if deleteResp.Code != http.StatusNoContent {
		t.Fatalf("DELETE %s status = %d, want %d", getPath, deleteResp.Code, http.StatusNoContent)
	}

	notFoundResp := performRequest(t, router, http.MethodGet, getPath, nil)
	if notFoundResp.Code != http.StatusNotFound {
		t.Fatalf("GET deleted %s status = %d, want %d", getPath, notFoundResp.Code, http.StatusNotFound)
	}
}
