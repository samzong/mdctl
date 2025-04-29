package exporter

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

var (
	// Match ATX-style headings (those starting with #)
	atxHeadingRegex = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	// Match Setext-style headings (underline style)
	setextHeading1Regex = regexp.MustCompile(`^=+\s*$`)
	setextHeading2Regex = regexp.MustCompile(`^-+\s*$`)
)

// ShiftHeadings Adjust heading levels in Markdown text
func ShiftHeadings(content string, shiftBy int) string {
	if shiftBy == 0 {
		return content
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	var result []string
	var prevLine string
	var isPrevLineHeading bool

	for scanner.Scan() {
		line := scanner.Text()

		// Handle ATX-style headings
		if matches := atxHeadingRegex.FindStringSubmatch(line); matches != nil {
			level := len(matches[1]) + shiftBy
			heading := matches[2]

			if level <= 6 {
				// Still valid heading level
				result = append(result, fmt.Sprintf("%s %s", strings.Repeat("#", level), heading))
			} else {
				// Exceeded max heading level, convert to bold text
				result = append(result, fmt.Sprintf("**%s**", heading))
			}
			isPrevLineHeading = false
		} else if setextHeading1Regex.MatchString(line) && prevLine != "" {
			// Handle Setext-style level 1 headings
			level := 1 + shiftBy
			if level <= 6 {
				result[len(result)-1] = fmt.Sprintf("%s %s", strings.Repeat("#", level), prevLine)
			} else {
				result[len(result)-1] = fmt.Sprintf("**%s**", prevLine)
			}
			isPrevLineHeading = true
		} else if setextHeading2Regex.MatchString(line) && prevLine != "" {
			// Handle Setext-style level 2 headings
			level := 2 + shiftBy
			if level <= 6 {
				result[len(result)-1] = fmt.Sprintf("%s %s", strings.Repeat("#", level), prevLine)
			} else {
				result[len(result)-1] = fmt.Sprintf("**%s**", prevLine)
			}
			isPrevLineHeading = true
		} else {
			// Ordinary line
			result = append(result, line)
			isPrevLineHeading = false
		}

		if !isPrevLineHeading {
			prevLine = line
		}
	}

	return strings.Join(result, "\n")
}

// AddTitleFromFilename Add heading from filename
func AddTitleFromFilename(content, filename string, level int) string {
	// Extract heading from filename (remove extension)
	title := strings.TrimSuffix(filename, ".md")
	title = strings.TrimSuffix(title, ".markdown")

	// Replace underscores and hyphens with spaces, making the heading more readable
	title = strings.ReplaceAll(title, "_", " ")
	title = strings.ReplaceAll(title, "-", " ")

	// Capitalize the first letter of each word
	title = strings.Title(title)

	// Create heading line
	var titleLine string
	if level <= 6 {
		titleLine = fmt.Sprintf("%s %s\n\n", strings.Repeat("#", level), title)
	} else {
		titleLine = fmt.Sprintf("**%s**\n\n", title)
	}

	return titleLine + content
}
