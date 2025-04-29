package llmstxt

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Extract page information from HTML content
func (g *Generator) extractPageInfo(urlStr string, resp *http.Response) (PageInfo, error) {
	// Create PageInfo object
	pageInfo := PageInfo{
		URL:     urlStr,
		Section: parseSection(urlStr),
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return pageInfo, err
	}

	// Extract title
	pageInfo.Title = extractTitle(doc)
	if g.config.VeryVerbose {
		g.logger.Printf("Extracted title from %s: %s", urlStr, pageInfo.Title)
	}

	if pageInfo.Title == "" {
		// If title cannot be extracted, use the last segment of the URL as the title
		pageInfo.Title = extractTitleFromURL(urlStr)
		if g.config.VeryVerbose {
			g.logger.Printf("Could not extract title, using URL-based title instead: %s", pageInfo.Title)
		}
	}

	// Extract description
	pageInfo.Description = extractDescription(doc)
	if g.config.VeryVerbose {
		g.logger.Printf("Extracted description from %s: %s", urlStr, truncateString(pageInfo.Description, 100))
	}

	// Extract content in full mode
	if g.config.FullMode {
		if g.config.VeryVerbose {
			g.logger.Printf("Extracting full content from %s", urlStr)
		}
		pageInfo.Content = extractContent(doc)
		if g.config.VeryVerbose {
			contentLen := len(pageInfo.Content)
			preview := truncateString(pageInfo.Content, 100)
			g.logger.Printf("Extracted content from %s (%d chars): %s", urlStr, contentLen, preview)
		}
	}

	return pageInfo, nil
}

// Helper function: truncate string and add ellipsis
func truncateString(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Extract section information from URL
func parseSection(urlStr string) string {
	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "ROOT"
	}

	// Split path
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")

	// If path is empty, return ROOT
	if len(pathParts) == 0 || pathParts[0] == "" {
		return "ROOT"
	}

	// Return first segment of path
	return pathParts[0]
}

// Extract title from HTML document
func extractTitle(doc *goquery.Document) string {
	// Try to extract from title tag
	title := doc.Find("title").First().Text()
	title = strings.TrimSpace(title)

	// If no title tag, try to extract from h1 tag
	if title == "" {
		title = doc.Find("h1").First().Text()
		title = strings.TrimSpace(title)
	}

	return title
}

// Extract title from URL
func extractTitleFromURL(urlStr string) string {
	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	// Get the last segment of the path
	basename := path.Base(parsedURL.Path)

	// Remove file extension
	basename = strings.TrimSuffix(basename, path.Ext(basename))

	// If basename is empty or is "/", use hostname
	if basename == "" || basename == "." || basename == "/" {
		return parsedURL.Hostname()
	}

	// Replace hyphens and underscores with spaces, and capitalize
	basename = strings.ReplaceAll(basename, "-", " ")
	basename = strings.ReplaceAll(basename, "_", " ")

	return strings.Title(basename)
}

// Extract description from HTML document
func extractDescription(doc *goquery.Document) string {
	var description string

	// Try meta description
	description, _ = doc.Find("meta[name='description']").Attr("content")
	if description != "" {
		return strings.TrimSpace(description)
	}

	// Try og:description
	description, _ = doc.Find("meta[property='og:description']").Attr("content")
	if description != "" {
		return strings.TrimSpace(description)
	}

	// Try twitter:description
	description, _ = doc.Find("meta[name='twitter:description']").Attr("content")
	if description != "" {
		return strings.TrimSpace(description)
	}

	// If none found, extract first text
	description = doc.Find("p").First().Text()
	if description != "" {
		// Limit length
		if len(description) > 200 {
			description = description[:197] + "..."
		}
		return strings.TrimSpace(description)
	}

	return "No description available"
}

// Extract content from HTML document
func extractContent(doc *goquery.Document) string {
	var content strings.Builder

	// Try to find main content area
	mainContent := doc.Find("article, main, #content, .content, .post-content").First()

	// If no specific content area found, use body
	if mainContent.Length() == 0 {
		mainContent = doc.Find("body")
	}

	// Extract all paragraphs
	mainContent.Find("p, h1, h2, h3, h4, h5, h6, ul, ol, blockquote").Each(func(i int, s *goquery.Selection) {
		// Get tag name
		tagName := goquery.NodeName(s)
		text := strings.TrimSpace(s.Text())

		if text == "" {
			return
		}

		// Format according to tag type
		switch tagName {
		case "h1":
			content.WriteString("# " + text + "\n\n")
		case "h2":
			content.WriteString("## " + text + "\n\n")
		case "h3":
			content.WriteString("### " + text + "\n\n")
		case "h4":
			content.WriteString("#### " + text + "\n\n")
		case "h5":
			content.WriteString("##### " + text + "\n\n")
		case "h6":
			content.WriteString("###### " + text + "\n\n")
		case "p":
			content.WriteString(text + "\n\n")
		case "blockquote":
			content.WriteString("> " + text + "\n\n")
		case "ul", "ol":
			s.Find("li").Each(func(j int, li *goquery.Selection) {
				liText := strings.TrimSpace(li.Text())
				if liText != "" {
					if tagName == "ul" {
						content.WriteString("- " + liText + "\n")
					} else {
						content.WriteString(fmt.Sprintf("%d. %s\n", j+1, liText))
					}
				}
			})
			content.WriteString("\n")
		}
	})

	// Limit content length
	contentStr := content.String()
	if len(contentStr) > 10000 {
		// Find last paragraph end position
		lastParaEnd := strings.LastIndex(contentStr[:10000], "\n\n")
		if lastParaEnd == -1 {
			lastParaEnd = 10000
		}
		contentStr = contentStr[:lastParaEnd] + "\n\n... (content truncated)"
	}

	return contentStr
}
