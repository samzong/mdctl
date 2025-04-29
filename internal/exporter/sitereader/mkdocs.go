package sitereader

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type MkDocsReader struct {
	Logger *log.Logger
}

type MkDocsConfig struct {
	Docs    []string `yaml:"nav"`
	DocsDir string   `yaml:"docs_dir"`
	Inherit string   `yaml:"INHERIT"`
}

func (r *MkDocsReader) Detect(dir string) bool {
	// Setting up the Logger
	if r.Logger == nil {
		r.Logger = log.New(io.Discard, "", 0)
	}

	// Check if mkdocs.yml file exists
	mkdocsPath := filepath.Join(dir, "mkdocs.yml")
	if _, err := os.Stat(mkdocsPath); os.IsNotExist(err) {
		// Try mkdocs.yaml
		mkdocsPath = filepath.Join(dir, "mkdocs.yaml")
		if _, err := os.Stat(mkdocsPath); os.IsNotExist(err) {
			r.Logger.Printf("No mkdocs.yml or mkdocs.yaml found in %s", dir)
			return false
		}
	}

	r.Logger.Printf("Found MkDocs configuration file: %s", mkdocsPath)
	return true
}

func (r *MkDocsReader) ReadStructure(dir string, configPath string, navPath string) ([]string, error) {
	// Setting up the Logger
	if r.Logger == nil {
		r.Logger = log.New(io.Discard, "", 0)
	}

	r.Logger.Printf("Reading MkDocs site structure from: %s", dir)
	if navPath != "" {
		r.Logger.Printf("Filtering by navigation path: %s", navPath)
	}

	// Find config file
	if configPath == "" {
		configNames := []string{"mkdocs.yml", "mkdocs.yaml"}
		var err error
		configPath, err = FindConfigFile(dir, configNames)
		if err != nil {
			r.Logger.Printf("Failed to find MkDocs config file: %s", err)
			return nil, fmt.Errorf("failed to find MkDocs config file: %s", err)
		}
	}
	r.Logger.Printf("Using config file: %s", configPath)

	// Read and parse config file, including handling INHERIT
	config, err := r.readAndMergeConfig(configPath, dir)
	if err != nil {
		r.Logger.Printf("Failed to read config file: %s", err)
		return nil, fmt.Errorf("failed to read config file: %s", err)
	}

	// Get docs directory
	docsDir := "docs"
	if docsDirValue, ok := config["docs_dir"]; ok {
		if docsDirStr, ok := docsDirValue.(string); ok {
			docsDir = docsDirStr
		}
	}
	docsDir = filepath.Join(dir, docsDir)
	r.Logger.Printf("Using docs directory: %s", docsDir)

	// Parse navigation structure
	var nav interface{}
	if navValue, ok := config["nav"]; ok {
		nav = navValue
	} else {
		// If no navigation config, try to find all Markdown files
		r.Logger.Println("No navigation configuration found, searching for all markdown files")
		return getAllMarkdownFiles(docsDir)
	}

	// Parse navigation structure, get file list
	files, err := parseNavigation(nav, docsDir, navPath)
	if err != nil {
		r.Logger.Printf("Failed to parse navigation: %s", err)
		return nil, fmt.Errorf("failed to parse navigation: %s", err)
	}

	r.Logger.Printf("Found %d files in navigation", len(files))
	return files, nil
}

// readAndMergeConfig Read and merge MkDocs config file, handling INHERIT directive
func (r *MkDocsReader) readAndMergeConfig(configPath string, baseDir string) (map[string]interface{}, error) {
	r.Logger.Printf("Reading and merging config file: %s", configPath)

	// Read main config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		r.Logger.Printf("Failed to read MkDocs config file: %s", err)
		return nil, fmt.Errorf("failed to read MkDocs config file: %s", err)
	}

	// Parse config file
	var config map[string]interface{}
	if err := yaml.Unmarshal(configData, &config); err != nil {
		r.Logger.Printf("Failed to parse MkDocs config file: %s", err)
		return nil, fmt.Errorf("failed to parse MkDocs config file: %s", err)
	}

	// Check if there's an INHERIT directive
	inheritValue, hasInherit := config["INHERIT"]
	if !hasInherit {
		// No inherit, return current config
		return config, nil
	}

	// Handle INHERIT directive
	inheritPath, ok := inheritValue.(string)
	if !ok {
		r.Logger.Printf("Invalid INHERIT value, expected string but got: %T", inheritValue)
		return nil, fmt.Errorf("invalid INHERIT value, expected string")
	}

	r.Logger.Printf("Found INHERIT directive pointing to: %s", inheritPath)

	// Parse inherit path, may be relative to current config file
	configDir := filepath.Dir(configPath)
	inheritFullPath := filepath.Join(configDir, inheritPath)

	// Read inherited config file
	inheritConfig, err := r.readAndMergeConfig(inheritFullPath, baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read inherited config file %s: %s", inheritFullPath, err)
	}

	// Merge config, current config takes precedence
	mergedConfig := make(map[string]interface{})

	// Copy inherit config first
	for k, v := range inheritConfig {
		mergedConfig[k] = v
	}

	// Override current config
	for k, v := range config {
		if k != "INHERIT" { // Don't copy INHERIT directive
			mergedConfig[k] = v
		}
	}

	r.Logger.Printf("Successfully merged config with inherited file")
	return mergedConfig, nil
}

// preprocessMarkdownFile Preprocess Markdown file, remove YAML front matter that may cause problems
func preprocessMarkdownFile(filePath string) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Check if there's YAML front matter
	contentStr := string(content)
	yamlFrontMatterRegex := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n`)

	// If there's YAML front matter, remove it
	if yamlFrontMatterRegex.MatchString(contentStr) {
		// Create temp file
		tempFile, err := os.CreateTemp("", "mdctl-*.md")
		if err != nil {
			return err
		}
		tempFilePath := tempFile.Name()
		tempFile.Close()

		// Remove YAML front matter
		processedContent := yamlFrontMatterRegex.ReplaceAllString(contentStr, "")

		// Write processed content to temp file
		if err := os.WriteFile(tempFilePath, []byte(processedContent), 0644); err != nil {
			os.Remove(tempFilePath)
			return err
		}

		// Replace original file
		if err := os.Rename(tempFilePath, filePath); err != nil {
			os.Remove(tempFilePath)
			return err
		}
	}

	return nil
}

// parseNavigation Parse MkDocs navigation structure
func parseNavigation(nav interface{}, docsDir string, navPath string) ([]string, error) {
	var files []string

	switch v := nav.(type) {
	case []interface{}:
		// Navigation is a list
		for _, item := range v {
			itemFiles, err := parseNavigation(item, docsDir, navPath)
			if err != nil {
				return nil, err
			}
			files = append(files, itemFiles...)
		}
	case map[string]interface{}:
		// Navigation is a map
		for title, value := range v {
			// If nav path is specified, check if current node title matches
			if navPath != "" {
				// Support simple path matching, e.g. "Section1/Subsection2"
				navParts := strings.Split(navPath, "/")
				if strings.TrimSpace(title) == strings.TrimSpace(navParts[0]) {
					// If it's a multi-level path, continue matching the next level
					if len(navParts) > 1 {
						subNavPath := strings.Join(navParts[1:], "/")
						itemFiles, err := parseNavigation(value, docsDir, subNavPath)
						if err != nil {
							return nil, err
						}
						files = append(files, itemFiles...)
						continue
					} else {
						// If it's a single-level path and matches, only handle this node
						itemFiles, err := parseNavigation(value, docsDir, "")
						if err != nil {
							return nil, err
						}
						files = append(files, itemFiles...)
						continue
					}
				} else {
					// Title doesn't match, skip this node
					continue
				}
			}

			// If no nav path is specified or already matched the path, handle normally
			itemFiles, err := parseNavigation(value, docsDir, "")
			if err != nil {
				return nil, err
			}
			files = append(files, itemFiles...)
		}
	case string:
		// Navigation item is a file path
		if strings.HasSuffix(v, ".md") {
			filePath := filepath.Join(docsDir, v)
			if _, err := os.Stat(filePath); err == nil {
				// If no nav path is specified or already handled in nav path filtering, add file
				if navPath == "" {
					files = append(files, filePath)
				}
			}
		}
	}

	return files, nil
}

// getAllMarkdownFiles Get all Markdown files in a directory
func getAllMarkdownFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".md" || ext == ".markdown" {
				files = append(files, path)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %s", dir, err)
	}

	return files, nil
}
