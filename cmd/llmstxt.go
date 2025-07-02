package cmd

import (
	"fmt"
	"os"

	"github.com/samzong/mdctl/internal/llmstxt"
	"github.com/spf13/cobra"
)

var (
	includePaths []string
	excludePaths []string
	ignorePaths  []string
	outputPath   string
	fullMode     bool
	concurrency  int
	timeout      int
	maxPages     int

	llmstxtCmd = &cobra.Command{
		Use:   "llmstxt [url]",
		Short: "Generate llms.txt from sitemap.xml",
		Long: `Generate a llms.txt file from a website's sitemap.xml. This file is a curated 
list of the website's pages in markdown format, perfect for training or fine-tuning 
language models.

In standard mode, only title and description are extracted. In full mode (-f flag), 
the content of each page is also extracted.

Use include/exclude patterns or ignore patterns to filter specific pages.

Examples:
  # Standard mode
  mdctl llmstxt https://example.com/sitemap.xml > llms.txt

  # Full-content mode
  mdctl llmstxt -f https://example.com/sitemap.xml > llms-full.txt

  # Filter out unwanted pages using ignore patterns
  mdctl llmstxt --ignore "*/admin/*" --ignore "*/private/*" https://example.com/sitemap.xml > llms.txt`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sitemapURL := args[0]

			// Create a generator and configure options
			// Combine excludePaths and ignorePaths for backward compatibility
			allExcludePaths := append(excludePaths, ignorePaths...)

			config := llmstxt.GeneratorConfig{
				SitemapURL:   sitemapURL,
				IncludePaths: includePaths,
				ExcludePaths: allExcludePaths,
				FullMode:     fullMode,
				Concurrency:  concurrency,
				Timeout:      timeout,
				UserAgent:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
				Verbose:      verbose,
				VeryVerbose:  veryVerbose,
				MaxPages:     maxPages,
			}

			generator := llmstxt.NewGenerator(config)

			// Execute generation
			content, err := generator.Generate()
			if err != nil {
				return err
			}

			// Output content
			if outputPath == "" {
				// Output to standard output
				fmt.Println(content)
			} else {
				// Output to file
				return os.WriteFile(outputPath, []byte(content), 0644)
			}

			return nil
		},
	}
)

func init() {
	llmstxtCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: stdout)")
	llmstxtCmd.Flags().StringSliceVarP(&includePaths, "include-path", "i", []string{}, "Glob patterns for paths to include (can be specified multiple times)")
	llmstxtCmd.Flags().StringSliceVarP(&excludePaths, "exclude-path", "e", []string{}, "Glob patterns for paths to exclude (can be specified multiple times)")
	llmstxtCmd.Flags().StringSliceVar(&ignorePaths, "ignore", []string{}, "Glob patterns for paths to ignore (can be specified multiple times)")
	llmstxtCmd.Flags().BoolVarP(&fullMode, "full", "f", false, "Enable full-content mode (extract page content)")
	llmstxtCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 5, "Number of concurrent requests")
	llmstxtCmd.Flags().IntVar(&timeout, "timeout", 30, "Request timeout in seconds")
	llmstxtCmd.Flags().IntVar(&maxPages, "max-pages", 0, "Maximum number of pages to process (0 for unlimited)")

	// Add command to core group
	llmstxtCmd.GroupID = "core"

	rootCmd.AddCommand(llmstxtCmd)
}
