package store

import (
	"errors"
	"sort"
	"sync"

	"github.com/skaveesh/ledger-lite/internal/domain"
)

var (
	errCategoryNameRequired    = errors.New("category name is required")
	errTransactionDateRequired = errors.New("transaction date is required")
	errBudgetMonthInvalid      = errors.New("budget month must be between 1 and 12")
	errBudgetYearInvalid       = errors.New("budget year must be greater than 0")
)

// MemoryStore keeps data in memory and is safe for concurrent access.
type MemoryStore struct {
	mu sync.RWMutex

	categories   map[int64]domain.Category
	transactions map[int64]domain.Transaction
	budgets      map[int64]domain.Budget

	nextCategoryID    int64
	nextTransactionID int64
	nextBudgetID      int64
}

// NewMemoryStore creates an empty in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		categories:        make(map[int64]domain.Category),
		transactions:      make(map[int64]domain.Transaction),
		budgets:           make(map[int64]domain.Budget),
		nextCategoryID:    1,
		nextTransactionID: 1,
		nextBudgetID:      1,
	}
}

func (s *MemoryStore) CreateCategory(category domain.Category) (domain.Category, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if category.Name == "" {
		return domain.Category{}, errCategoryNameRequired
	}

	category.ID = s.nextCategoryID
	s.nextCategoryID++
	s.categories[category.ID] = category

	return category, nil
}

func (s *MemoryStore) GetCategory(id int64) (domain.Category, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	category, ok := s.categories[id]
	return category, ok
}

func (s *MemoryStore) ListCategories() []domain.Category {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]domain.Category, 0, len(s.categories))
	for _, item := range s.categories {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items
}

func (s *MemoryStore) UpdateCategory(id int64, category domain.Category) (domain.Category, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.categories[id]; !ok {
		return domain.Category{}, false, nil
	}
	if category.Name == "" {
		return domain.Category{}, true, errCategoryNameRequired
	}

	category.ID = id
	s.categories[id] = category
	return category, true, nil
}

func (s *MemoryStore) DeleteCategory(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.categories[id]; !ok {
		return false
	}
	delete(s.categories, id)
	return true
}

func (s *MemoryStore) CreateTransaction(transaction domain.Transaction) (domain.Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if transaction.Date.IsZero() {
		return domain.Transaction{}, errTransactionDateRequired
	}

	transaction.ID = s.nextTransactionID
	s.nextTransactionID++
	s.transactions[transaction.ID] = transaction
	return transaction, nil
}

func (s *MemoryStore) GetTransaction(id int64) (domain.Transaction, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.transactions[id]
	return item, ok
}

func (s *MemoryStore) ListTransactions() []domain.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]domain.Transaction, 0, len(s.transactions))
	for _, item := range s.transactions {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items
}

func (s *MemoryStore) DeleteTransaction(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.transactions[id]; !ok {
		return false
	}
	delete(s.transactions, id)
	return true
}

func (s *MemoryStore) CreateBudget(budget domain.Budget) (domain.Budget, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if budget.Month < 1 || budget.Month > 12 {
		return domain.Budget{}, errBudgetMonthInvalid
	}
	if budget.Year < 1 {
		return domain.Budget{}, errBudgetYearInvalid
	}

	budget.ID = s.nextBudgetID
	s.nextBudgetID++
	s.budgets[budget.ID] = budget
	return budget, nil
}

func (s *MemoryStore) GetBudget(id int64) (domain.Budget, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.budgets[id]
	return item, ok
}

func (s *MemoryStore) ListBudgets() []domain.Budget {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]domain.Budget, 0, len(s.budgets))
	for _, item := range s.budgets {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items
}

func (s *MemoryStore) UpdateBudget(id int64, budget domain.Budget) (domain.Budget, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.budgets[id]; !ok {
		return domain.Budget{}, false, nil
	}
	if budget.Month < 1 || budget.Month > 12 {
		return domain.Budget{}, true, errBudgetMonthInvalid
	}
	if budget.Year < 1 {
		return domain.Budget{}, true, errBudgetYearInvalid
	}

	budget.ID = id
	s.budgets[id] = budget
	return budget, true, nil
}

func (s *MemoryStore) DeleteBudget(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.budgets[id]; !ok {
		return false
	}
	delete(s.budgets, id)
	return true
}
