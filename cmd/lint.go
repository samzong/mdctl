package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samzong/mdctl/internal/linter"
	"github.com/spf13/cobra"
)

var (
	autoFix     bool
	configRules []string
	outputFormat string
	rulesFile   string
	enableRules []string
	disableRules []string
)

var lintCmd = &cobra.Command{
	Use:   "lint [files...]",
	Short: "Lint markdown files for syntax issues",
	Long: `Lint markdown files using markdownlint rules to find syntax issues.

This command will scan markdown files and report any syntax issues found.
It can also automatically fix issues when --fix flag is used.

Examples:
  # Lint a single file
  mdctl lint README.md

  # Lint multiple files
  mdctl lint docs/*.md

  # Lint with auto-fix
  mdctl lint --fix README.md

  # Lint with custom rules configuration
  mdctl lint --config .markdownlint.json README.md

  # Enable specific rules
  mdctl lint --enable MD001,MD003 README.md

  # Disable specific rules
  mdctl lint --disable MD013,MD033 README.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("at least one markdown file must be specified")
		}

		// Expand file patterns
		var files []string
		for _, arg := range args {
			matches, err := filepath.Glob(arg)
			if err != nil {
				return fmt.Errorf("invalid file pattern %s: %v", arg, err)
			}
			if len(matches) == 0 {
				// If no glob matches, check if it's a direct file
				if _, err := os.Stat(arg); err == nil {
					files = append(files, arg)
				} else {
					fmt.Printf("Warning: No files found matching pattern: %s\n", arg)
				}
			} else {
				files = append(files, matches...)
			}
		}

		// Filter for markdown files
		var markdownFiles []string
		for _, file := range files {
			if strings.HasSuffix(strings.ToLower(file), ".md") || strings.HasSuffix(strings.ToLower(file), ".markdown") {
				markdownFiles = append(markdownFiles, file)
			}
		}

		if len(markdownFiles) == 0 {
			return fmt.Errorf("no markdown files found")
		}

		// Create linter configuration
		config := &linter.Config{
			AutoFix:      autoFix,
			OutputFormat: outputFormat,
			RulesFile:    rulesFile,
			EnableRules:  enableRules,
			DisableRules: disableRules,
			Verbose:      verbose,
		}

		// Create linter instance
		mdLinter := linter.New(config)

		// Process files
		var totalIssues int
		var totalFixed int
		
		for _, file := range markdownFiles {
			if verbose {
				fmt.Printf("Linting: %s\n", file)
			}

			result, err := mdLinter.LintFile(file)
			if err != nil {
				fmt.Printf("Error linting %s: %v\n", file, err)
				continue
			}

			totalIssues += len(result.Issues)
			totalFixed += result.FixedCount

			// Display results based on output format
			if err := displayResults(file, result, config); err != nil {
				return fmt.Errorf("error displaying results: %v", err)
			}
		}

		// Summary
		if verbose || len(markdownFiles) > 1 {
			fmt.Printf("\nSummary:\n")
			fmt.Printf("  Files processed: %d\n", len(markdownFiles))
			fmt.Printf("  Total issues: %d\n", totalIssues)
			if autoFix {
				fmt.Printf("  Issues fixed: %d\n", totalFixed)
			}
		}

		// Exit with error code if issues found and not in fix mode
		if totalIssues > 0 && !autoFix {
			os.Exit(1)
		}

		return nil
	},
}

func displayResults(filename string, result *linter.Result, config *linter.Config) error {
	switch config.OutputFormat {
	case "json":
		return displayJSONResults(filename, result)
	case "github":
		return displayGitHubResults(filename, result)
	default:
		return displayDefaultResults(filename, result, config)
	}
}

func displayDefaultResults(filename string, result *linter.Result, config *linter.Config) error {
	if len(result.Issues) == 0 {
		if config.Verbose {
			fmt.Printf("✓ %s: No issues found\n", filename)
		}
		return nil
	}

	fmt.Printf("%s:\n", filename)
	for _, issue := range result.Issues {
		status := "✗"
		if issue.Fixed {
			status = "✓"
		}
		
		fmt.Printf("  %s Line %d: %s (%s)\n", 
			status, issue.Line, issue.Message, issue.Rule)
		
		if config.Verbose && issue.Context != "" {
			fmt.Printf("    Context: %s\n", issue.Context)
		}
	}

	if config.AutoFix && result.FixedCount > 0 {
		fmt.Printf("  Fixed %d issues\n", result.FixedCount)
	}

	return nil
}

func displayJSONResults(filename string, result *linter.Result) error {
	// TODO: Implement JSON output format
	fmt.Printf("JSON output not yet implemented\n")
	return nil
}

func displayGitHubResults(filename string, result *linter.Result) error {
	// TODO: Implement GitHub Actions output format
	fmt.Printf("GitHub Actions output not yet implemented\n")
	return nil
}

func init() {
	lintCmd.Flags().BoolVar(&autoFix, "fix", false, "Automatically fix issues where possible")
	lintCmd.Flags().StringVar(&outputFormat, "format", "default", "Output format: default, json, github")
	lintCmd.Flags().StringVar(&rulesFile, "config", "", "Path to markdownlint configuration file")
	lintCmd.Flags().StringSliceVar(&enableRules, "enable", []string{}, "Enable specific rules (comma-separated)")
	lintCmd.Flags().StringSliceVar(&disableRules, "disable", []string{}, "Disable specific rules (comma-separated)")

	lintCmd.GroupID = "core"
}