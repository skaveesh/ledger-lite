# LedgerLite 90-Day Checklist (30 min/day)

## Week 1: Setup & Domain
- [x] Day 1: Create repo, README, Go module.
- [x] Day 2: Decide folder structure; add `.gitignore`.
- [x] Day 3: Implement simple `main.go` with “LedgerLite”.
- [x] Day 4: Define data structs: Transaction, Category, Budget.
- [x] Day 5: Create in-memory store interface.
- [x] Day 6: Implement in-memory store.
- [x] Day 7: Write tests for in-memory store.

## Week 2: SQLite Setup
- [x] Day 8: Add SQLite dependency.
- [x] Day 9: Write DB init and schema migration.
- [x] Day 10: Implement SQLite store interface.
- [x] Day 11: Add CRUD for Category in SQLite.
- [x] Day 12: Add CRUD for Transaction in SQLite.
- [x] Day 13: Add CRUD for Budget in SQLite.
- [x] Day 14: Write tests for SQLite store (basic).

## Week 3: REST API Skeleton
- [x] Day 15: Add HTTP server skeleton.
- [x] Day 16: Implement `/health`.
- [x] Day 17: Add router library (chi or std).
- [x] Day 18: Add `/categories` GET.
- [x] Day 19: Add `/categories` POST.
- [x] Day 20: Add `/transactions` GET.
- [x] Day 21: Add `/transactions` POST.

## Week 4: API Continued
- [x] Day 22: Add `/transactions/{id}` GET.
- [x] Day 23: Add `/transactions/{id}` DELETE.
- [x] Day 24: Add `/budgets` GET.
- [x] Day 25: Add `/budgets` POST.
- [x] Day 26: Add `/budgets/{id}` PUT.
- [x] Day 27: Add validation helpers.
- [x] Day 28: Add error handling middleware.

## Week 5: API Tests
- [x] Day 29: Add HTTP test setup helpers.
- [x] Day 30: Test `/categories` GET/POST.
- [x] Day 31: Test `/transactions` GET/POST.
- [x] Day 32: Test `/transactions/{id}` GET/DELETE.
- [x] Day 33: Test `/budgets` GET/POST.
- [x] Day 34: Test `/budgets/{id}` PUT.
- [x] Day 35: Add test data builders.

## Week 6: CLI Basics
- [x] Day 36: Add CLI skeleton (cobra or flag).
- [x] Day 37: Add `category add` command.
- [x] Day 38: Add `category list` command.
- [x] Day 39: Add `transaction add` command.
- [x] Day 40: Add `transaction list` command.
- [x] Day 41: Add `budget set` command.
- [x] Day 42: Add CLI help docs in README.

## Week 7: CLI Improvements
- [ ] Day 43: Add `transaction delete` command.
- [ ] Day 44: Add `category update` command.
- [ ] Day 45: Add `budget list` command.
- [ ] Day 46: Add CLI config file support.
- [ ] Day 47: Add pretty output formatting.
- [ ] Day 48: Add CLI error messages.
- [ ] Day 49: Write CLI tests (basic).

## Week 8: Web UI Skeleton
- [ ] Day 50: Add HTML templates folder.
- [ ] Day 51: Add `/ui` route.
- [ ] Day 52: Create base layout template.
- [ ] Day 53: Create transactions list page.
- [ ] Day 54: Create categories list page.
- [ ] Day 55: Create budget list page.
- [ ] Day 56: Add form to add transaction.

## Week 9: UI Enhancements
- [ ] Day 57: Add date filter for transactions.
- [ ] Day 58: Add category filter.
- [ ] Day 59: Add monthly summary page.
- [ ] Day 60: Add minimal CSS styling.
- [ ] Day 61: Add delete buttons with POST.
- [ ] Day 62: Add UI error handling.
- [ ] Day 63: Add navigation links.

## Week 10: Polishing
- [ ] Day 64: Add JSON schema/validation.
- [ ] Day 65: Add pagination for transactions.
- [ ] Day 66: Add sorting (date, amount).
- [ ] Day 67: Add logging middleware.
- [ ] Day 68: Add graceful shutdown.
- [ ] Day 69: Add request ID middleware.
- [ ] Day 70: Add basic metrics endpoint.

## Week 11: Documentation & Deployment
- [ ] Day 71: Update README with setup.
- [ ] Day 72: Add API docs section.
- [ ] Day 73: Add CLI usage examples.
- [ ] Day 74: Add UI screenshots (optional).
- [ ] Day 75: Add Dockerfile.
- [ ] Day 76: Add docker-compose for DB.
- [ ] Day 77: Write deployment notes.

## Week 12: Stretch Goals
- [ ] Day 78: Add recurring transactions.
- [ ] Day 79: Add CSV export endpoint.
- [ ] Day 80: Add CSV import CLI.
- [ ] Day 81: Add budget alerts.
- [ ] Day 82: Add category spending summary.
- [ ] Day 83: Add net cashflow endpoint.
- [ ] Day 84: Add savings goal tracker.

## Week 13: Cleanup & Review
- [ ] Day 85: Refactor packages (store, api).
- [ ] Day 86: Improve test coverage.
- [ ] Day 87: Run `go vet` and fix warnings.
- [ ] Day 88: Improve error types.
- [ ] Day 89: Optimize query performance.
- [ ] Day 90: Final cleanup & backlog list.

## CLI Help
Use the CLI with an explicit SQLite path (or omit `--db` to use `ledgerlite.db` in the current folder).

```bash
# show general help
go run ./cmd/cli -- help

# categories
go run ./cmd/cli -- category add --name Food
go run ./cmd/cli -- category list

# transactions
# date must be RFC3339
go run ./cmd/cli -- transaction add --category-id 1 --amount-cents 4500 --description "Groceries" --date "2026-03-10T12:00:00Z"
go run ./cmd/cli -- transaction list

# budgets
go run ./cmd/cli -- budget set --category-id 1 --month 3 --year 2026 --amount-limit-cents 150000
```
