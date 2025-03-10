package exporter

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/samzong/mdctl/internal/exporter/formats"
	"github.com/samzong/mdctl/internal/exporter/sitereader"
)

// ExportOptions defines export options
type ExportOptions struct {
	Template            string      // Word template file path
	GenerateToc         bool        // Whether to generate table of contents
	ShiftHeadingLevelBy int         // Heading level offset
	FileAsTitle         bool        // Whether to use filename as section title
	Format              string      // Output format (docx, pdf, epub)
	SiteType            string      // Site type (mkdocs, hugo, docusaurus)
	Verbose             bool        // Whether to enable verbose logging
	Logger              *log.Logger // Logger
	SourceDirs          []string    // List of source directories for processing image paths
	TocDepth            int         // Table of contents depth, default is 3
	NavPath             string      // Specified navigation path to export
}

// Exporter defines exporter interface
type Exporter interface {
	ExportFile(input string, output string, options ExportOptions) error
	ExportDirectory(inputDir string, output string, options ExportOptions) error
}

// DefaultExporter is the default exporter implementation
type DefaultExporter struct {
	logger *log.Logger
}

// NewExporter creates a new exporter
func NewExporter() *DefaultExporter {
	return &DefaultExporter{
		logger: log.New(os.Stdout, "[EXPORTER] ", log.LstdFlags),
	}
}

// ExportFile exports a single Markdown file
func (e *DefaultExporter) ExportFile(input, output string, options ExportOptions) error {
	// Set logger
	if options.Logger != nil {
		e.logger = options.Logger
	} else if !options.Verbose {
		e.logger = log.New(io.Discard, "", 0)
	}

	e.logger.Printf("Exporting file: %s -> %s", input, output)

	// Check if file exists
	if _, err := os.Stat(input); os.IsNotExist(err) {
		e.logger.Printf("Error: input file does not exist: %s", input)
		return fmt.Errorf("input file does not exist: %s", input)
	}
	e.logger.Printf("Input file exists: %s", input)

	// Create output directory (if it doesn't exist)
	outputDir := filepath.Dir(output)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		e.logger.Printf("Error: failed to create output directory: %s", err)
		return fmt.Errorf("failed to create output directory: %s", err)
	}
	e.logger.Printf("Output directory created/verified: %s", outputDir)

	// Add source directory to SourceDirs
	sourceDir := filepath.Dir(input)
	if options.SourceDirs == nil {
		options.SourceDirs = []string{sourceDir}
	} else {
		// Check if already exists
		found := false
		for _, dir := range options.SourceDirs {
			if dir == sourceDir {
				found = true
				break
			}
		}
		if !found {
			options.SourceDirs = append(options.SourceDirs, sourceDir)
		}
	}
	e.logger.Printf("Added source directory to resource paths: %s", sourceDir)

	// Get formatter for the specified format
	e.logger.Printf("Getting formatter for format: %s", options.Format)
	formatter, err := formats.GetFormatter(options.Format, e.logger)
	if err != nil {
		e.logger.Printf("Error getting formatter: %s", err)
		return err
	}
	e.logger.Printf("Using formatter: %s", formatter.GetFormatName())

	// Validate options
	if err := formatter.ValidateOptions(convertOptions(options)); err != nil {
		e.logger.Printf("Error validating options: %s", err)
		return err
	}

	// Use formatter to export
	e.logger.Println("Starting export process...")
	err = formatter.Format(input, output, convertOptions(options))
	if err != nil {
		e.logger.Printf("Export failed: %s", err)
		return err
	}

	e.logger.Printf("File export completed successfully: %s", output)
	return nil
}

// ExportDirectory exports Markdown files in a directory
func (e *DefaultExporter) ExportDirectory(inputDir, output string, options ExportOptions) error {
	// Set logger
	if options.Logger != nil {
		e.logger = options.Logger
	} else if !options.Verbose {
		e.logger = log.New(io.Discard, "", 0)
	}

	e.logger.Printf("Exporting directory: %s -> %s", inputDir, output)

	// Check if directory exists
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		e.logger.Printf("Error: input directory does not exist: %s", inputDir)
		return fmt.Errorf("input directory does not exist: %s", inputDir)
	}
	e.logger.Printf("Input directory exists: %s", inputDir)

	// Create output directory (if it doesn't exist)
	outputDir := filepath.Dir(output)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		e.logger.Printf("Error: failed to create output directory: %s", err)
		return fmt.Errorf("failed to create output directory: %s", err)
	}
	e.logger.Printf("Output directory created/verified: %s", outputDir)

	// Initialize SourceDirs (if nil)
	if options.SourceDirs == nil {
		options.SourceDirs = []string{inputDir}
	} else {
		// Check if already exists
		found := false
		for _, dir := range options.SourceDirs {
			if dir == inputDir {
				found = true
				break
			}
		}
		if !found {
			options.SourceDirs = append(options.SourceDirs, inputDir)
		}
	}
	e.logger.Printf("Added input directory to resource paths: %s", inputDir)

	// Depending on site type, choose different processing
	var files []string
	var err error

	if options.SiteType != "" && options.SiteType != "basic" {
		// Use site reader to get file list
		e.logger.Printf("Using site reader for site type: %s", options.SiteType)
		reader, err := sitereader.GetSiteReader(options.SiteType, options.Verbose, e.logger)
		if err != nil {
			e.logger.Printf("Error getting site reader: %s", err)
			return err
		}

		// Detect if it's the specified type of site
		e.logger.Printf("Detecting if directory is a %s site...", options.SiteType)
		if !reader.Detect(inputDir) {
			e.logger.Printf("Error: directory %s does not appear to be a %s site", inputDir, options.SiteType)
			return fmt.Errorf("directory %s does not appear to be a %s site", inputDir, options.SiteType)
		}
		e.logger.Printf("Directory confirmed as %s site", options.SiteType)

		e.logger.Println("Reading site structure...")
		files, err = reader.ReadStructure(inputDir, "", options.NavPath)
		if err != nil {
			e.logger.Printf("Error reading site structure: %s", err)
			return err
		}
		e.logger.Printf("Found %d files in site structure", len(files))
	} else {
		// Basic directory mode: sort files by name
		e.logger.Println("Using basic directory mode, sorting files by name")
		files, err = GetMarkdownFilesInDir(inputDir)
		if err != nil {
			e.logger.Printf("Error getting markdown files: %s", err)
			return err
		}
		e.logger.Printf("Found %d markdown files in directory", len(files))
	}

	if len(files) == 0 {
		e.logger.Printf("Error: no markdown files found in directory: %s", inputDir)
		return fmt.Errorf("no markdown files found in directory: %s", inputDir)
	}

	// If there's only one file, export directly
	if len(files) == 1 {
		e.logger.Printf("Only one file found, exporting directly: %s", files[0])
		return e.ExportFile(files[0], output, options)
	}

	// Merge multiple files
	e.logger.Printf("Merging %d files...", len(files))
	merger := &Merger{
		ShiftHeadingLevelBy: options.ShiftHeadingLevelBy,
		FileAsTitle:         options.FileAsTitle,
		Logger:              e.logger,
		SourceDirs:          make([]string, 0),
		Verbose:             options.Verbose,
	}

	// Create temporary file
	e.logger.Println("Creating temporary file for merged content...")
	tempFile, err := os.CreateTemp("", "mdctl-merged-*.md")
	if err != nil {
		e.logger.Printf("Error creating temporary file: %s", err)
		return fmt.Errorf("failed to create temporary file: %s", err)
	}
	tempFilePath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempFilePath)
	e.logger.Printf("Temporary file created: %s", tempFilePath)

	// Merge files
	e.logger.Println("Merging files...")
	if err := merger.Merge(files, tempFilePath); err != nil {
		e.logger.Printf("Error merging files: %s", err)
		return fmt.Errorf("failed to merge files: %s", err)
	}
	e.logger.Println("Files merged successfully")

	// Add merger collected source directories to options
	if merger.SourceDirs != nil && len(merger.SourceDirs) > 0 {
		e.logger.Printf("Adding %d source directories from merger", len(merger.SourceDirs))
		for _, dir := range merger.SourceDirs {
			// Check if already exists
			found := false
			for _, existingDir := range options.SourceDirs {
				if existingDir == dir {
					found = true
					break
				}
			}
			if !found {
				options.SourceDirs = append(options.SourceDirs, dir)
				e.logger.Printf("Added source directory: %s", dir)
			}
		}
	}

	// Export merged file
	e.logger.Println("Exporting merged file...")
	return e.ExportFile(tempFilePath, output, options)
}

// GetMarkdownFilesInDir returns a list of markdown files in a directory, sorted by name
func GetMarkdownFilesInDir(dir string) ([]string, error) {
	var files []string

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %s", err)
	}

	// Filter markdown files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".md") || strings.HasSuffix(strings.ToLower(name), ".markdown") {
			files = append(files, filepath.Join(dir, name))
		}
	}

	// Sort files by name
	sort.Strings(files)

	return files, nil
}

// convertOptions 将内部ExportOptions转换为formats.ExportOptions
func convertOptions(options ExportOptions) formats.ExportOptions {
	return formats.ExportOptions{
		Template:            options.Template,
		GenerateToc:         options.GenerateToc,
		ShiftHeadingLevelBy: options.ShiftHeadingLevelBy,
		FileAsTitle:         options.FileAsTitle,
		Verbose:             options.Verbose,
		Logger:              options.Logger,
		SourceDirs:          options.SourceDirs,
		TocDepth:            options.TocDepth,
		NavPath:             options.NavPath,
	}
}

// CheckPandocAvailability 检查 Pandoc 是否可用
func CheckPandocAvailability() error {
	return formats.CheckPandocAvailability()
}
