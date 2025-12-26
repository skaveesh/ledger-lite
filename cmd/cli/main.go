package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/skaveesh/ledger-lite/internal/domain"
	"github.com/skaveesh/ledger-lite/internal/store/sqlite"
)

type cliConfig struct {
	DB string `json:"db"`
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("ledger-cli", flag.ContinueOnError)
	dbPath := fs.String("db", "ledgerlite.db", "path to SQLite database file")
	configPath := fs.String("config", "", "path to CLI config JSON file")
	fs.SetOutput(os.Stdout)

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *configPath != "" {
		cfg, err := loadConfig(*configPath)
		if err != nil {
			return err
		}
		if cfg.DB != "" && !hasOption(args, "--db") {
			*dbPath = cfg.DB
		}
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		printUsage()
		return nil
	}

	switch remaining[0] {
	case "category":
		return runCategory(*dbPath, remaining[1:])
	case "transaction":
		return runTransaction(*dbPath, remaining[1:])
	case "budget":
		return runBudget(*dbPath, remaining[1:])
	case "help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", remaining[0])
	}
}

func loadConfig(path string) (cliConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return cliConfig{}, fmt.Errorf("read config: %w", err)
	}

	var cfg cliConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cliConfig{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func hasOption(args []string, option string) bool {
	for i := range args {
		if args[i] == option || strings.HasPrefix(args[i], option+"=") {
			return true
		}
	}
	return false
}

func formatCents(amount int64) string {
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	return fmt.Sprintf("%s$%d.%02d", sign, amount/100, amount%100)
}

func runCategory(dbPath string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing category subcommand")
	}

	s, err := sqlite.New(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	switch args[0] {
	case "add":
		addFS := flag.NewFlagSet("category add", flag.ContinueOnError)
		name := addFS.String("name", "", "category name")
		addFS.SetOutput(os.Stdout)
		if err := addFS.Parse(args[1:]); err != nil {
			return err
		}
		if *name == "" {
			return fmt.Errorf("--name is required")
		}

		category, err := s.CreateCategory(domain.Category{Name: *name})
		if err != nil {
			return err
		}
		fmt.Printf("category created: id=%d name=%s\n", category.ID, category.Name)
		return nil
	case "list":
		categories := s.ListCategories()
		if len(categories) == 0 {
			fmt.Println("no categories")
			return nil
		}
		tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
		_, _ = fmt.Fprintln(tw, "ID\tName")
		for _, c := range categories {
			_, _ = fmt.Fprintf(tw, "%d\t%s\n", c.ID, c.Name)
		}
		_ = tw.Flush()
		return nil
	case "update":
		updateFS := flag.NewFlagSet("category update", flag.ContinueOnError)
		id := updateFS.Int64("id", 0, "category id")
		name := updateFS.String("name", "", "updated category name")
		updateFS.SetOutput(os.Stdout)
		if err := updateFS.Parse(args[1:]); err != nil {
			return err
		}
		if *id == 0 || *name == "" {
			return fmt.Errorf("--id and --name are required")
		}

		updated, ok, err := s.UpdateCategory(*id, domain.Category{Name: *name})
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("category %d not found", *id)
		}

		fmt.Printf("category updated: id=%d name=%s\n", updated.ID, updated.Name)
		return nil
	default:
		return fmt.Errorf("unknown category subcommand: %s", args[0])
	}
}

func runTransaction(dbPath string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing transaction subcommand")
	}

	s, err := sqlite.New(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	switch args[0] {
	case "add":
		addFS := flag.NewFlagSet("transaction add", flag.ContinueOnError)
		categoryID := addFS.Int64("category-id", 0, "category id")
		amountCents := addFS.Int64("amount-cents", 0, "amount in cents")
		description := addFS.String("description", "", "transaction description")
		dateStr := addFS.String("date", "", "transaction date in RFC3339")
		addFS.SetOutput(os.Stdout)
		if err := addFS.Parse(args[1:]); err != nil {
			return err
		}
		if *categoryID == 0 || *amountCents == 0 || *dateStr == "" {
			return fmt.Errorf("--category-id, --amount-cents and --date are required")
		}

		parsedDate, err := time.Parse(time.RFC3339, *dateStr)
		if err != nil {
			return fmt.Errorf("invalid --date, expected RFC3339: %w", err)
		}

		transaction, err := s.CreateTransaction(domain.Transaction{
			CategoryID:  *categoryID,
			AmountCents: *amountCents,
			Description: *description,
			Date:        parsedDate,
		})
		if err != nil {
			return err
		}

		fmt.Printf("transaction created: id=%d category_id=%d amount_cents=%d\n", transaction.ID, transaction.CategoryID, transaction.AmountCents)
		return nil
	case "list":
		transactions := s.ListTransactions()
		if len(transactions) == 0 {
			fmt.Println("no transactions")
			return nil
		}
		tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
		_, _ = fmt.Fprintln(tw, "ID\tCategory\tAmount\tDescription\tDate")
		for _, tr := range transactions {
			_, _ = fmt.Fprintf(tw, "%d\t%d\t%s\t%s\t%s\n", tr.ID, tr.CategoryID, formatCents(tr.AmountCents), tr.Description, tr.Date.Format(time.RFC3339))
		}
		_ = tw.Flush()
		return nil
	case "delete":
		delFS := flag.NewFlagSet("transaction delete", flag.ContinueOnError)
		id := delFS.Int64("id", 0, "transaction id")
		delFS.SetOutput(os.Stdout)
		if err := delFS.Parse(args[1:]); err != nil {
			return err
		}
		if *id == 0 {
			return fmt.Errorf("--id is required")
		}
		if !s.DeleteTransaction(*id) {
			return fmt.Errorf("transaction %d not found", *id)
		}
		fmt.Printf("transaction deleted: id=%d\n", *id)
		return nil
	default:
		return fmt.Errorf("unknown transaction subcommand: %s", args[0])
	}
}

func runBudget(dbPath string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing budget subcommand")
	}

	s, err := sqlite.New(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	switch args[0] {
	case "set":
		setFS := flag.NewFlagSet("budget set", flag.ContinueOnError)
		categoryID := setFS.Int64("category-id", 0, "category id")
		month := setFS.Int("month", 0, "month number (1-12)")
		year := setFS.Int("year", 0, "year")
		amountLimitCents := setFS.Int64("amount-limit-cents", 0, "budget amount limit in cents")
		setFS.SetOutput(os.Stdout)
		if err := setFS.Parse(args[1:]); err != nil {
			return err
		}
		if *categoryID == 0 || *month == 0 || *year == 0 || *amountLimitCents == 0 {
			return fmt.Errorf("--category-id, --month, --year and --amount-limit-cents are required")
		}

		budget, err := s.CreateBudget(domain.Budget{
			CategoryID:       *categoryID,
			Month:            *month,
			Year:             *year,
			AmountLimitCents: *amountLimitCents,
		})
		if err != nil {
			return err
		}

		fmt.Printf("budget set: id=%d category_id=%d month=%d year=%d amount_limit_cents=%d\n", budget.ID, budget.CategoryID, budget.Month, budget.Year, budget.AmountLimitCents)
		return nil
	case "list":
		budgets := s.ListBudgets()
		if len(budgets) == 0 {
			fmt.Println("no budgets")
			return nil
		}
		tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
		_, _ = fmt.Fprintln(tw, "ID\tCategory\tMonth\tYear\tLimit")
		for _, b := range budgets {
			_, _ = fmt.Fprintf(tw, "%d\t%d\t%d\t%d\t%s\n", b.ID, b.CategoryID, b.Month, b.Year, formatCents(b.AmountLimitCents))
		}
		_ = tw.Flush()
		return nil
	default:
		return fmt.Errorf("unknown budget subcommand: %s", args[0])
	}
}

func printUsage() {
	fmt.Println("LedgerLite CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ledger-cli [--db path] <command> <subcommand> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  category")
	fmt.Println("  transaction")
	fmt.Println("  budget")
	fmt.Println("  help")
	fmt.Println()
	fmt.Println("Global options:")
	fmt.Println("  --db <path>       SQLite database path (default: ledgerlite.db)")
	fmt.Println("  --config <path>   JSON config file with {\"db\":\"path/to/db\"}")
}
