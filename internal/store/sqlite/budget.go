package sqlite

import (
	"errors"
	"fmt"

	"github.com/skaveesh/ledger-lite/internal/domain"
)

func (s *Store) CreateBudget(budget domain.Budget) (domain.Budget, error) {
	if budget.Month < 1 || budget.Month > 12 {
		return domain.Budget{}, errors.New("budget month must be between 1 and 12")
	}
	if budget.Year < 1 {
		return domain.Budget{}, errors.New("budget year must be greater than 0")
	}

	res, err := s.db.Exec(
		"INSERT INTO budgets(category_id, month, year, amount_limit_cents) VALUES(?, ?, ?, ?)",
		budget.CategoryID,
		budget.Month,
		budget.Year,
		budget.AmountLimitCents,
	)
	if err != nil {
		return domain.Budget{}, fmt.Errorf("insert budget: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return domain.Budget{}, fmt.Errorf("budget id: %w", err)
	}
	budget.ID = id
	return budget, nil
}

func (s *Store) GetBudget(id int64) (domain.Budget, bool) {
	row := s.db.QueryRow("SELECT id, category_id, month, year, amount_limit_cents FROM budgets WHERE id = ?", id)
	var b domain.Budget
	if err := row.Scan(&b.ID, &b.CategoryID, &b.Month, &b.Year, &b.AmountLimitCents); err != nil {
		return domain.Budget{}, false
	}
	return b, true
}

func (s *Store) ListBudgets() []domain.Budget {
	rows, err := s.db.Query("SELECT id, category_id, month, year, amount_limit_cents FROM budgets ORDER BY id")
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()

	items := make([]domain.Budget, 0)
	for rows.Next() {
		var b domain.Budget
		if err := rows.Scan(&b.ID, &b.CategoryID, &b.Month, &b.Year, &b.AmountLimitCents); err != nil {
			return items
		}
		items = append(items, b)
	}
	return items
}

func (s *Store) DeleteBudget(id int64) bool {
	res, err := s.db.Exec("DELETE FROM budgets WHERE id = ?", id)
	if err != nil {
		return false
	}
	affected, err := res.RowsAffected()
	return err == nil && affected > 0
}
