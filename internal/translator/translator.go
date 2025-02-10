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

type FrontMatter struct {
	Translated bool `yaml:"translated"`
}

// Progress 用于跟踪翻译进度
type Progress struct {
	Total      int
	Current    int
	SourceFile string
	TargetFile string
}

func TranslateContent(content string, targetLang string, cfg *config.Config) (string, error) {
	prompt := strings.Replace(cfg.TranslatePrompt, "{TARGET_LANG}", targetLang, 1)

	messages := []OpenAIMessage{
		{Role: "system", Content: prompt},
		{Role: "user", Content: content},
	}

	reqBody := OpenAIRequest{
		Model:       cfg.ModelName,
		Messages:    messages,
		Temperature: cfg.Temperature,
		TopP:        cfg.TopP,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", cfg.OpenAIEndpointURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.OpenAIAPIKey)

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

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %v\nResponse body: %s", err, string(body))
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no translation result\nResponse body: %s", string(body))
	}

	return openAIResp.Choices[0].Message.Content, nil
}

func ProcessFile(srcPath, dstPath string, targetLang string, cfg *config.Config, force bool, progress *Progress) error {
	if progress != nil {
		progress.SourceFile = filepath.Base(srcPath)
		progress.TargetFile = filepath.Base(dstPath)
		if progress.Total > 1 {
			fmt.Printf("开始翻译 [%d/%d] %s...\n", progress.Current+1, progress.Total, progress.SourceFile)
		} else {
			fmt.Printf("开始翻译 %s...\n", progress.SourceFile)
		}
	}

	content, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	// 检查和更新 Front Matter
	parts := bytes.SplitN(content, []byte("---"), 3)
	if len(parts) == 3 {
		var frontMatter FrontMatter
		if err := yaml.Unmarshal(parts[1], &frontMatter); err != nil {
			return err
		}

		if frontMatter.Translated && !force {
			if progress != nil {
				if progress.Total > 1 {
					fmt.Printf("跳过 [%d/%d] %s (已翻译)\n", progress.Current+1, progress.Total, progress.SourceFile)
				} else {
					fmt.Printf("跳过 %s (已翻译)\n", progress.SourceFile)
				}
				progress.Current++
			}
			return fmt.Errorf("file already translated, use -F to force translate")
		}

		frontMatter.Translated = true
		newFrontMatter, err := yaml.Marshal(frontMatter)
		if err != nil {
			return err
		}

		translatedContent, err := TranslateContent(string(parts[2]), targetLang, cfg)
		if err != nil {
			return err
		}

		// 确保目标目录存在
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		// 写入新文件
		output := fmt.Sprintf("---\n%s---\n%s", string(newFrontMatter), translatedContent)
		if err := os.WriteFile(dstPath, []byte(output), 0644); err != nil {
			return err
		}

		if progress != nil {
			if progress.Total > 1 {
				fmt.Printf("✓ 完成翻译 [%d/%d] %s -> %s\n", progress.Current+1, progress.Total, progress.SourceFile, progress.TargetFile)
			} else {
				fmt.Printf("✓ 完成翻译 %s -> %s\n", progress.SourceFile, progress.TargetFile)
			}
			progress.Current++
		}
		return nil
	}

	// 如果没有 Front Matter，直接翻译整个文件
	translatedContent, err := TranslateContent(string(content), targetLang, cfg)
	if err != nil {
		return err
	}

	// 添加新的 Front Matter
	frontMatter := FrontMatter{Translated: true}
	newFrontMatter, err := yaml.Marshal(frontMatter)
	if err != nil {
		return err
	}

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}

	// 写入新文件
	output := fmt.Sprintf("---\n%s---\n%s", string(newFrontMatter), translatedContent)
	if err := os.WriteFile(dstPath, []byte(output), 0644); err != nil {
		return err
	}

	if progress != nil {
		if progress.Total > 1 {
			fmt.Printf("✓ 完成翻译 [%d/%d] %s -> %s\n", progress.Current+1, progress.Total, progress.SourceFile, progress.TargetFile)
		} else {
			fmt.Printf("✓ 完成翻译 %s -> %s\n", progress.SourceFile, progress.TargetFile)
		}
		progress.Current++
	}
	return nil
}

func ProcessDirectory(srcDir, dstDir string, targetLang string, cfg *config.Config, force bool) error {
	// 首先统计需要处理的文件数量
	var totalFiles int
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			totalFiles++
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("正在扫描目录 [%s]...\n", srcDir)
	fmt.Printf("找到 %d 个 Markdown 文件需要翻译\n", totalFiles)

	progress := &Progress{
		Total:   totalFiles,
		Current: 0,
	}

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 只处理 markdown 文件
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		// 计算目标文件路径
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// 如果源目录和目标目录相同，生成 name_lang.md 格式的文件
		if srcDir == dstDir {
			dir := filepath.Dir(path)
			base := filepath.Base(path)
			ext := filepath.Ext(base)
			nameWithoutExt := strings.TrimSuffix(base, ext)
			dstPath := filepath.Join(dir, nameWithoutExt+"_"+targetLang+ext)
			return ProcessFile(path, dstPath, targetLang, cfg, force, progress)
		}

		// 如果指定了不同的目标目录，使用指定的目录结构
		dstPath := filepath.Join(dstDir, relPath)
		return ProcessFile(path, dstPath, targetLang, cfg, force, progress)
	})

	if err == nil && totalFiles > 0 {
		fmt.Printf("所有文件翻译完成！共处理 %d 个文件\n", totalFiles)
	}

	return err
}
