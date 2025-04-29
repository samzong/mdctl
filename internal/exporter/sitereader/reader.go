package sitereader

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// SiteReader Define Site Reader Interface
type SiteReader interface {
	// Detect if given directory is this type of site
	Detect(dir string) bool

	// Read site structure, return sorted list of files
	// navPath parameter is used to specify the navigation path to export, empty to export all
	ReadStructure(dir string, configPath string, navPath string) ([]string, error)
}

// GetSiteReader Return the appropriate reader based on site type
func GetSiteReader(siteType string, verbose bool, logger *log.Logger) (SiteReader, error) {
	// If no logger is provided, create a default one
	if logger == nil {
		if verbose {
			logger = log.New(os.Stdout, "[SITE-READER] ", log.LstdFlags)
		} else {
			logger = log.New(io.Discard, "", 0)
		}
	}

	logger.Printf("Creating site reader for type: %s", siteType)

	switch siteType {
	case "mkdocs":
		logger.Println("Using MkDocs site reader")
		return &MkDocsReader{Logger: logger}, nil
	case "hugo":
		logger.Println("Hugo site type is not yet implemented")
		return nil, fmt.Errorf("hugo site type is not yet implemented")
	case "docusaurus":
		logger.Println("Docusaurus site type is not yet implemented")
		return nil, fmt.Errorf("docusaurus site type is not yet implemented")
	default:
		logger.Printf("Unsupported site type: %s", siteType)
		return nil, fmt.Errorf("unsupported site type: %s", siteType)
	}
}

// FindConfigFile Find config file in given directory
func FindConfigFile(dir string, configNames []string) (string, error) {
	// If no config file name is provided, use default values
	if len(configNames) == 0 {
		configNames = []string{"config.yml", "config.yaml"}
	}

	// Find config file
	for _, name := range configNames {
		configPath := filepath.Join(dir, name)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	return "", fmt.Errorf("no config file found in %s", dir)
}
