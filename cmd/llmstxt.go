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

Examples:
  # Standard mode
  mdctl llmstxt https://example.com/sitemap.xml > llms.txt

  # Full-content mode
  mdctl llmstxt -f https://example.com/sitemap.xml > llms-full.txt`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sitemapURL := args[0]

			// 创建生成器并配置选项
			config := llmstxt.GeneratorConfig{
				SitemapURL:   sitemapURL,
				IncludePaths: includePaths,
				ExcludePaths: excludePaths,
				FullMode:     fullMode,
				Concurrency:  concurrency,
				Timeout:      timeout,
				UserAgent:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
				Verbose:      verbose,
				VeryVerbose:  veryVerbose,
				MaxPages:     maxPages,
			}

			generator := llmstxt.NewGenerator(config)

			// 执行生成
			content, err := generator.Generate()
			if err != nil {
				return err
			}

			// 输出内容
			if outputPath == "" {
				// 输出到标准输出
				fmt.Println(content)
			} else {
				// 输出到文件
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
	llmstxtCmd.Flags().BoolVarP(&fullMode, "full", "f", false, "Enable full-content mode (extract page content)")
	llmstxtCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 5, "Number of concurrent requests")
	llmstxtCmd.Flags().IntVar(&timeout, "timeout", 30, "Request timeout in seconds")
	llmstxtCmd.Flags().IntVar(&maxPages, "max-pages", 0, "Maximum number of pages to process (0 for unlimited)")

	// 将命令添加到核心命令组
	llmstxtCmd.GroupID = "core"

	// 注册到rootCmd
	rootCmd.AddCommand(llmstxtCmd)
}
