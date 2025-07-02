package linter

import (
	"testing"
)

func TestMD047_FileEndingCheck(t *testing.T) {
	tests := []struct {
		name        string
		lines       []string
		expectIssue bool
		description string
	}{
		{
			name:        "file ends with single newline",
			lines:       []string{"# Title", "Content", ""},
			expectIssue: false,
			description: "should not trigger issue when file ends with single newline",
		},
		{
			name:        "file does not end with newline",
			lines:       []string{"# Title", "Content"},
			expectIssue: true,
			description: "should trigger issue when file doesn't end with newline",
		},
		{
			name:        "file ends with multiple newlines",
			lines:       []string{"# Title", "Content", "", ""},
			expectIssue: true,
			description: "should trigger issue when file ends with multiple newlines",
		},
		{
			name:        "empty file",
			lines:       []string{},
			expectIssue: false,
			description: "should not trigger issue for empty file",
		},
	}

	rule := &MD047{BaseRule: BaseRule{id: "MD047", description: "Files should end with a single newline character", enabled: true}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := rule.Check(tt.lines)
			hasIssue := len(issues) > 0

			if hasIssue != tt.expectIssue {
				t.Errorf("%s: expected issue=%t, got issue=%t", tt.description, tt.expectIssue, hasIssue)
				if hasIssue {
					for _, issue := range issues {
						t.Logf("Issue: %s", issue.Message)
					}
				}
			}
		})
	}
}

func TestMD032_ListBlankLines(t *testing.T) {
	tests := []struct {
		name        string
		lines       []string
		expectCount int
		description string
	}{
		{
			name: "list with proper blank lines",
			lines: []string{
				"# Title",
				"",
				"- Item 1",
				"- Item 2",
				"",
				"Content after list",
			},
			expectCount: 0,
			description: "should not trigger issue when list has proper blank lines",
		},
		{
			name: "list without blank line before",
			lines: []string{
				"# Title",
				"Some text",
				"- Item 1",
				"",
				"Content after list",
			},
			expectCount: 1,
			description: "should trigger issue when list doesn't have blank line before",
		},
		{
			name: "list without blank line after",
			lines: []string{
				"# Title",
				"",
				"- Item 1",
				"Content after list",
			},
			expectCount: 1,
			description: "should trigger issue when list doesn't have blank line after",
		},
		{
			name: "list without blank lines before and after",
			lines: []string{
				"# Title",
				"Some text",
				"- Item 1",
				"Content after list",
			},
			expectCount: 2,
			description: "should trigger 2 issues when list doesn't have blank lines before and after",
		},
	}

	rule := &MD032{BaseRule: BaseRule{id: "MD032", description: "Lists should be surrounded by blank lines", enabled: true}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := rule.Check(tt.lines)

			if len(issues) != tt.expectCount {
				t.Errorf("%s: expected %d issues, got %d issues", tt.description, tt.expectCount, len(issues))
				for i, issue := range issues {
					t.Logf("Issue %d: Line %d - %s", i+1, issue.Line, issue.Message)
				}
			}
		})
	}
}

func TestRegexPrecompilation(t *testing.T) {
	tests := []struct {
		name string
		rule Rule
	}{
		{"MD018", &MD018{BaseRule: BaseRule{id: "MD018", enabled: true}}},
		{"MD019", &MD019{BaseRule: BaseRule{id: "MD019", enabled: true}}},
		{"MD023", &MD023{BaseRule: BaseRule{id: "MD023", enabled: true}}},
		{"MD032", &MD032{BaseRule: BaseRule{id: "MD032", enabled: true}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call Check method to trigger regex compilation
			_ = tt.rule.Check([]string{"# Test", "Content"})

			// Check that pattern was compiled for rules that have patterns
			switch rule := tt.rule.(type) {
			case *MD018:
				if rule.pattern == nil {
					t.Error("MD018 pattern was not compiled")
				}
			case *MD019:
				if rule.pattern == nil {
					t.Error("MD019 pattern was not compiled")
				}
			case *MD023:
				if rule.pattern == nil {
					t.Error("MD023 pattern was not compiled")
				}
			case *MD032:
				if rule.pattern == nil {
					t.Error("MD032 pattern was not compiled")
				}
			}
		})
	}
}

func TestMD018_NoSpaceAfterHash(t *testing.T) {
	rule := &MD018{BaseRule: BaseRule{id: "MD018", enabled: true}}

	tests := []struct {
		line        string
		expectIssue bool
	}{
		{"# Proper heading", false},
		{"#Bad heading", true},
		{"## Another proper heading", false},
		{"##Bad heading", true},
		{"### Yet another proper heading", false},
		{"###Bad heading", true},
		{"Not a heading", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			issues := rule.Check([]string{tt.line})
			hasIssue := len(issues) > 0

			if hasIssue != tt.expectIssue {
				t.Errorf("Line %q: expected issue=%t, got issue=%t", tt.line, tt.expectIssue, hasIssue)
			}
		})
	}
}

func TestMD019_MultipleSpacesAfterHash(t *testing.T) {
	rule := &MD019{BaseRule: BaseRule{id: "MD019", enabled: true}}

	tests := []struct {
		line        string
		expectIssue bool
	}{
		{"# Proper heading", false},
		{"#  Bad heading", true},
		{"## Another proper heading", false},
		{"##  Bad heading", true},
		{"###   Very bad heading", true},
		{"Not a heading", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			issues := rule.Check([]string{tt.line})
			hasIssue := len(issues) > 0

			if hasIssue != tt.expectIssue {
				t.Errorf("Line %q: expected issue=%t, got issue=%t", tt.line, tt.expectIssue, hasIssue)
			}
		})
	}
}

func TestMD023_HeadingAtStartOfLine(t *testing.T) {
	rule := &MD023{BaseRule: BaseRule{id: "MD023", enabled: true}}

	tests := []struct {
		line        string
		expectIssue bool
	}{
		{"# Proper heading", false},
		{" # Bad heading", true},
		{"  ## Very bad heading", true},
		{"Not a heading", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			issues := rule.Check([]string{tt.line})
			hasIssue := len(issues) > 0

			if hasIssue != tt.expectIssue {
				t.Errorf("Line %q: expected issue=%t, got issue=%t", tt.line, tt.expectIssue, hasIssue)
			}
		})
	}
}
