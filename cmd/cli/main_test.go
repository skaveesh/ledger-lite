package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/skaveesh/ledger-lite/internal/domain"
	"github.com/skaveesh/ledger-lite/internal/store/sqlite"
)

func newTestDBPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "ledgerlite-cli-test.db")
}

func TestRunCategoryAdd(t *testing.T) {
	dbPath := newTestDBPath(t)
	if err := run([]string{"--db", dbPath, "category", "add", "--name", "Food"}); err != nil {
		t.Fatalf("run category add: %v", err)
	}

	s, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = s.Close() }()

	cats := s.ListCategories()
	if len(cats) != 1 || cats[0].Name != "Food" {
		t.Fatalf("categories = %+v, want one Food category", cats)
	}
}

func TestRunTransactionDelete(t *testing.T) {
	dbPath := newTestDBPath(t)
	s, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	cat, err := s.CreateCategory(domain.Category{Name: "Food"})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	tr, err := s.CreateTransaction(domain.Transaction{
		CategoryID:  cat.ID,
		AmountCents: 1200,
		Description: "Coffee",
		Date:        time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	_ = s.Close()

	if err := run([]string{"--db", dbPath, "transaction", "delete", "--id", "1"}); err != nil {
		t.Fatalf("run transaction delete: %v", err)
	}

	s2, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("reopen db: %v", err)
	}
	defer func() { _ = s2.Close() }()

	if _, ok := s2.GetTransaction(tr.ID); ok {
		t.Fatal("transaction still exists after delete")
	}
}

func TestRunWithConfigDB(t *testing.T) {
	dbPath := newTestDBPath(t)
	configPath := filepath.Join(t.TempDir(), "cli-config.json")
	if err := os.WriteFile(configPath, []byte("{\"db\":\""+dbPath+"\"}"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := run([]string{"--config", configPath, "category", "add", "--name", "Bills"}); err != nil {
		t.Fatalf("run with config: %v", err)
	}

	s, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = s.Close() }()

	cats := s.ListCategories()
	if len(cats) != 1 || cats[0].Name != "Bills" {
		t.Fatalf("categories = %+v, want one Bills category", cats)
	}
}
