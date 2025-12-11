package main

import "time"

func buildCategoryPayload(name string) map[string]any {
	return map[string]any{
		"name": name,
	}
}

func buildTransactionPayload(categoryID int64, amountCents int64, description string, date time.Time) map[string]any {
	return map[string]any{
		"categoryID":  categoryID,
		"amountCents": amountCents,
		"description": description,
		"date":        date.Format(time.RFC3339),
	}
}

func buildBudgetPayload(categoryID int64, month int, year int, amountLimitCents int64) map[string]any {
	return map[string]any{
		"categoryID":       categoryID,
		"month":            month,
		"year":             year,
		"amountLimitCents": amountLimitCents,
	}
}
