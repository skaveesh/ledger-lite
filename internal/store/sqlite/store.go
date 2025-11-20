package sqlite

import (
	"database/sql"

	rootstore "github.com/skaveesh/ledger-lite/internal/store"
)

// Store implements SQLite-backed persistence.
type Store struct {
	db *sql.DB
}

// Ensure SQLite store satisfies the shared storage contract.
var _ rootstore.Store = (*Store)(nil)

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}
