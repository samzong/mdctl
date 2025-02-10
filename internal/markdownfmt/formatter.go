package markdownfmt

import (
	"fmt"
	"regexp"
	"strings"
)

// Formatter for formatting markdown content
type Formatter struct {
	// Whether formatting is enabled
	enabled bool
}

// New creates a new formatter
func New(enabled bool) *Formatter {
	return &Formatter{
		enabled: enabled,
	}
}

// Format formats markdown content
func (f *Formatter) Format(content string) string {
	if !f.enabled {
		return content
	}

	// 1. Split content into lines
	lines := strings.Split(content, "\n")

	// 2. Process each line
	var formatted []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Process headings: ensure there are blank lines before and after
		if isHeading(line) {
			// If not the first line and previous line is not blank, add a blank line
			if i > 0 && len(strings.TrimSpace(lines[i-1])) > 0 {
				formatted = append(formatted, "")
			}
			// Normalize heading format (one space after #)
			line = formatHeading(line)
			formatted = append(formatted, line)
			// If not the last line, add a blank line
			if i < len(lines)-1 {
				formatted = append(formatted, "")
			}
			continue
		}

		// Process spaces in links
		line = formatMarkdownLinks(line)

		// Process content in parentheses
		line = formatParentheses(line)

		// Process spaces between Chinese and English text
		line = formatChineseEnglishSpace(line)

		formatted = append(formatted, line)
	}

	// 3. Handle consecutive blank lines
	formatted = removeConsecutiveBlankLines(formatted)

	// 4. Join lines
	result := strings.Join(formatted, "\n")

	return result
}

// isHeading checks if the line is a heading
func isHeading(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#")
}

// formatHeading formats the heading line
func formatHeading(line string) string {
	// Remove leading spaces
	line = strings.TrimSpace(line)
	// Ensure only one space between # and text
	re := regexp.MustCompile(`^(#+)\s*`)
	return re.ReplaceAllString(line, "$1 ")
}

// formatParentheses processes the format within parentheses
func formatParentheses(line string) string {
	// First handle http/https links by temporarily replacing them
	linkPattern := regexp.MustCompile(`\([^)]*https?://[^)]+\)`)
	links := linkPattern.FindAllString(line, -1)
	for i, link := range links {
		line = strings.Replace(line, link, fmt.Sprintf("__LINK_PLACEHOLDER_%d__", i), 1)
	}

	// Process regular parentheses content
	re := regexp.MustCompile(`\(([^)]+)\)`)
	line = re.ReplaceAllStringFunc(line, func(match string) string {
		// Extract content within parentheses
		content := match[1 : len(match)-1]
		// Clean leading and trailing spaces
		content = strings.TrimSpace(content)
		// Replace consecutive spaces with a single space
		content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
		return fmt.Sprintf("(%s)", content)
	})

	// Restore links
	for i, link := range links {
		line = strings.Replace(line, fmt.Sprintf("__LINK_PLACEHOLDER_%d__", i), link, 1)
	}

	return line
}

// formatMarkdownLinks processes spaces in markdown links
func formatMarkdownLinks(line string) string {
	// Match markdown link format [text](url), including possible spaces
	linkPattern := regexp.MustCompile(`\[(.*?)\]\(\s*(.*?)\s*\)`)

	// Process spaces in link text and URL
	line = linkPattern.ReplaceAllStringFunc(line, func(match string) string {
		// Extract link text and URL
		parts := linkPattern.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}

		text := parts[1]
		url := parts[2]

		// Clean spaces in URL
		url = strings.TrimSpace(url)
		// Remove all spaces and invisible characters in URL
		url = regexp.MustCompile(`[\s\p{Zs}\p{C}]+`).ReplaceAllString(url, "")

		// Keep spaces in link text, but clean leading/trailing spaces and consecutive spaces
		text = strings.TrimSpace(text)
		text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

		// Reassemble link
		return fmt.Sprintf("[%s](%s)", text, url)
	})

	// Process spaces in heading links
	headingLinkPattern := regexp.MustCompile(`\]\(#(.*?)\)`)
	line = headingLinkPattern.ReplaceAllStringFunc(line, func(match string) string {
		parts := headingLinkPattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}

		anchor := parts[1]
		// Remove all spaces
		anchor = regexp.MustCompile(`\s+`).ReplaceAllString(anchor, "")
		return fmt.Sprintf("](#%s)", anchor)
	})

	return line
}

// formatChineseEnglishSpace adds spaces between Chinese and English text
func formatChineseEnglishSpace(line string) string {
	// Match boundaries between Chinese and English/numbers
	re := regexp.MustCompile(`([\p{Han}])([A-Za-z0-9])`)
	line = re.ReplaceAllString(line, "$1 $2")

	re = regexp.MustCompile(`([A-Za-z0-9])([\p{Han}])`)
	line = re.ReplaceAllString(line, "$1 $2")

	return line
}

// removeConsecutiveBlankLines removes consecutive blank lines
func removeConsecutiveBlankLines(lines []string) []string {
	var result []string
	isPrevLineBlank := false

	for _, line := range lines {
		isCurrentLineBlank := len(strings.TrimSpace(line)) == 0

		if !isCurrentLineBlank || !isPrevLineBlank {
			result = append(result, line)
		}

		isPrevLineBlank = isCurrentLineBlank
	}

	return result
}
