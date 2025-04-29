package exporter

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// PandocExporter Use Pandoc to export Markdown files
type PandocExporter struct {
	PandocPath string
	Logger     *log.Logger
}

// Export Use Pandoc to export Markdown files
func (e *PandocExporter) Export(input, output string, options ExportOptions) error {
	// If no logger is provided, create a default one
	if e.Logger == nil {
		if options.Verbose {
			e.Logger = log.New(os.Stdout, "[PANDOC] ", log.LstdFlags)
		} else {
			e.Logger = log.New(io.Discard, "", 0)
		}
	}

	e.Logger.Printf("Starting Pandoc export: %s -> %s", input, output)

	// Ensure output path is absolute
	absOutput, err := filepath.Abs(output)
	if err != nil {
		e.Logger.Printf("Failed to get absolute path for output: %s", err)
		return fmt.Errorf("failed to get absolute path for output: %s", err)
	}
	e.Logger.Printf("Using absolute output path: %s", absOutput)

	// Create a temporary file for sanitized content
	e.Logger.Println("Creating sanitized copy of input file...")
	tempFile, err := createSanitizedCopy(input, e.Logger)
	if err != nil {
		e.Logger.Printf("Failed to create sanitized copy: %s", err)
		return fmt.Errorf("failed to create sanitized copy: %s", err)
	}
	defer os.Remove(tempFile)
	e.Logger.Printf("Sanitized copy created: %s", tempFile)

	// Build Pandoc command arguments
	e.Logger.Println("Building Pandoc command arguments...")
	args := []string{
		tempFile,
		"-o", absOutput,
		"--standalone",
		"--pdf-engine=xelatex",
		"-V", "mainfont=SimSun", // Use SimSun as the main font
		"--wrap=preserve",
		"--embed-resources", // Embed resources into output file
	}

	// Add resource path parameters, helping Pandoc find images
	// Collect all possible resource paths
	resourcePaths := make(map[string]bool)

	// Add input file directory
	inputDir := filepath.Dir(input)
	resourcePaths[inputDir] = true
	e.Logger.Printf("Added input file directory to resource paths: %s", inputDir)

	// Add current working directory
	workingDir, err := os.Getwd()
	if err == nil {
		resourcePaths[workingDir] = true
		e.Logger.Printf("Added current working directory to resource paths: %s", workingDir)
	}

	// Add output file directory
	outputDir := filepath.Dir(absOutput)
	resourcePaths[outputDir] = true
	e.Logger.Printf("Added output file directory to resource paths: %s", outputDir)

	// Add source file directories to resource paths
	if len(options.SourceDirs) > 0 {
		for _, dir := range options.SourceDirs {
			resourcePaths[dir] = true
			e.Logger.Printf("Added source file directory to resource paths: %s", dir)
		}
	}

	// Add all resource paths to Pandoc arguments
	for path := range resourcePaths {
		args = append(args, "--resource-path", path)
	}

	// Add template parameter
	if options.Template != "" {
		e.Logger.Printf("Using template: %s", options.Template)
		args = append(args, "--reference-doc", options.Template)
	}

	// Add directory parameter
	if options.GenerateToc {
		e.Logger.Println("Generating table of contents")
		args = append(args, "--toc")

		// Add directory depth parameter
		if options.TocDepth > 0 {
			e.Logger.Printf("Setting table of contents depth to: %d", options.TocDepth)
			args = append(args, "--toc-depth", fmt.Sprintf("%d", options.TocDepth))
		}
	}

	// Add heading level offset parameter
	if options.ShiftHeadingLevelBy != 0 {
		e.Logger.Printf("Shifting heading levels by: %d", options.ShiftHeadingLevelBy)
		args = append(args, "--shift-heading-level-by", fmt.Sprintf("%d", options.ShiftHeadingLevelBy))
	}

	// Add specific parameters based on output format
	e.Logger.Printf("Using output format: %s", options.Format)
	switch options.Format {
	case "pdf":
		// PDF format needs special handling for Chinese
		e.Logger.Println("Adding PDF-specific parameters for CJK support")
		args = append(args,
			"-V", "CJKmainfont=SimSun", // CJK font settings
			"-V", "documentclass=article",
			"-V", "geometry=margin=1in")
	case "epub":
		// EPUB format specific parameters
		e.Logger.Println("Adding EPUB-specific parameters")
		args = append(args, "--epub-chapter-level=1")
	}

	// Execute Pandoc command
	e.Logger.Printf("Executing Pandoc command: %s %s", e.PandocPath, strings.Join(args, " "))
	cmd := exec.Command(e.PandocPath, args...)

	// Set working directory to input file directory, which helps Pandoc find relative paths for images
	cmd.Dir = inputDir

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		// If execution fails, try to look at input file content for debugging
		e.Logger.Printf("Pandoc execution failed: %s", err)
		e.Logger.Printf("Pandoc output: %s", string(outputBytes))

		inputContent, readErr := os.ReadFile(tempFile)
		if readErr == nil {
			// Only show the first 500 characters to avoid too much output
			contentPreview := string(inputContent)
			if len(contentPreview) > 500 {
				contentPreview = contentPreview[:500] + "..."
			}
			e.Logger.Printf("Input file preview:\n%s", contentPreview)
			return fmt.Errorf("pandoc execution failed: %s\nOutput: %s\nCommand: %s\nInput file preview:\n%s",
				err, string(outputBytes), strings.Join(cmd.Args, " "), contentPreview)
		}

		return fmt.Errorf("pandoc execution failed: %s\nOutput: %s\nCommand: %s",
			err, string(outputBytes), strings.Join(cmd.Args, " "))
	}

	e.Logger.Printf("Pandoc export completed successfully: %s", output)
	return nil
}

// createSanitizedCopy Create a sanitized temporary file copy
func createSanitizedCopy(inputFile string, logger *log.Logger) (string, error) {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}

	// Read input file content
	logger.Printf("Reading input file: %s", inputFile)
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return "", fmt.Errorf("failed to read input file: %s", err)
	}

	// Convert content to string
	contentStr := string(content)

	// Remove YAML front matter
	logger.Println("Removing YAML front matter...")
	yamlFrontMatterRegex := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n`)
	if yamlFrontMatterRegex.MatchString(contentStr) {
		logger.Println("YAML front matter found, removing it")
		contentStr = yamlFrontMatterRegex.ReplaceAllString(contentStr, "")
	}

	// Fix lines that may cause YAML parsing errors
	logger.Println("Fixing potential YAML parsing issues...")
	lines := strings.Split(contentStr, "\n")
	var cleanedLines []string
	fixedLines := 0

	for _, line := range lines {
		// Skip lines that may cause YAML parsing errors
		if strings.Contains(line, ":") && !strings.Contains(line, ": ") && !strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "\t") {
			// In this case, there should be a space after the colon, but there isn't, which may cause YAML parsing errors
			// Try to fix it
			fixedLine := strings.Replace(line, ":", ": ", 1)
			cleanedLines = append(cleanedLines, fixedLine)
			fixedLines++
			logger.Printf("Fixed line with missing space after colon: %s -> %s", line, fixedLine)
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "- ") && len(line) > 1 {
			// In this case, there should be a space after the dash, but there isn't, which may cause YAML parsing errors
			// Try to fix it
			fixedLine := strings.Replace(line, "-", "- ", 1)
			cleanedLines = append(cleanedLines, fixedLine)
			fixedLines++
			logger.Printf("Fixed line with missing space after dash: %s -> %s", line, fixedLine)
		} else {
			cleanedLines = append(cleanedLines, line)
		}
	}

	logger.Printf("Fixed %d lines with potential YAML issues", fixedLines)

	// Create a temporary file
	tempDir := os.TempDir()
	tempFilePath := filepath.Join(tempDir, "mdctl-sanitized-"+filepath.Base(inputFile))

	// Write sanitized content to temporary file
	logger.Printf("Writing sanitized content to temporary file: %s", tempFilePath)
	err = os.WriteFile(tempFilePath, []byte(strings.Join(cleanedLines, "\n")), 0644)
	if err != nil {
		return "", err
	}

	return tempFilePath, nil
}

// preprocessInputFile Preprocess input file, removing content that may cause Pandoc parsing errors
func preprocessInputFile(filePath string) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// Check for unconventional YAML front matter
	yamlFrontMatterRegex := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n`)
	if yamlFrontMatterRegex.MatchString(contentStr) {
		// Extract YAML front matter content
		matches := yamlFrontMatterRegex.FindStringSubmatch(contentStr)
		if len(matches) > 1 {
			yamlContent := matches[1]

			// Check if YAML content has formatting issues
			if strings.Contains(yamlContent, "\n-") && !strings.Contains(yamlContent, "\n- ") {
				// Fix formatting issue: ensure there's a space after the dash
				fixedYaml := strings.ReplaceAll(yamlContent, "\n-", "\n- ")
				fixedContent := strings.Replace(contentStr, yamlContent, fixedYaml, 1)

				// Write back to file
				return os.WriteFile(filePath, []byte(fixedContent), 0644)
			}
		}

		// If YAML format has other issues, remove entire front matter
		processedContent := yamlFrontMatterRegex.ReplaceAllString(contentStr, "")
		return os.WriteFile(filePath, []byte(processedContent), 0644)
	}

	return nil
}

// CheckPandocAvailability Check if Pandoc is available
func CheckPandocAvailability() error {
	cmd := exec.Command("pandoc", "--version")
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pandoc is not available: %s\n\nPlease install Pandoc to use the export feature:\n\n"+
			"macOS: brew install pandoc\n"+
			"Ubuntu/Debian: sudo apt-get install pandoc\n"+
			"Windows: choco install pandoc\n\n"+
			"For more information, visit: https://pandoc.org/installing.html", err)
	}

	// Check version
	versionStr := string(outputBytes)
	if !strings.Contains(versionStr, "pandoc") {
		return fmt.Errorf("unexpected pandoc version output: %s", versionStr)
	}

	return nil
}
