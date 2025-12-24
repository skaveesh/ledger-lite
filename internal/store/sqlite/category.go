package sqlite

import (
	"errors"
	"fmt"

	"github.com/skaveesh/ledger-lite/internal/domain"
)

func (s *Store) CreateCategory(category domain.Category) (domain.Category, error) {
	if category.Name == "" {
		return domain.Category{}, errors.New("category name is required")
	}

	res, err := s.db.Exec("INSERT INTO categories(name) VALUES(?)", category.Name)
	if err != nil {
		return domain.Category{}, fmt.Errorf("insert category: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return domain.Category{}, fmt.Errorf("category id: %w", err)
	}
	category.ID = id
	return category, nil
}

func (s *Store) GetCategory(id int64) (domain.Category, bool) {
	row := s.db.QueryRow("SELECT id, name FROM categories WHERE id = ?", id)
	var category domain.Category
	if err := row.Scan(&category.ID, &category.Name); err != nil {
		return domain.Category{}, false
	}
	return category, true
}

func (s *Store) ListCategories() []domain.Category {
	rows, err := s.db.Query("SELECT id, name FROM categories ORDER BY id")
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()

	items := make([]domain.Category, 0)
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return items
		}
		items = append(items, c)
	}
	return items
}

func (s *Store) UpdateCategory(id int64, category domain.Category) (domain.Category, bool, error) {
	if category.Name == "" {
		return domain.Category{}, true, errors.New("category name is required")
	}

	res, err := s.db.Exec("UPDATE categories SET name = ? WHERE id = ?", category.Name, id)
	if err != nil {
		return domain.Category{}, true, fmt.Errorf("update category: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return domain.Category{}, true, fmt.Errorf("update category rows affected: %w", err)
	}
	if affected == 0 {
		return domain.Category{}, false, nil
	}

	category.ID = id
	return category, true, nil
}

func (s *Store) DeleteCategory(id int64) bool {
	res, err := s.db.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		return false
	}
	affected, err := res.RowsAffected()
	return err == nil && affected > 0
}
