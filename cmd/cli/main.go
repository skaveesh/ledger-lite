package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/skaveesh/ledger-lite/internal/domain"
	"github.com/skaveesh/ledger-lite/internal/store/sqlite"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("ledger-cli", flag.ContinueOnError)
	dbPath := fs.String("db", "ledgerlite.db", "path to SQLite database file")
	fs.SetOutput(os.Stdout)

	if err := fs.Parse(args); err != nil {
		return err
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
		for _, c := range categories {
			fmt.Printf("%d\t%s\n", c.ID, c.Name)
		}
		return nil
	default:
		return fmt.Errorf("unknown category subcommand: %s", args[0])
	}
}

func runTransaction(dbPath string, args []string) error {
	_ = dbPath
	_ = args
	return fmt.Errorf("transaction command not implemented")
}

func runBudget(dbPath string, args []string) error {
	_ = dbPath
	_ = args
	return fmt.Errorf("budget command not implemented")
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
}
