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

// SupportedLanguages 定义支持的语言映射
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

// IsLanguageSupported 检查语言是否支持
func IsLanguageSupported(lang string) bool {
	_, ok := SupportedLanguages[lang]
	return ok
}

// GetSupportedLanguages 获取支持的语言列表
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

// Progress 用于跟踪翻译进度
type Progress struct {
	Total      int
	Current    int
	SourceFile string
	TargetFile string
}

// ProgressCallback 进度回调函数类型
type ProgressCallback func(progress Progress)

// Translator 翻译器结构体
type Translator struct {
	config   *config.Config
	format   bool
	progress ProgressCallback
}

// New 创建新的翻译器实例
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

// TranslateContent 翻译内容
func (t *Translator) TranslateContent(content string, lang string) (string, error) {
	// 移除可能存在的 front matter
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

	// 获取翻译后的内容
	translatedContent := response.Choices[0].Message.Content

	// 移除可能的 markdown 代码块标记
	translatedContent = strings.TrimPrefix(translatedContent, "```markdown\n")
	translatedContent = strings.TrimSuffix(translatedContent, "\n```")
	translatedContent = strings.TrimSpace(translatedContent)

	// 如果启用了格式化，对翻译后的内容进行格式化
	if t.format {
		formatter := markdownfmt.New(true)
		translatedContent = formatter.Format(translatedContent)
	}

	return translatedContent, nil
}

// removeFrontMatter 移除内容中的 front matter
func removeFrontMatter(content string) string {
	// 如果内容以 --- 开头，说明可能包含 front matter
	trimmedContent := strings.TrimSpace(content)
	if strings.HasPrefix(trimmedContent, "---") {
		parts := strings.SplitN(trimmedContent, "---", 3)
		if len(parts) >= 3 {
			return strings.TrimSpace(parts[2])
		}
	}
	return content
}

// ProcessFile 处理单个文件的翻译
func ProcessFile(srcPath, dstPath, targetLang string, cfg *config.Config, format bool, force bool) error {
	t := New(cfg, format)

	// 检查目标路径是否是目录
	dstInfo, err := os.Stat(dstPath)
	if err == nil && dstInfo.IsDir() {
		dstPath = filepath.Join(dstPath, filepath.Base(srcPath))
	}

	// 检查目标文件是否已经存在
	if _, err := os.Stat(dstPath); err == nil {
		dstContent, err := os.ReadFile(dstPath)
		if err != nil {
			return fmt.Errorf("failed to read target file: %v", err)
		}

		// 检查是否已翻译
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

	// 读取源文件内容
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %v", err)
	}

	// 解析 front matter
	var frontMatter map[string]interface{}
	contentToTranslate := string(content)

	// 检查并解析 front matter
	if strings.HasPrefix(contentToTranslate, "---\n") {
		parts := strings.SplitN(contentToTranslate[4:], "\n---\n", 2)
		if len(parts) == 2 {
			if err := yaml.Unmarshal([]byte(parts[0]), &frontMatter); err != nil {
				return fmt.Errorf("failed to parse front matter: %v", err)
			}
			contentToTranslate = parts[1]
		}
	}

	// 翻译内容
	translatedContent, err := t.TranslateContent(contentToTranslate, targetLang)
	if err != nil {
		return fmt.Errorf("failed to translate content: %v", err)
	}

	// 更新 front matter
	if frontMatter == nil {
		frontMatter = make(map[string]interface{})
	}
	frontMatter["translated"] = true

	// 生成新的文件内容
	frontMatterBytes, err := yaml.Marshal(frontMatter)
	if err != nil {
		return fmt.Errorf("failed to marshal front matter: %v", err)
	}

	newContent := fmt.Sprintf("---\n%s\n---\n\n%s", string(frontMatterBytes), translatedContent)

	// 创建目标文件的目录（如果不存在）
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	// 写入翻译后的内容到目标文件
	if err := os.WriteFile(dstPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write target file: %v", err)
	}

	return nil
}

// ProcessDirectory 处理目录中的所有 markdown 文件
func ProcessDirectory(srcDir, dstDir string, targetLang string, cfg *config.Config, force bool, format bool) error {
	// 首先计算需要处理的文件总数
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

	// 创建翻译器实例
	t := New(cfg, format)
	current := 0

	// 遍历源目录
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 只处理 markdown 文件
		ext := filepath.Ext(path)
		if ext != ".md" {
			return nil
		}

		current++

		// 获取相对路径
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		var dstPath string
		if dstDir == "" {
			// 如果目标目录为空，在源文件所在目录创建翻译文件
			dir := filepath.Dir(path)
			base := filepath.Base(path)
			nameWithoutExt := strings.TrimSuffix(base, ext)
			dstPath = filepath.Join(dir, nameWithoutExt+"_"+targetLang+ext)
		} else {
			// 如果指定了不同的目标目录，使用指定的目录结构
			dstPath = filepath.Join(dstDir, relPath)
		}

		t.progress(Progress{
			Total:      total,
			Current:    current,
			SourceFile: path,
			TargetFile: dstPath,
		})

		// 处理文件
		if err := ProcessFile(path, dstPath, targetLang, cfg, format, force); err != nil {
			return fmt.Errorf("failed to process file %s: %v", path, err)
		}

		return nil
	})
}
