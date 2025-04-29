package llmstxt

import (
	"strings"
	"unicode"
)

// Format to Markdown content
func (g *Generator) formatContent(sections map[string][]PageInfo) string {
	var buf strings.Builder

	// Get sorted section list
	sectionNames := g.getSortedSections(sections)

	// Find root page info
	var rootPage PageInfo
	if rootPages, ok := sections["ROOT"]; ok && len(rootPages) > 0 {
		rootPage = rootPages[0]
	}

	// Add document title
	buf.WriteString("# ")
	buf.WriteString(rootPage.Title)
	buf.WriteString("\n\n")

	// Add document description
	buf.WriteString("> ")
	buf.WriteString(rootPage.Description)
	buf.WriteString("\n\n")

	// Handle each section
	for _, section := range sectionNames {
		// Skip ROOT section, because it's already used for title and description
		if section == "ROOT" {
			continue
		}

		// Add section title
		buf.WriteString("## ")
		buf.WriteString(capitalizeString(section))
		buf.WriteString("\n\n")

		// Add page info for each page in section
		for _, page := range sections[section] {
			buf.WriteString("- [")
			buf.WriteString(page.Title)
			buf.WriteString("](")
			buf.WriteString(page.URL)
			buf.WriteString("): ")
			buf.WriteString(page.Description)
			buf.WriteString("\n")

			// Add page content in full mode
			if g.config.FullMode && page.Content != "" {
				buf.WriteString("\n")
				buf.WriteString(page.Content)
				buf.WriteString("\n")
			}

			buf.WriteString("\n")
		}
	}

	return buf.String()
}

// Capitalize first letter, lowercase the rest
func capitalizeString(str string) string {
	if str == "" {
		return ""
	}

	runes := []rune(str)
	return string(unicode.ToUpper(runes[0])) + strings.ToLower(string(runes[1:]))
}
