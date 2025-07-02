package linter

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
)

// Rule represents a markdown linting rule
type Rule interface {
	ID() string
	Description() string
	Check(lines []string) []*Issue
	Enabled() bool
	SetEnabled(enabled bool)
}

// BaseRule provides common functionality for rules
type BaseRule struct {
	id          string
	description string
	enabled     bool
}

func (r *BaseRule) ID() string          { return r.id }
func (r *BaseRule) Description() string { return r.description }
func (r *BaseRule) Enabled() bool       { return r.enabled }
func (r *BaseRule) SetEnabled(enabled bool) { r.enabled = enabled }

// RuleSet manages a collection of linting rules
type RuleSet struct {
	rules map[string]Rule
}

// NewRuleSet creates a new rule set with default rules
func NewRuleSet() *RuleSet {
	rs := &RuleSet{
		rules: make(map[string]Rule),
	}
	
	// Add default rules
	rs.addRule(&MD001{BaseRule: BaseRule{id: "MD001", description: "Heading levels should only increment by one level at a time", enabled: true}})
	rs.addRule(&MD003{BaseRule: BaseRule{id: "MD003", description: "Heading style should be consistent", enabled: true}})
	rs.addRule(&MD009{BaseRule: BaseRule{id: "MD009", description: "Trailing spaces", enabled: true}})
	rs.addRule(&MD010{BaseRule: BaseRule{id: "MD010", description: "Hard tabs", enabled: true}})
	rs.addRule(&MD012{BaseRule: BaseRule{id: "MD012", description: "Multiple consecutive blank lines", enabled: true}})
	rs.addRule(&MD013{BaseRule: BaseRule{id: "MD013", description: "Line length", enabled: true}})
	rs.addRule(&MD018{BaseRule: BaseRule{id: "MD018", description: "No space after hash on atx style heading", enabled: true}})
	rs.addRule(&MD019{BaseRule: BaseRule{id: "MD019", description: "Multiple spaces after hash on atx style heading", enabled: true}})
	rs.addRule(&MD023{BaseRule: BaseRule{id: "MD023", description: "Headings must start at the beginning of the line", enabled: true}})
	rs.addRule(&MD032{BaseRule: BaseRule{id: "MD032", description: "Lists should be surrounded by blank lines", enabled: true}})
	rs.addRule(&MD047{BaseRule: BaseRule{id: "MD047", description: "Files should end with a single newline character", enabled: true}})
	
	return rs
}

func (rs *RuleSet) addRule(rule Rule) {
	rs.rules[rule.ID()] = rule
}

// GetEnabledRules returns all enabled rules
func (rs *RuleSet) GetEnabledRules() []Rule {
	var enabled []Rule
	for _, rule := range rs.rules {
		if rule.Enabled() {
			enabled = append(enabled, rule)
		}
	}
	return enabled
}

// EnableOnly enables only the specified rules
func (rs *RuleSet) EnableOnly(ruleIDs []string) {
	// Disable all rules first
	for _, rule := range rs.rules {
		rule.SetEnabled(false)
	}
	
	// Enable specified rules
	for _, id := range ruleIDs {
		if rule, exists := rs.rules[id]; exists {
			rule.SetEnabled(true)
		}
	}
}

// Disable disables the specified rules
func (rs *RuleSet) Disable(ruleIDs []string) {
	for _, id := range ruleIDs {
		if rule, exists := rs.rules[id]; exists {
			rule.SetEnabled(false)
		}
	}
}

// LoadFromFile loads rule configuration from a JSON file
func (rs *RuleSet) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}
	
	// Apply configuration
	for ruleID, setting := range config {
		if rule, exists := rs.rules[ruleID]; exists {
			switch v := setting.(type) {
			case bool:
				rule.SetEnabled(v)
			case map[string]interface{}:
				// For now, just check if it's explicitly disabled
				if enabled, ok := v["enabled"].(bool); ok {
					rule.SetEnabled(enabled)
				}
			}
		}
	}
	
	return nil
}

// Rule implementations

// MD001: Heading levels should only increment by one level at a time
type MD001 struct {
	BaseRule
}

func (r *MD001) Check(lines []string) []*Issue {
	var issues []*Issue
	lastLevel := 0
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			level := 0
			for _, char := range line {
				if char == '#' {
					level++
				} else {
					break
				}
			}
			
			if lastLevel > 0 && level > lastLevel+1 {
				issues = append(issues, &Issue{
					Line:    i + 1,
					Rule:    r.ID(),
					Message: "Heading levels should only increment by one level at a time",
					Context: line,
				})
			}
			lastLevel = level
		}
	}
	
	return issues
}

// MD003: Heading style should be consistent
type MD003 struct {
	BaseRule
}

func (r *MD003) Check(lines []string) []*Issue {
	var issues []*Issue
	var firstStyle string
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			currentStyle := "atx"
			if firstStyle == "" {
				firstStyle = currentStyle
			} else if firstStyle != currentStyle {
				issues = append(issues, &Issue{
					Line:    i + 1,
					Rule:    r.ID(),
					Message: "Heading style should be consistent",
					Context: line,
				})
			}
		}
	}
	
	return issues
}

// MD009: Trailing spaces
type MD009 struct {
	BaseRule
}

func (r *MD009) Check(lines []string) []*Issue {
	var issues []*Issue
	
	for i, line := range lines {
		if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
			issues = append(issues, &Issue{
				Line:    i + 1,
				Rule:    r.ID(),
				Message: "Trailing spaces",
				Context: line,
			})
		}
	}
	
	return issues
}

// MD010: Hard tabs
type MD010 struct {
	BaseRule
}

func (r *MD010) Check(lines []string) []*Issue {
	var issues []*Issue
	
	for i, line := range lines {
		if strings.Contains(line, "\t") {
			issues = append(issues, &Issue{
				Line:    i + 1,
				Rule:    r.ID(),
				Message: "Hard tabs",
				Context: line,
			})
		}
	}
	
	return issues
}

// MD012: Multiple consecutive blank lines
type MD012 struct {
	BaseRule
}

func (r *MD012) Check(lines []string) []*Issue {
	var issues []*Issue
	consecutiveBlank := 0
	
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			consecutiveBlank++
			if consecutiveBlank > 1 {
				issues = append(issues, &Issue{
					Line:    i + 1,
					Rule:    r.ID(),
					Message: "Multiple consecutive blank lines",
				})
			}
		} else {
			consecutiveBlank = 0
		}
	}
	
	return issues
}

// MD013: Line length
type MD013 struct {
	BaseRule
}

func (r *MD013) Check(lines []string) []*Issue {
	var issues []*Issue
	maxLength := 80 // Default line length limit
	
	for i, line := range lines {
		if len(line) > maxLength {
			issues = append(issues, &Issue{
				Line:    i + 1,
				Rule:    r.ID(),
				Message: "Line too long",
				Context: line,
			})
		}
	}
	
	return issues
}

// MD018: No space after hash on atx style heading
type MD018 struct {
	BaseRule
}

func (r *MD018) Check(lines []string) []*Issue {
	var issues []*Issue
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if match, _ := regexp.MatchString(`^#+[^# ]`, line); match {
			issues = append(issues, &Issue{
				Line:    i + 1,
				Rule:    r.ID(),
				Message: "No space after hash on atx style heading",
				Context: line,
			})
		}
	}
	
	return issues
}

// MD019: Multiple spaces after hash on atx style heading
type MD019 struct {
	BaseRule
}

func (r *MD019) Check(lines []string) []*Issue {
	var issues []*Issue
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if match, _ := regexp.MatchString(`^#+  +`, line); match {
			issues = append(issues, &Issue{
				Line:    i + 1,
				Rule:    r.ID(),
				Message: "Multiple spaces after hash on atx style heading",
				Context: line,
			})
		}
	}
	
	return issues
}

// MD023: Headings must start at the beginning of the line
type MD023 struct {
	BaseRule
}

func (r *MD023) Check(lines []string) []*Issue {
	var issues []*Issue
	
	for i, line := range lines {
		if match, _ := regexp.MatchString(`^ +#`, line); match {
			issues = append(issues, &Issue{
				Line:    i + 1,
				Rule:    r.ID(),
				Message: "Headings must start at the beginning of the line",
				Context: line,
			})
		}
	}
	
	return issues
}

// MD032: Lists should be surrounded by blank lines
type MD032 struct {
	BaseRule
}

func (r *MD032) Check(lines []string) []*Issue {
	var issues []*Issue
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if match, _ := regexp.MatchString(`^[*+-] `, line); match {
			// Check if previous line is not blank (and not start of file)
			if i > 0 && strings.TrimSpace(lines[i-1]) != "" {
				issues = append(issues, &Issue{
					Line:    i + 1,
					Rule:    r.ID(),
					Message: "Lists should be surrounded by blank lines",
					Context: line,
				})
			}
		}
	}
	
	return issues
}

// MD047: Files should end with a single newline character
type MD047 struct {
	BaseRule
}

func (r *MD047) Check(lines []string) []*Issue {
	var issues []*Issue
	
	if len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		if lastLine != "" {
			issues = append(issues, &Issue{
				Line:    len(lines),
				Rule:    r.ID(),
				Message: "Files should end with a single newline character",
			})
		}
	}
	
	return issues
}