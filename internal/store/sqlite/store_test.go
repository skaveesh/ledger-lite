package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/skaveesh/ledger-lite/internal/domain"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ledgerlite-test.db")
	s, err := New(path)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})
	return s
}

func TestCategoryCRUD(t *testing.T) {
	s := newTestStore(t)

	created, err := s.CreateCategory(domain.Category{Name: "Food"})
	if err != nil {
		t.Fatalf("CreateCategory() error = %v", err)
	}
	if created.ID == 0 {
		t.Fatal("CreateCategory() returned ID 0")
	}

	got, ok := s.GetCategory(created.ID)
	if !ok || got.Name != "Food" {
		t.Fatalf("GetCategory() = (%v, %v), want Food", got, ok)
	}

	if len(s.ListCategories()) != 1 {
		t.Fatalf("ListCategories() len = %d, want 1", len(s.ListCategories()))
	}

	if !s.DeleteCategory(created.ID) {
		t.Fatal("DeleteCategory() = false, want true")
	}
}

func TestTransactionCRUD(t *testing.T) {
	s := newTestStore(t)
	category, err := s.CreateCategory(domain.Category{Name: "Bills"})
	if err != nil {
		t.Fatalf("CreateCategory() error = %v", err)
	}

	created, err := s.CreateTransaction(domain.Transaction{
		CategoryID:  category.ID,
		AmountCents: 9999,
		Description: "Internet",
		Date:        time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("CreateTransaction() error = %v", err)
	}

	got, ok := s.GetTransaction(created.ID)
	if !ok || got.AmountCents != 9999 {
		t.Fatalf("GetTransaction() = (%v, %v), want amount 9999", got, ok)
	}

	if len(s.ListTransactions()) != 1 {
		t.Fatalf("ListTransactions() len = %d, want 1", len(s.ListTransactions()))
	}

	if !s.DeleteTransaction(created.ID) {
		t.Fatal("DeleteTransaction() = false, want true")
	}
}

func TestBudgetCRUD(t *testing.T) {
	s := newTestStore(t)
	category, err := s.CreateCategory(domain.Category{Name: "Utilities"})
	if err != nil {
		t.Fatalf("CreateCategory() error = %v", err)
	}

	created, err := s.CreateBudget(domain.Budget{
		CategoryID:       category.ID,
		Month:            2,
		Year:             2026,
		AmountLimitCents: 120000,
	})
	if err != nil {
		t.Fatalf("CreateBudget() error = %v", err)
	}

	got, ok := s.GetBudget(created.ID)
	if !ok || got.AmountLimitCents != 120000 {
		t.Fatalf("GetBudget() = (%v, %v), want amount 120000", got, ok)
	}

	if len(s.ListBudgets()) != 1 {
		t.Fatalf("ListBudgets() len = %d, want 1", len(s.ListBudgets()))
	}

	if !s.DeleteBudget(created.ID) {
		t.Fatal("DeleteBudget() = false, want true")
	}
}
