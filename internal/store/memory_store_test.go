package store

import (
	"testing"
	"time"

	"github.com/skaveesh/ledger-lite/internal/domain"
)

func TestMemoryStoreCategoryCRUD(t *testing.T) {
	s := NewMemoryStore()

	created, err := s.CreateCategory(domain.Category{Name: "Food"})
	if err != nil {
		t.Fatalf("CreateCategory() error = %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("CreateCategory() ID = %d, want 1", created.ID)
	}

	got, ok := s.GetCategory(created.ID)
	if !ok || got.Name != "Food" {
		t.Fatalf("GetCategory() = (%v, %v), want Food", got, ok)
	}

	listed := s.ListCategories()
	if len(listed) != 1 {
		t.Fatalf("ListCategories() len = %d, want 1", len(listed))
	}

	if !s.DeleteCategory(created.ID) {
		t.Fatal("DeleteCategory() = false, want true")
	}
	if _, ok := s.GetCategory(created.ID); ok {
		t.Fatal("GetCategory() still found deleted category")
	}
}

func TestMemoryStoreTransactionCRUD(t *testing.T) {
	s := NewMemoryStore()

	date := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	created, err := s.CreateTransaction(domain.Transaction{
		CategoryID:  1,
		AmountCents: 1500,
		Description: "Lunch",
		Date:        date,
	})
	if err != nil {
		t.Fatalf("CreateTransaction() error = %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("CreateTransaction() ID = %d, want 1", created.ID)
	}

	got, ok := s.GetTransaction(created.ID)
	if !ok || got.AmountCents != 1500 {
		t.Fatalf("GetTransaction() = (%v, %v), want amount 1500", got, ok)
	}

	if len(s.ListTransactions()) != 1 {
		t.Fatalf("ListTransactions() len = %d, want 1", len(s.ListTransactions()))
	}

	if !s.DeleteTransaction(created.ID) {
		t.Fatal("DeleteTransaction() = false, want true")
	}
}

func TestMemoryStoreBudgetCRUD(t *testing.T) {
	s := NewMemoryStore()

	created, err := s.CreateBudget(domain.Budget{
		CategoryID:       1,
		Month:            1,
		Year:             2026,
		AmountLimitCents: 50000,
	})
	if err != nil {
		t.Fatalf("CreateBudget() error = %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("CreateBudget() ID = %d, want 1", created.ID)
	}

	got, ok := s.GetBudget(created.ID)
	if !ok || got.AmountLimitCents != 50000 {
		t.Fatalf("GetBudget() = (%v, %v), want amount 50000", got, ok)
	}

	if len(s.ListBudgets()) != 1 {
		t.Fatalf("ListBudgets() len = %d, want 1", len(s.ListBudgets()))
	}

	if !s.DeleteBudget(created.ID) {
		t.Fatal("DeleteBudget() = false, want true")
	}
}

