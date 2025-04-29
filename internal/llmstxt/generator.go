package llmstxt

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"time"
)

// GeneratorConfig contains the configuration required to generate llms.txt
type GeneratorConfig struct {
	SitemapURL   string
	IncludePaths []string
	ExcludePaths []string
	FullMode     bool
	Concurrency  int
	Timeout      int
	UserAgent    string
	Verbose      bool
	VeryVerbose  bool // More detailed log output
	MaxPages     int  // Maximum number of pages to process, 0 means no limit
}

// PageInfo stores page information
type PageInfo struct {
	Title       string
	URL         string
	Description string
	Content     string // Page content, only filled in full mode
	Section     string // First segment of URL path as section
}

// Generator is the llms.txt generator
type Generator struct {
	config GeneratorConfig
	logger *log.Logger
}

// NewGenerator creates a new generator instance
func NewGenerator(config GeneratorConfig) *Generator {
	var logger *log.Logger
	if config.Verbose || config.VeryVerbose {
		logger = log.New(os.Stdout, "[LLMSTXT] ", log.LstdFlags)
	} else {
		logger = log.New(io.Discard, "", 0)
	}

	return &Generator{
		config: config,
		logger: logger,
	}
}

// Generate performs the generation process and returns the generated content
func (g *Generator) Generate() (string, error) {
	startTime := time.Now()
	g.logger.Printf("Starting generation for sitemap: %s", g.config.SitemapURL)
	if g.config.FullMode {
		g.logger.Println("Full-content mode enabled")
	}

	// 1. Parse sitemap.xml to get URL list
	urls, err := g.parseSitemap()
	if err != nil {
		return "", fmt.Errorf("failed to parse sitemap: %w", err)
	}
	g.logger.Printf("Found %d URLs in sitemap", len(urls))

	// 2. Filter URLs (based on include/exclude mode)
	urls = g.filterURLs(urls)
	g.logger.Printf("%d URLs after filtering", len(urls))

	// 2.1. Apply max page limit
	if g.config.MaxPages > 0 && len(urls) > g.config.MaxPages {
		g.logger.Printf("Limiting to %d pages as requested (--max-pages)", g.config.MaxPages)
		urls = urls[:g.config.MaxPages]
	}

	// 3. Create worker pool and get page info
	pages, err := g.fetchPages(urls)
	if err != nil {
		return "", fmt.Errorf("failed to fetch pages: %w", err)
	}

	// 4. Group pages by section
	sections := g.groupBySections(pages)

	// 5. Format to Markdown content
	content := g.formatContent(sections)

	elapsedTime := time.Since(startTime).Round(time.Millisecond)
	g.logger.Printf("Generation completed successfully in %v", elapsedTime)
	return content, nil
}

// Group pages by section
func (g *Generator) groupBySections(pages []PageInfo) map[string][]PageInfo {
	sections := make(map[string][]PageInfo)

	for _, page := range pages {
		sections[page.Section] = append(sections[page.Section], page)
	}

	// Sort pages within each section by URL path length
	for section, sectionPages := range sections {
		sort.Slice(sectionPages, func(i, j int) bool {
			return len(sectionPages[i].URL) < len(sectionPages[j].URL)
		})
		sections[section] = sectionPages
	}

	return sections
}

// Get sorted section name list, ensuring ROOT section is always first
func (g *Generator) getSortedSections(sections map[string][]PageInfo) []string {
	sectionNames := make([]string, 0, len(sections))

	// Add ROOT section first (if exists)
	if _, hasRoot := sections["ROOT"]; hasRoot {
		sectionNames = append(sectionNames, "ROOT")
	}

	// Add other sections and sort alphabetically
	for section := range sections {
		if section != "ROOT" {
			sectionNames = append(sectionNames, section)
		}
	}

	// Only sort if there are non-ROOT sections
	if len(sectionNames) > 1 {
		// Only sort non-ROOT sections
		nonRootSections := sectionNames[1:]
		sort.Strings(nonRootSections)
	}

	return sectionNames
}
