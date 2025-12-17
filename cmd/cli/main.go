package main

import (
	"flag"
	"fmt"
	"os"
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
	_ = dbPath

	remaining := fs.Args()
	if len(remaining) == 0 {
		printUsage()
		return nil
	}

	switch remaining[0] {
	case "category":
		return runCategory(remaining[1:])
	case "transaction":
		return runTransaction(remaining[1:])
	case "budget":
		return runBudget(remaining[1:])
	case "help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", remaining[0])
	}
}

func runCategory(args []string) error {
	_ = args
	return fmt.Errorf("category command not implemented")
}

func runTransaction(args []string) error {
	_ = args
	return fmt.Errorf("transaction command not implemented")
}

func runBudget(args []string) error {
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
