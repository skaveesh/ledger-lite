package main

import (
	"net/http"
	"testing"

	"github.com/skaveesh/ledger-lite/internal/domain"
)

func TestCategoriesGetEmpty(t *testing.T) {
	_, router := newTestServer()

	rr := performRequest(t, router, http.MethodGet, "/categories", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /categories status = %d, want %d", rr.Code, http.StatusOK)
	}

	got := decodeJSON[[]domain.Category](t, rr)
	if len(got) != 0 {
		t.Fatalf("GET /categories len = %d, want 0", len(got))
	}
}

func TestCategoriesPostThenGet(t *testing.T) {
	_, router := newTestServer()

	createResp := performRequest(t, router, http.MethodPost, "/categories", map[string]any{
		"name": "Food",
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("POST /categories status = %d, want %d", createResp.Code, http.StatusCreated)
	}
	created := decodeJSON[domain.Category](t, createResp)
	if created.ID == 0 {
		t.Fatal("POST /categories returned ID 0")
	}

	listResp := performRequest(t, router, http.MethodGet, "/categories", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("GET /categories status = %d, want %d", listResp.Code, http.StatusOK)
	}
	list := decodeJSON[[]domain.Category](t, listResp)
	if len(list) != 1 || list[0].Name != "Food" {
		t.Fatalf("GET /categories response = %+v, want one Food category", list)
	}
}
