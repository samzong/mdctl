package llmstxt

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Fetch pages concurrently using a worker pool
func (g *Generator) fetchPages(urls []string) ([]PageInfo, error) {
	g.logger.Printf("Starting to fetch %d pages with concurrency %d", len(urls), g.config.Concurrency)

	// Create result and error channels
	resultChan := make(chan PageInfo, len(urls))
	errorChan := make(chan error, len(urls))

	// Create work channel, controlling concurrency
	workChan := make(chan string, len(urls))

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < g.config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for urlStr := range workChan {
				pageInfo, err := g.fetchPageContent(urlStr)
				if err != nil {
					g.logger.Printf("Warning: failed to fetch page %s: %v", urlStr, err)
					errorChan <- fmt.Errorf("failed to fetch page %s: %w", urlStr, err)
					continue
				}
				resultChan <- pageInfo
			}
		}()
	}

	// Send all URLs to work channel
	for _, urlStr := range urls {
		workChan <- urlStr
	}
	close(workChan)

	// Wait for all work to finish
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// Collect results
	var results []PageInfo
	for result := range resultChan {
		results = append(results, result)
		g.logger.Printf("Fetched page: %s", result.URL)
	}

	// Check for errors (don't interrupt processing, just log warnings)
	for err := range errorChan {
		g.logger.Printf("Warning: %v", err)
	}

	g.logger.Printf("Successfully fetched %d/%d pages", len(results), len(urls))

	return results, nil
}

// Get the content of a single page
func (g *Generator) fetchPageContent(urlStr string) (PageInfo, error) {
	// Set HTTP client
	client := &http.Client{
		Timeout: time.Duration(g.config.Timeout) * time.Second,
	}

	// Build request
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return PageInfo{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent
	req.Header.Set("User-Agent", g.config.UserAgent)

	// Send request
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return PageInfo{}, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PageInfo{}, fmt.Errorf("failed to fetch page, status code: %d", resp.StatusCode)
	}

	// Extract page information
	pageInfo, err := g.extractPageInfo(urlStr, resp)
	if err != nil {
		return PageInfo{}, fmt.Errorf("failed to extract page info: %w", err)
	}

	// Record timing information
	elapsed := time.Since(start).Round(time.Millisecond)
	g.logger.Printf("Fetched %s in %v", urlStr, elapsed)

	return pageInfo, nil
}
