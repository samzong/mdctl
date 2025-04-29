package llmstxt

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gobwas/glob"
)

// Sitemap XML structure
type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []struct {
		Loc        string `xml:"loc"`
		LastMod    string `xml:"lastmod,omitempty"`
		ChangeFreq string `xml:"changefreq,omitempty"`
		Priority   string `xml:"priority,omitempty"`
	} `xml:"url"`
}

// SitemapIndex XML structure
type SitemapIndex struct {
	XMLName  xml.Name `xml:"sitemapindex"`
	Sitemaps []struct {
		Loc     string `xml:"loc"`
		LastMod string `xml:"lastmod,omitempty"`
	} `xml:"sitemap"`
}

// Parse sitemap.xml file and return all URLs
func (g *Generator) parseSitemap() ([]string, error) {
	g.logger.Printf("Parsing sitemap from %s", g.config.SitemapURL)

	// Set HTTP client
	client := &http.Client{
		Timeout: time.Duration(g.config.Timeout) * time.Second,
	}

	// Build request
	req, err := http.NewRequest("GET", g.config.SitemapURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent
	req.Header.Set("User-Agent", g.config.UserAgent)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sitemap: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch sitemap, status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sitemap content: %w", err)
	}

	// Try to parse as standard sitemap
	var sitemap Sitemap
	if err := xml.Unmarshal(body, &sitemap); err == nil && len(sitemap.URLs) > 0 {
		g.logger.Println("Parsed standard sitemap")
		return g.extractURLsFromSitemap(sitemap), nil
	}

	// Try to parse as sitemap index
	var sitemapIndex SitemapIndex
	if err := xml.Unmarshal(body, &sitemapIndex); err == nil && len(sitemapIndex.Sitemaps) > 0 {
		g.logger.Println("Parsed sitemap index, fetching child sitemaps")
		return g.fetchSitemapIndex(sitemapIndex, client)
	}

	// If all parsing fails, try to handle as text sitemap (one URL per line)
	lines := string(body)
	if len(lines) > 0 {
		g.logger.Println("Parsing as text sitemap")
		return g.parseTextSitemap(lines), nil
	}

	return nil, fmt.Errorf("could not parse sitemap, unknown format")
}

// Extract URLs from standard sitemap
func (g *Generator) extractURLsFromSitemap(sitemap Sitemap) []string {
	urls := make([]string, 0, len(sitemap.URLs))
	for _, urlEntry := range sitemap.URLs {
		if urlEntry.Loc != "" {
			urls = append(urls, urlEntry.Loc)
		}
	}
	return urls
}

// Get all child sitemap URLs from sitemap index
func (g *Generator) fetchSitemapIndex(index SitemapIndex, client *http.Client) ([]string, error) {
	var allURLs []string

	for _, sitemapEntry := range index.Sitemaps {
		if sitemapEntry.Loc == "" {
			continue
		}

		g.logger.Printf("Fetching child sitemap: %s", sitemapEntry.Loc)

		// Build request
		req, err := http.NewRequest("GET", sitemapEntry.Loc, nil)
		if err != nil {
			g.logger.Printf("Warning: failed to create request for child sitemap %s: %v", sitemapEntry.Loc, err)
			continue
		}

		// Set User-Agent
		req.Header.Set("User-Agent", g.config.UserAgent)

		// Send request
		resp, err := client.Do(req)
		if err != nil {
			g.logger.Printf("Warning: failed to fetch child sitemap %s: %v", sitemapEntry.Loc, err)
			continue
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			g.logger.Printf("Warning: failed to read child sitemap %s: %v", sitemapEntry.Loc, err)
			continue
		}

		// Parse child sitemap
		var childSitemap Sitemap
		if err := xml.Unmarshal(body, &childSitemap); err != nil {
			g.logger.Printf("Warning: failed to parse child sitemap %s: %v", sitemapEntry.Loc, err)
			continue
		}

		// Extract URLs
		childURLs := g.extractURLsFromSitemap(childSitemap)
		g.logger.Printf("Found %d URLs in child sitemap %s", len(childURLs), sitemapEntry.Loc)
		allURLs = append(allURLs, childURLs...)
	}

	return allURLs, nil
}

// Parse text sitemap (one URL per line)
func (g *Generator) parseTextSitemap(content string) []string {
	lines := splitLines(content)
	var urls []string

	for _, line := range lines {
		line = normalizeURL(line)
		if isValidURL(line) {
			urls = append(urls, line)
		}
	}

	return urls
}

// Filter URLs based on include/exclude mode
func (g *Generator) filterURLs(urls []string) []string {
	if len(g.config.IncludePaths) == 0 && len(g.config.ExcludePaths) == 0 {
		return urls // No filtering rules, return directly
	}

	// Compile include/exclude mode
	var includeMatchers, excludeMatchers []glob.Glob
	for _, pattern := range g.config.IncludePaths {
		matcher, err := glob.Compile(pattern)
		if err != nil {
			g.logger.Printf("Warning: invalid include pattern '%s': %v", pattern, err)
			continue
		}
		includeMatchers = append(includeMatchers, matcher)
	}

	for _, pattern := range g.config.ExcludePaths {
		matcher, err := glob.Compile(pattern)
		if err != nil {
			g.logger.Printf("Warning: invalid exclude pattern '%s': %v", pattern, err)
			continue
		}
		excludeMatchers = append(excludeMatchers, matcher)
	}

	var filteredURLs []string
	for _, url := range urls {
		// If there are include rules, one of them must match
		if len(includeMatchers) > 0 {
			matched := false
			for _, matcher := range includeMatchers {
				if matcher.Match(url) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// If any exclude rules match, exclude
		excluded := false
		for _, matcher := range excludeMatchers {
			if matcher.Match(url) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		filteredURLs = append(filteredURLs, url)
	}

	return filteredURLs
}

// Helper function: split text by line
func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

// Helper function: normalize URL (remove spaces, etc.)
func normalizeURL(url string) string {
	return url
}

// Helper function: check if URL is valid
func isValidURL(url string) bool {
	return url != ""
}
