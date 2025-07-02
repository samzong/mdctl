package linter

import (
	"fmt"
	"os"
	"strings"

	"github.com/samzong/mdctl/internal/markdownfmt"
)

// Config holds the linter configuration
type Config struct {
	AutoFix      bool
	OutputFormat string
	RulesFile    string
	EnableRules  []string
	DisableRules []string
	Verbose      bool
}

// Issue represents a linting issue
type Issue struct {
	Line    int    `json:"line"`
	Column  int    `json:"column,omitempty"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
	Context string `json:"context,omitempty"`
	Fixed   bool   `json:"fixed,omitempty"`
}

// Result holds the linting results for a file
type Result struct {
	Filename   string   `json:"filename"`
	Issues     []*Issue `json:"issues"`
	FixedCount int      `json:"fixed_count"`
}

// Linter performs markdown linting
type Linter struct {
	config    *Config
	rules     *RuleSet
	formatter *markdownfmt.Formatter
}

// New creates a new linter instance
func New(config *Config) *Linter {
	rules := NewRuleSet()
	
	// Apply rule configuration
	if len(config.EnableRules) > 0 {
		rules.EnableOnly(config.EnableRules)
	}
	
	if len(config.DisableRules) > 0 {
		rules.Disable(config.DisableRules)
	}

	// Load custom rules from file if specified
	if config.RulesFile != "" {
		if err := rules.LoadFromFile(config.RulesFile); err != nil && config.Verbose {
			fmt.Printf("Warning: Could not load rules file %s: %v\n", config.RulesFile, err)
		}
	}

	return &Linter{
		config:    config,
		rules:     rules,
		formatter: markdownfmt.New(true), // Enable formatter for auto-fix
	}
}

// LintFile lints a single markdown file
func (l *Linter) LintFile(filename string) (*Result, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	return l.LintContent(filename, string(content))
}

// LintContent lints markdown content
func (l *Linter) LintContent(filename, content string) (*Result, error) {
	result := &Result{
		Filename: filename,
		Issues:   []*Issue{},
	}

	lines := strings.Split(content, "\n")
	
	// Apply all enabled rules
	for _, rule := range l.rules.GetEnabledRules() {
		issues := rule.Check(lines)
		result.Issues = append(result.Issues, issues...)
	}

	// Apply auto-fix if requested
	if l.config.AutoFix && len(result.Issues) > 0 {
		fixedContent, fixedCount := l.applyFixes(content, result.Issues)
		result.FixedCount = fixedCount
		
		// Write fixed content back to file
		if fixedCount > 0 {
			if err := os.WriteFile(filename, []byte(fixedContent), 0644); err != nil {
				return nil, fmt.Errorf("failed to write fixed content: %v", err)
			}
			
			// Mark issues as fixed
			for _, issue := range result.Issues {
				if issue.Rule != "MD013" { // Don't mark line length issues as fixed automatically
					issue.Fixed = true
				}
			}
		}
	}

	return result, nil
}

// applyFixes applies automatic fixes to the content
func (l *Linter) applyFixes(content string, issues []*Issue) (string, int) {
	// Use the existing formatter for basic fixes
	fixedContent := l.formatter.Format(content)
	
	fixedCount := 0
	if fixedContent != content {
		fixedCount = l.countFixableIssues(issues)
	}

	return fixedContent, fixedCount
}

// countFixableIssues counts how many issues can be automatically fixed
func (l *Linter) countFixableIssues(issues []*Issue) int {
	fixableRules := map[string]bool{
		"MD009": true, // Trailing spaces
		"MD010": true, // Hard tabs
		"MD012": true, // Multiple consecutive blank lines
		"MD018": true, // No space after hash on atx style heading
		"MD019": true, // Multiple spaces after hash on atx style heading
		"MD023": true, // Headings must start at the beginning of the line
		"MD047": true, // Files should end with a single newline character
	}

	count := 0
	for _, issue := range issues {
		if fixableRules[issue.Rule] {
			count++
		}
	}
	return count
}