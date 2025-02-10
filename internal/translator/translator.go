package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/samzong/mdctl/internal/config"
	"github.com/samzong/mdctl/internal/markdownfmt"
	"gopkg.in/yaml.v3"
)

// SupportedLanguages defines the mapping of supported languages
var SupportedLanguages = map[string]string{
	"zh": "中文",
	"en": "English",
	"ja": "日本語",
	"ko": "한국어",
	"fr": "Français",
	"de": "Deutsch",
	"es": "Español",
	"it": "Italiano",
	"ru": "Русский",
	"pt": "Português",
	"vi": "Tiếng Việt",
	"th": "ไทย",
	"ar": "العربية",
	"hi": "हिन्दी",
}

// IsLanguageSupported checks if the language is supported
func IsLanguageSupported(lang string) bool {
	_, ok := SupportedLanguages[lang]
	return ok
}

// GetSupportedLanguages returns a list of supported languages
func GetSupportedLanguages() string {
	var langs []string
	for code, name := range SupportedLanguages {
		langs = append(langs, fmt.Sprintf("%s (%s)", code, name))
	}
	sort.Strings(langs)
	return strings.Join(langs, ", ")
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	TopP        float64         `json:"top_p"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Progress is used to track translation progress
type Progress struct {
	Total      int
	Current    int
	SourceFile string
	TargetFile string
}

// ProgressCallback defines the progress callback function type
type ProgressCallback func(progress Progress)

// Translator struct for the translator
type Translator struct {
	config   *config.Config
	format   bool
	progress ProgressCallback
}

// New creates a new translator instance
func New(cfg *config.Config, format bool) *Translator {
	return &Translator{
		config: cfg,
		format: format,
		progress: func(p Progress) {
			if p.Total > 1 {
				fmt.Printf("Translating file [%d/%d]: %s\n", p.Current, p.Total, p.SourceFile)
			}
		},
	}
}

// TranslateContent translates the content
func (t *Translator) TranslateContent(content string, lang string) (string, error) {
	// Remove potential front matter
	content = removeFrontMatter(content)

	prompt := strings.Replace(t.config.TranslatePrompt, "{TARGET_LANG}", lang, 1)

	messages := []OpenAIMessage{
		{Role: "system", Content: prompt},
		{Role: "user", Content: content},
	}

	reqBody := OpenAIRequest{
		Model:       t.config.ModelName,
		Messages:    messages,
		Temperature: t.config.Temperature,
		TopP:        t.config.TopP,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", t.config.OpenAIEndpointURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.config.OpenAIAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v\nResponse body: %s", err, string(body))
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no translation result\nResponse body: %s", string(body))
	}

	// Get translated content
	translatedContent := response.Choices[0].Message.Content

	// Remove potential markdown code block markers
	translatedContent = strings.TrimPrefix(translatedContent, "```markdown\n")
	translatedContent = strings.TrimSuffix(translatedContent, "\n```")
	translatedContent = strings.TrimSpace(translatedContent)

	// If formatting is enabled, format the translated content
	if t.format {
		formatter := markdownfmt.New(true)
		translatedContent = formatter.Format(translatedContent)
	}

	return translatedContent, nil
}

// removeFrontMatter removes front matter from content
func removeFrontMatter(content string) string {
	// If content starts with ---, it may contain front matter
	trimmedContent := strings.TrimSpace(content)
	if strings.HasPrefix(trimmedContent, "---") {
		parts := strings.SplitN(trimmedContent, "---", 3)
		if len(parts) >= 3 {
			return strings.TrimSpace(parts[2])
		}
	}
	return content
}

// ProcessFile handles translation of a single file
func ProcessFile(srcPath, dstPath, targetLang string, cfg *config.Config, format bool, force bool) error {
	t := New(cfg, format)

	// Check if target path is a directory
	dstInfo, err := os.Stat(dstPath)
	if err == nil && dstInfo.IsDir() {
		dstPath = filepath.Join(dstPath, filepath.Base(srcPath))
	}

	// Check if target file already exists
	if _, err := os.Stat(dstPath); err == nil {
		dstContent, err := os.ReadFile(dstPath)
		if err != nil {
			return fmt.Errorf("failed to read target file: %v", err)
		}

		// Check if already translated
		var dstFrontMatter map[string]interface{}
		if strings.HasPrefix(string(dstContent), "---\n") {
			parts := strings.SplitN(string(dstContent)[4:], "\n---\n", 2)
			if len(parts) == 2 {
				if err := yaml.Unmarshal([]byte(parts[0]), &dstFrontMatter); err != nil {
					return fmt.Errorf("failed to parse target file front matter: %v", err)
				}
				if translated, ok := dstFrontMatter["translated"].(bool); ok && translated {
					if !force {
						fmt.Printf("Skipping %s (already translated, use -F to force translate)\n", srcPath)
						return nil
					}
					fmt.Printf("Force translating %s\n", srcPath)
				}
			}
		}
	}

	// Read source file content
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %v", err)
	}

	// Parse front matter
	var frontMatter map[string]interface{}
	contentToTranslate := string(content)

	// Check and parse front matter
	if strings.HasPrefix(contentToTranslate, "---\n") {
		parts := strings.SplitN(contentToTranslate[4:], "\n---\n", 2)
		if len(parts) == 2 {
			if err := yaml.Unmarshal([]byte(parts[0]), &frontMatter); err != nil {
				return fmt.Errorf("failed to parse front matter: %v", err)
			}
			contentToTranslate = parts[1]
		}
	}

	// Translate content
	translatedContent, err := t.TranslateContent(contentToTranslate, targetLang)
	if err != nil {
		return fmt.Errorf("failed to translate content: %v", err)
	}

	// Update front matter
	if frontMatter == nil {
		frontMatter = make(map[string]interface{})
	}
	frontMatter["translated"] = true

	// Generate new file content
	frontMatterBytes, err := yaml.Marshal(frontMatter)
	if err != nil {
		return fmt.Errorf("failed to marshal front matter: %v", err)
	}

	newContent := fmt.Sprintf("---\n%s---\n\n%s", string(frontMatterBytes), translatedContent)

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	// Write translated content to target file
	if err := os.WriteFile(dstPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write target file: %v", err)
	}

	return nil
}

// ProcessDirectory processes all markdown files in the directory
func ProcessDirectory(srcDir, dstDir string, targetLang string, cfg *config.Config, force bool, format bool) error {
	// First calculate the total number of files to process
	var total int
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			total++
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to count files: %v", err)
	}

	fmt.Printf("Found %d markdown files to translate\n", total)

	// Create translator instance
	t := New(cfg, format)
	current := 0

	// Walk through source directory
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process markdown files
		ext := filepath.Ext(path)
		if ext != ".md" {
			return nil
		}

		current++

		// Get relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		var dstPath string
		if dstDir == "" {
			// If target directory is empty, create translation file in source directory
			dir := filepath.Dir(path)
			base := filepath.Base(path)
			nameWithoutExt := strings.TrimSuffix(base, ext)
			dstPath = filepath.Join(dir, nameWithoutExt+"_"+targetLang+ext)
		} else {
			// If a different target directory is specified, use the specified directory structure
			dstPath = filepath.Join(dstDir, relPath)
		}

		t.progress(Progress{
			Total:      total,
			Current:    current,
			SourceFile: path,
			TargetFile: dstPath,
		})

		// Process file
		if err := ProcessFile(path, dstPath, targetLang, cfg, format, force); err != nil {
			return fmt.Errorf("failed to process file %s: %v", path, err)
		}

		return nil
	})
}
