package store

import "github.com/skaveesh/ledger-lite/internal/domain"

// Store defines storage operations for core LedgerLite entities.
type Store interface {
	CreateCategory(category domain.Category) (domain.Category, error)
	GetCategory(id int64) (domain.Category, bool)
	ListCategories() []domain.Category
	DeleteCategory(id int64) bool

	CreateTransaction(transaction domain.Transaction) (domain.Transaction, error)
	GetTransaction(id int64) (domain.Transaction, bool)
	ListTransactions() []domain.Transaction
	DeleteTransaction(id int64) bool

	CreateBudget(budget domain.Budget) (domain.Budget, error)
	GetBudget(id int64) (domain.Budget, bool)
	ListBudgets() []domain.Budget
	DeleteBudget(id int64) bool
}
