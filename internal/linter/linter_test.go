package linter

import (
	"os"
	"testing"
)

func TestLinter_LintContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectRules []string // Expected rule IDs that should trigger
		expectCount int      // Expected number of issues
	}{
		{
			name:        "valid markdown",
			content:     "# Title\n\nThis is valid markdown.\n",
			expectRules: []string{},
			expectCount: 0,
		},
		{
			name:        "trailing spaces",
			content:     "# Title  \n\nContent with trailing spaces.  \n",
			expectRules: []string{"MD009"},
			expectCount: 2,
		},
		{
			name:        "hard tabs",
			content:     "# Title\n\n\tContent with hard tab.\n",
			expectRules: []string{"MD010"},
			expectCount: 1,
		},
		{
			name:        "multiple blank lines",
			content:     "# Title\n\n\n\nContent after multiple blank lines.\n",
			expectRules: []string{"MD012"},
			expectCount: 2, // MD012 triggers for each set of consecutive blank lines
		},
		{
			name:        "no space after hash",
			content:     "#Title\n\nContent.\n",
			expectRules: []string{"MD018"},
			expectCount: 1,
		},
		{
			name:        "multiple spaces after hash",
			content:     "#  Title\n\nContent.\n",
			expectRules: []string{"MD019"},
			expectCount: 1,
		},
		{
			name:        "heading not at start of line",
			content:     "Some text\n # Title\n\nContent.\n",
			expectRules: []string{"MD023"},
			expectCount: 1,
		},
		{
			name:        "list without blank line before",
			content:     "# Title\nSome text\n- List item\n\nContent.\n",
			expectRules: []string{"MD032"},
			expectCount: 1,
		},
		{
			name:        "list without blank line after",
			content:     "# Title\n\n- List item\nSome text\n",
			expectRules: []string{"MD032"},
			expectCount: 1,
		},
		{
			name:        "file not ending with newline",
			content:     "# Title\n\nContent without final newline",
			expectRules: []string{"MD047"},
			expectCount: 1,
		},
		{
			name:        "file ending with multiple newlines",
			content:     "# Title\n\nContent.\n\n",
			expectRules: []string{"MD047", "MD012"},
			expectCount: 2, // Both MD047 and MD012 trigger
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linter := New(&Config{})
			result, err := linter.LintContent("test.md", tt.content)

			if err != nil {
				t.Fatalf("LintContent failed: %v", err)
			}

			if len(result.Issues) != tt.expectCount {
				t.Errorf("Expected %d issues, got %d", tt.expectCount, len(result.Issues))
				for _, issue := range result.Issues {
					t.Logf("Issue: %s - %s", issue.Rule, issue.Message)
				}
			}

			// Check that expected rules are triggered
			foundRules := make(map[string]bool)
			for _, issue := range result.Issues {
				foundRules[issue.Rule] = true
			}

			for _, expectedRule := range tt.expectRules {
				if !foundRules[expectedRule] {
					t.Errorf("Expected rule %s to be triggered, but it wasn't", expectedRule)
				}
			}
		})
	}
}

func TestLinter_AutoFix(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectFixed    bool
		expectFixCount int
		expectRules    []string
	}{
		{
			name:           "fix trailing spaces",
			content:        "# Title  \n\nContent with trailing spaces.  \n",
			expectFixed:    true,
			expectFixCount: 2,
			expectRules:    []string{"MD009"},
		},
		{
			name:           "fix hard tabs",
			content:        "# Title\n\n\tContent with hard tab.\n",
			expectFixed:    true,
			expectFixCount: 1,
			expectRules:    []string{"MD010"},
		},
		{
			name:           "fix multiple blank lines",
			content:        "# Title\n\n\n\nContent after multiple blank lines.\n",
			expectFixed:    true,
			expectFixCount: 2, // MD012 triggers multiple times
			expectRules:    []string{"MD012"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tmpFile, err := os.CreateTemp("", "test_*.md")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())
			defer os.Remove(tmpFile.Name() + ".orig") // Remove backup file

			// Write content to temp file
			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			// Run linter with auto-fix
			linter := New(&Config{AutoFix: true})
			result, err := linter.LintFile(tmpFile.Name())

			if err != nil {
				t.Fatalf("LintFile failed: %v", err)
			}

			if tt.expectFixed && result.FixedCount != tt.expectFixCount {
				t.Errorf("Expected %d fixes, got %d", tt.expectFixCount, result.FixedCount)
			}

			// Check that backup file was created
			if tt.expectFixed {
				if _, err := os.Stat(tmpFile.Name() + ".orig"); os.IsNotExist(err) {
					t.Error("Expected backup file to be created, but it wasn't")
				}
			}
		})
	}
}

func TestLinter_BackupCreation(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test_*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer os.Remove(tmpFile.Name() + ".orig")

	originalContent := "# Title  \n\nContent with trailing spaces.  \n"
	if _, err := tmpFile.WriteString(originalContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Run linter with auto-fix
	linter := New(&Config{AutoFix: true})
	_, err = linter.LintFile(tmpFile.Name())

	if err != nil {
		t.Fatalf("LintFile failed: %v", err)
	}

	// Check that backup file exists and contains original content
	backupContent, err := os.ReadFile(tmpFile.Name() + ".orig")
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if string(backupContent) != originalContent {
		t.Errorf("Backup content doesn't match original.\nExpected: %q\nGot: %q", originalContent, string(backupContent))
	}
}
