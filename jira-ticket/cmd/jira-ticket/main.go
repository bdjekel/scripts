// Package main provides the CLI entry point for the jira-ticket tool.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/user/jira-ticket/internal/generator"
	"github.com/user/jira-ticket/internal/jiracli"
	"github.com/user/jira-ticket/pkg/config"
)

const (
	exitSuccess = 0
	exitError   = 1
)

func main() {
	os.Exit(run(os.Args[1:]))
}

// run executes the CLI and returns the exit code.
func run(args []string) int {
	// Define flags
	fs := flag.NewFlagSet("jira-ticket", flag.ContinueOnError)

	help := fs.Bool("help", false, "Display usage information")
	dryRun := fs.Bool("dry-run", false, "Parse only, don't create tickets")
	verbose := fs.Bool("verbose", false, "Enable detailed logging")
	onDuplicate := fs.String("on-duplicate", "skip", "Action on duplicate: 'skip' (default) or 'fail'")

	// Parse flags
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitError
	}

	// Handle help flag
	if *help {
		printUsage(fs)
		return exitSuccess
	}

	// Get positional argument (requirements file path)
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: missing requirements file path")
		printUsage(fs)
		return exitError
	}

	filePath := fs.Arg(0)

	// Validate on-duplicate flag
	if *onDuplicate != "skip" && *onDuplicate != "fail" {
		fmt.Fprintf(os.Stderr, "Error: --on-duplicate must be 'skip' or 'fail', got '%s'\n", *onDuplicate)
		return exitError
	}

	// Load and validate configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		return exitError
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		return exitError
	}

	// Create Jira CLI adapter
	adapter := jiracli.NewAdapter(jiracli.Config{
		ProjectKey: cfg.ProjectKey,
		ServerURL:  cfg.ServerURL,
	})

	// Check jira-cli is installed (skip for dry-run mode)
	if !*dryRun {
		if err := adapter.CheckInstalled(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return exitError
		}
	}

	// Create generator
	gen := generator.NewGenerator(adapter)

	// Run generator
	opts := generator.Options{
		DryRun:      *dryRun,
		Verbose:     *verbose,
		OnDuplicate: *onDuplicate,
	}

	result, err := gen.Generate(filePath, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitError
	}

	// Print summary
	printSummary(result, *verbose)

	// Return exit code based on failures
	if len(result.Failed) > 0 {
		return exitError
	}

	return exitSuccess
}

// printUsage prints the CLI usage information.
func printUsage(fs *flag.FlagSet) {
	fmt.Println("Usage: jira-ticket [options] <requirements.md>")
	fmt.Println()
	fmt.Println("Creates Jira Epics and Stories from a requirements markdown file.")
	fmt.Println()
	fmt.Println("Options:")
	fs.PrintDefaults()
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  JIRA_PROJECT_KEY  Jira project key (required)")
	fmt.Println("  JIRA_SERVER_URL   Jira server URL (required)")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  jira-ticket --dry-run requirements.md")
	fmt.Println("  jira-ticket --verbose --on-duplicate=fail requirements.md")
}

// printSummary prints the generation result summary.
func printSummary(result *generator.Result, verbose bool) {
	fmt.Println()
	fmt.Println("=== Summary ===")

	if len(result.Created) > 0 {
		fmt.Printf("Created: %d tickets\n", len(result.Created))
		if verbose {
			for _, key := range result.Created {
				fmt.Printf("  - %s\n", key)
			}
		}
	}

	if len(result.Skipped) > 0 {
		fmt.Printf("Skipped: %d duplicates\n", len(result.Skipped))
		if verbose {
			for _, key := range result.Skipped {
				fmt.Printf("  - %s\n", key)
			}
		}
	}

	if len(result.Failed) > 0 {
		fmt.Printf("Failed: %d tickets\n", len(result.Failed))
		for _, failed := range result.Failed {
			fmt.Printf("  - %s (%s): %s\n", failed.Summary, failed.Type, failed.Error)
		}
	}

	if len(result.Created) == 0 && len(result.Skipped) == 0 && len(result.Failed) == 0 {
		fmt.Println("No tickets processed (dry-run or empty file)")
	}
}
