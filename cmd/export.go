package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/samzong/mdctl/internal/exporter"
	"github.com/spf13/cobra"
)

var (
	exportFile          string
	exportDir           string
	siteType            string
	exportOutput        string
	exportTemplate      string
	exportFormat        string
	generateToc         bool
	shiftHeadingLevelBy int
	fileAsTitle         bool
	verbose             bool
	tocDepth            int
	navPath             string
	logger              *log.Logger

	exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export markdown files to other formats",
		Long: `Export markdown files to other formats like DOCX, PDF, EPUB.
Uses Pandoc as the underlying conversion tool.

Examples:
  mdctl export -f README.md -o output.docx
  mdctl export -d docs/ -o documentation.docx
  mdctl export -d docs/ -s mkdocs -o site_docs.docx
  mdctl export -d docs/ -o report.docx -t templates/corporate.docx
  mdctl export -d docs/ -o documentation.docx --shift-heading-level-by 2
  mdctl export -d docs/ -o documentation.docx --toc --toc-depth 4
  mdctl export -d docs/ -o documentation.pdf -F pdf`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 初始化日志
			if verbose {
				logger = log.New(os.Stdout, "[EXPORT] ", log.LstdFlags)
			} else {
				logger = log.New(io.Discard, "", 0)
			}

			logger.Println("Starting export process...")

			// 参数验证
			if exportFile == "" && exportDir == "" {
				return fmt.Errorf("either source file (-f) or source directory (-d) must be specified")
			}
			if exportFile != "" && exportDir != "" {
				return fmt.Errorf("cannot specify both source file (-f) and source directory (-d)")
			}
			if exportOutput == "" {
				return fmt.Errorf("output file (-o) must be specified")
			}

			logger.Printf("Validating parameters: file=%s, dir=%s, output=%s, format=%s, site-type=%s",
				exportFile, exportDir, exportOutput, exportFormat, siteType)

			// 检查 Pandoc 是否可用
			logger.Println("Checking Pandoc availability...")
			if err := exporter.CheckPandocAvailability(); err != nil {
				return err
			}
			logger.Println("Pandoc is available.")

			// 创建导出选项
			options := exporter.ExportOptions{
				Template:            exportTemplate,
				GenerateToc:         generateToc,
				ShiftHeadingLevelBy: shiftHeadingLevelBy,
				FileAsTitle:         fileAsTitle,
				Format:              exportFormat,
				SiteType:            siteType,
				Verbose:             verbose,
				Logger:              logger,
				TocDepth:            tocDepth,
				NavPath:             navPath,
			}

			logger.Printf("Export options: template=%s, toc=%v, toc-depth=%d, shift-heading=%d, file-as-title=%v",
				exportTemplate, generateToc, tocDepth, shiftHeadingLevelBy, fileAsTitle)

			// 执行导出
			exp := exporter.NewExporter()
			var err error

			if exportFile != "" {
				logger.Printf("Exporting single file: %s -> %s", exportFile, exportOutput)
				err = exp.ExportFile(exportFile, exportOutput, options)
			} else {
				logger.Printf("Exporting directory: %s -> %s", exportDir, exportOutput)
				err = exp.ExportDirectory(exportDir, exportOutput, options)
			}

			if err != nil {
				logger.Printf("Export failed: %s", err)
				return err
			}

			logger.Println("Export completed successfully.")
			return nil
		},
	}
)

func init() {
	exportCmd.Flags().StringVarP(&exportFile, "file", "f", "", "Source markdown file to export")
	exportCmd.Flags().StringVarP(&exportDir, "dir", "d", "", "Source directory containing markdown files to export")
	exportCmd.Flags().StringVarP(&siteType, "site-type", "s", "basic", "Site type (basic, mkdocs, hugo, docusaurus)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file path")
	exportCmd.Flags().StringVarP(&exportTemplate, "template", "t", "", "Word template file path")
	exportCmd.Flags().StringVarP(&exportFormat, "format", "F", "docx", "Output format (docx, pdf, epub)")
	exportCmd.Flags().BoolVar(&generateToc, "toc", false, "Generate table of contents")
	exportCmd.Flags().IntVar(&shiftHeadingLevelBy, "shift-heading-level-by", 0, "Shift heading level by N")
	exportCmd.Flags().BoolVar(&fileAsTitle, "file-as-title", false, "Use filename as section title")
	exportCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	exportCmd.Flags().IntVar(&tocDepth, "toc-depth", 3, "Depth of table of contents (default 3)")
	exportCmd.Flags().StringVarP(&navPath, "nav-path", "n", "", "Specify the navigation path to export (e.g. 'Section1/Subsection2')")
}
