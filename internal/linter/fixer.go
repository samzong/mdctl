package linter

import (
	"regexp"
	"strings"
)

// Fixer provides auto-fix functionality for markdown issues
type Fixer struct {
	rules map[string]func([]string) ([]string, int)
}

// NewFixer creates a new fixer instance
func NewFixer() *Fixer {
	f := &Fixer{
		rules: make(map[string]func([]string) ([]string, int)),
	}
	
	// Register fix functions for each rule
	f.rules["MD009"] = f.fixTrailingSpaces
	f.rules["MD010"] = f.fixHardTabs
	f.rules["MD012"] = f.fixMultipleBlankLines
	f.rules["MD018"] = f.fixNoSpaceAfterHash
	f.rules["MD019"] = f.fixMultipleSpacesAfterHash
	f.rules["MD023"] = f.fixHeadingIndentation
	f.rules["MD032"] = f.fixListSpacing
	f.rules["MD047"] = f.fixFileEndNewline
	
	return f
}

// ApplyFixes applies fixes for the given issues
func (f *Fixer) ApplyFixes(content string, issues []*Issue) (string, int) {
	lines := strings.Split(content, "\n")
	totalFixed := 0
	
	// Group issues by rule for efficient processing
	ruleIssues := make(map[string][]*Issue)
	for _, issue := range issues {
		ruleIssues[issue.Rule] = append(ruleIssues[issue.Rule], issue)
	}
	
	// Apply fixes for each rule
	for rule, ruleSpecificIssues := range ruleIssues {
		if fixFunc, exists := f.rules[rule]; exists {
			var fixed int
			lines, fixed = fixFunc(lines)
			totalFixed += fixed
			
			// Mark issues as fixed
			for _, issue := range ruleSpecificIssues {
				issue.Fixed = true
			}
		}
	}
	
	return strings.Join(lines, "\n"), totalFixed
}

// fixTrailingSpaces removes trailing spaces from lines
func (f *Fixer) fixTrailingSpaces(lines []string) ([]string, int) {
	fixed := 0
	for i, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if trimmed != line {
			lines[i] = trimmed
			fixed++
		}
	}
	return lines, fixed
}

// fixHardTabs replaces hard tabs with spaces
func (f *Fixer) fixHardTabs(lines []string) ([]string, int) {
	fixed := 0
	for i, line := range lines {
		if strings.Contains(line, "\t") {
			lines[i] = strings.ReplaceAll(line, "\t", "    ")
			fixed++
		}
	}
	return lines, fixed
}

// fixMultipleBlankLines removes consecutive blank lines
func (f *Fixer) fixMultipleBlankLines(lines []string) ([]string, int) {
	var result []string
	fixed := 0
	prevBlank := false
	
	for _, line := range lines {
		isBlank := strings.TrimSpace(line) == ""
		
		if isBlank && prevBlank {
			fixed++ // Count removed blank lines
			continue
		}
		
		result = append(result, line)
		prevBlank = isBlank
	}
	
	return result, fixed
}

// fixNoSpaceAfterHash adds space after hash in headings
func (f *Fixer) fixNoSpaceAfterHash(lines []string) ([]string, int) {
	fixed := 0
	re := regexp.MustCompile(`^(#+)([^# ])`)
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if re.MatchString(trimmed) {
			lines[i] = re.ReplaceAllString(trimmed, "$1 $2")
			fixed++
		}
	}
	
	return lines, fixed
}

// fixMultipleSpacesAfterHash removes extra spaces after hash in headings
func (f *Fixer) fixMultipleSpacesAfterHash(lines []string) ([]string, int) {
	fixed := 0
	re := regexp.MustCompile(`^(#+)\s{2,}`)
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if re.MatchString(trimmed) {
			lines[i] = re.ReplaceAllString(trimmed, "$1 ")
			fixed++
		}
	}
	
	return lines, fixed
}

// fixHeadingIndentation removes leading spaces from headings
func (f *Fixer) fixHeadingIndentation(lines []string) ([]string, int) {
	fixed := 0
	re := regexp.MustCompile(`^ +(#.*)`)
	
	for i, line := range lines {
		if re.MatchString(line) {
			lines[i] = re.ReplaceAllString(line, "$1")
			fixed++
		}
	}
	
	return lines, fixed
}

// fixListSpacing adds blank lines around lists
func (f *Fixer) fixListSpacing(lines []string) ([]string, int) {
	fixed := 0
	var result []string
	listRe := regexp.MustCompile(`^(\s*[*+-] )`)
	
	for i, line := range lines {
		if listRe.MatchString(line) {
			// Check if previous line needs a blank line
			if i > 0 && strings.TrimSpace(lines[i-1]) != "" && len(result) > 0 {
				result = append(result, "")
				fixed++
			}
		}
		result = append(result, line)
	}
	
	return result, fixed
}

// fixFileEndNewline ensures file ends with single newline
func (f *Fixer) fixFileEndNewline(lines []string) ([]string, int) {
	if len(lines) == 0 {
		return lines, 0
	}
	
	// Remove trailing empty lines
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	
	// Add single empty line at the end
	lines = append(lines, "")
	
	return lines, 1
}