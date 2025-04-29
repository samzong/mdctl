package exporter

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// Merger Merge multiple Markdown files
type Merger struct {
	ShiftHeadingLevelBy int
	FileAsTitle         bool
	Logger              *log.Logger
	// Store all source directories, used to set Pandoc's resource paths
	SourceDirs []string
	// Whether to enable verbose logging
	Verbose bool
}

// Merge Merge multiple Markdown files into a single target file
func (m *Merger) Merge(sources []string, target string) error {
	// If no logger is provided, create a default one
	if m.Logger == nil {
		if m.Verbose {
			m.Logger = log.New(os.Stdout, "[MERGER] ", log.LstdFlags)
		} else {
			m.Logger = log.New(io.Discard, "", 0)
		}
	}

	if len(sources) == 0 {
		m.Logger.Println("Error: no source files provided")
		return fmt.Errorf("no source files provided")
	}

	m.Logger.Printf("Merging %d files into: %s", len(sources), target)
	var mergedContent strings.Builder

	// Initialize source directory list
	m.SourceDirs = make([]string, 0, len(sources))
	sourceDirsMap := make(map[string]bool) // Used for deduplication

	// Process each source file
	for i, source := range sources {
		m.Logger.Printf("Processing file %d/%d: %s", i+1, len(sources), source)

		// Get source file's directory and add to list (deduplication)
		sourceDir := filepath.Dir(source)
		if !sourceDirsMap[sourceDir] {
			sourceDirsMap[sourceDir] = true
			m.SourceDirs = append(m.SourceDirs, sourceDir)
		}

		// Read file content
		content, err := os.ReadFile(source)
		if err != nil {
			m.Logger.Printf("Error reading file %s: %s", source, err)
			return fmt.Errorf("failed to read file %s: %s", source, err)
		}

		// Process content
		processedContent := string(content)

		// Ensure content is valid UTF-8
		if !utf8.ValidString(processedContent) {
			m.Logger.Printf("File %s contains invalid UTF-8, attempting to convert from GBK", source)
			// Attempt to convert content from GBK to UTF-8
			reader := transform.NewReader(bytes.NewReader(content), simplifiedchinese.GBK.NewDecoder())
			decodedContent, err := io.ReadAll(reader)
			if err != nil {
				m.Logger.Printf("Failed to decode content from file %s: %s", source, err)
				return fmt.Errorf("failed to decode content from file %s: %s", source, err)
			}
			processedContent = string(decodedContent)
			m.Logger.Printf("Successfully converted content from GBK to UTF-8")
		}

		// Remove YAML front matter
		m.Logger.Println("Removing YAML front matter...")
		processedContent = removeYAMLFrontMatter(processedContent)

		// Process image paths
		m.Logger.Println("Processing image paths...")
		processedContent, err = processImagePaths(processedContent, source, m.Logger, m.Verbose)
		if err != nil {
			m.Logger.Printf("Error processing image paths: %s", err)
			return fmt.Errorf("failed to process image paths: %s", err)
		}

		// Adjust heading levels
		if m.ShiftHeadingLevelBy != 0 {
			m.Logger.Printf("Shifting heading levels by %d", m.ShiftHeadingLevelBy)
			processedContent = ShiftHeadings(processedContent, m.ShiftHeadingLevelBy)
		}

		// Add filename as title
		if m.FileAsTitle {
			filename := filepath.Base(source)
			m.Logger.Printf("Adding filename as title: %s", filename)
			processedContent = AddTitleFromFilename(processedContent, filename, 1+m.ShiftHeadingLevelBy)
		}

		// Add to merged content
		m.Logger.Printf("Adding processed content to merged result (length: %d bytes)", len(processedContent))
		mergedContent.WriteString(processedContent)

		// If not the last file, add separator
		if i < len(sources)-1 {
			mergedContent.WriteString("\n\n")
		}
	}

	// Final content
	finalContent := mergedContent.String()

	// Check again for any YAML-related issues
	m.Logger.Println("Sanitizing final content...")
	finalContent = sanitizeContent(finalContent)

	// Write target file, ensuring UTF-8 encoding
	m.Logger.Printf("Writing merged content to target file: %s (size: %d bytes)", target, len(finalContent))
	err := os.WriteFile(target, []byte(finalContent), 0644)
	if err != nil {
		m.Logger.Printf("Error writing merged content: %s", err)
		return fmt.Errorf("failed to write merged content to %s: %s", target, err)
	}

	m.Logger.Printf("Successfully merged %d files into: %s", len(sources), target)
	return nil
}

// processImagePaths Process image paths in Markdown, converting relative paths to paths relative to the command execution location
func processImagePaths(content, sourcePath string, logger *log.Logger, verbose bool) (string, error) {
	// If no logger is provided, create a default one
	if logger == nil {
		if verbose {
			logger = log.New(os.Stdout, "[IMAGE] ", log.LstdFlags)
		} else {
			logger = log.New(io.Discard, "", 0)
		}
	}

	// Get source file's directory
	sourceDir := filepath.Dir(sourcePath)
	if verbose {
		logger.Printf("Processing image paths: source file directory = %s", sourceDir)
	}

	// Get current working directory (location of command execution)
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get current working directory: %v", err)
	}
	if verbose {
		logger.Printf("Current working directory = %s", workingDir)
	}

	// Get absolute path of source file's directory
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return "", fmt.Errorf("unable to get absolute path of source file's directory: %v", err)
	}
	if verbose {
		logger.Printf("Source file's directory absolute path = %s", absSourceDir)
	}

	// Match Markdown image syntax: ![alt](path)
	imageRegex := regexp.MustCompile(`!\[(.*?)\]\((.*?)\)`)

	// Replace all image paths
	processedContent := imageRegex.ReplaceAllStringFunc(content, func(match string) string {
		// Extract image path
		submatches := imageRegex.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match // If match is incorrect, keep as-is
		}

		altText := submatches[1]
		imagePath := submatches[2]
		if verbose {
			logger.Printf("Found image: alt = %s, path = %s", altText, imagePath)
		}

		// If image is a web image (starts with http:// or https://), keep as-is
		if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
			if verbose {
				logger.Printf("Keeping web image path: %s", imagePath)
			}
			return match
		}

		// Parse image's absolute path
		var absoluteImagePath string
		if filepath.IsAbs(imagePath) {
			absoluteImagePath = imagePath
		} else {
			// For relative paths, convert to absolute path first
			absoluteImagePath = filepath.Join(absSourceDir, imagePath)
		}
		if verbose {
			logger.Printf("Image path: relative path = %s, absolute path = %s", imagePath, absoluteImagePath)
		}

		// Check if image file exists
		if _, err := os.Stat(absoluteImagePath); os.IsNotExist(err) {
			if verbose {
				logger.Printf("Image does not exist: %s", absoluteImagePath)
			}
			// Image does not exist, try to find it in adjacent directories
			// For example, if path is ../images/image.png, try to find it in the images subdirectory of the parent directory of the source file's directory
			if strings.HasPrefix(imagePath, "../") {
				parentDir := filepath.Dir(absSourceDir)
				relPath := strings.TrimPrefix(imagePath, "../")
				alternativePath := filepath.Join(parentDir, relPath)
				if verbose {
					logger.Printf("Trying alternative path: %s", alternativePath)
				}
				if _, err := os.Stat(alternativePath); err == nil {
					absoluteImagePath = alternativePath
					if verbose {
						logger.Printf("Found image in alternative path: %s", absoluteImagePath)
					}
				} else {
					// Still not found, keep as-is
					if verbose {
						logger.Printf("Image does not exist in alternative path: %s", alternativePath)
					}
					return match
				}
			} else {
				// Image not found, keep as-is
				return match
			}
		}

		// Calculate image's path relative to current working directory
		relPath, err := filepath.Rel(workingDir, absoluteImagePath)
		if err != nil {
			if verbose {
				logger.Printf("Unable to calculate relative path, keeping original path: %s, error: %v", imagePath, err)
			}
			return match
		}

		// Update image reference with path relative to current working directory
		newRef := fmt.Sprintf("![%s](%s)", altText, relPath)
		if verbose {
			logger.Printf("Updating image reference: %s -> %s", match, newRef)
		}
		return newRef
	})

	return processedContent, nil
}

// removeYAMLFrontMatter Remove YAML front matter
func removeYAMLFrontMatter(content string) string {
	// Match YAML front matter
	yamlFrontMatterRegex := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n`)
	return yamlFrontMatterRegex.ReplaceAllString(content, "")
}

// sanitizeContent Clean content, removing content that may cause Pandoc parsing errors
func sanitizeContent(content string) string {
	// Remove lines that may cause YAML parsing errors
	lines := strings.Split(content, "\n")
	var cleanedLines []string

	for _, line := range lines {
		// Skip lines that may cause YAML parsing errors
		if strings.Contains(line, ":") && !strings.Contains(line, ": ") && !strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "\t") {
			// In this case, there should be a space after the colon, but there isn't, which may cause YAML parsing errors
			// Try to fix it
			fixedLine := strings.Replace(line, ":", ": ", 1)
			cleanedLines = append(cleanedLines, fixedLine)
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "- ") && len(line) > 1 {
			// In this case, there should be a space after the dash, but there isn't, which may cause YAML parsing errors
			// Try to fix it
			fixedLine := strings.Replace(line, "-", "- ", 1)
			cleanedLines = append(cleanedLines, fixedLine)
		} else {
			cleanedLines = append(cleanedLines, line)
		}
	}

	return strings.Join(cleanedLines, "\n")
}
