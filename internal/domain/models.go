package domain

import "time"

// Transaction represents money movement for a category on a given date.
type Transaction struct {
	ID          int64
	CategoryID  int64
	AmountCents int64
	Description string
	Date        time.Time
}

// Category groups transactions for reporting and budgeting.
type Category struct {
	ID   int64
	Name string
}

// Budget defines a monthly spending target for a category.
type Budget struct {
	ID               int64
	CategoryID       int64
	Month            int
	Year             int
	AmountLimitCents int64
}
