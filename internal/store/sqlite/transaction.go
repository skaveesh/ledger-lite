package sqlite

import (
	"errors"
	"fmt"
	"time"

	"github.com/skaveesh/ledger-lite/internal/domain"
)

func (s *Store) CreateTransaction(transaction domain.Transaction) (domain.Transaction, error) {
	if transaction.Date.IsZero() {
		return domain.Transaction{}, errors.New("transaction date is required")
	}

	res, err := s.db.Exec(
		"INSERT INTO transactions(category_id, amount_cents, description, date) VALUES(?, ?, ?, ?)",
		transaction.CategoryID,
		transaction.AmountCents,
		transaction.Description,
		transaction.Date.Format(time.RFC3339),
	)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("insert transaction: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("transaction id: %w", err)
	}
	transaction.ID = id
	return transaction, nil
}

func (s *Store) GetTransaction(id int64) (domain.Transaction, bool) {
	row := s.db.QueryRow("SELECT id, category_id, amount_cents, description, date FROM transactions WHERE id = ?", id)
	var t domain.Transaction
	var rawDate string
	if err := row.Scan(&t.ID, &t.CategoryID, &t.AmountCents, &t.Description, &rawDate); err != nil {
		return domain.Transaction{}, false
	}

	parsed, err := time.Parse(time.RFC3339, rawDate)
	if err != nil {
		return domain.Transaction{}, false
	}
	t.Date = parsed
	return t, true
}

func (s *Store) ListTransactions() []domain.Transaction {
	rows, err := s.db.Query("SELECT id, category_id, amount_cents, description, date FROM transactions ORDER BY id")
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()

	items := make([]domain.Transaction, 0)
	for rows.Next() {
		var t domain.Transaction
		var rawDate string
		if err := rows.Scan(&t.ID, &t.CategoryID, &t.AmountCents, &t.Description, &rawDate); err != nil {
			return items
		}
		parsed, err := time.Parse(time.RFC3339, rawDate)
		if err != nil {
			return items
		}
		t.Date = parsed
		items = append(items, t)
	}
	return items
}

func (s *Store) DeleteTransaction(id int64) bool {
	res, err := s.db.Exec("DELETE FROM transactions WHERE id = ?", id)
	if err != nil {
		return false
	}
	affected, err := res.RowsAffected()
	return err == nil && affected > 0
}
