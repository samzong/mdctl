package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samzong/mdctl/internal/config"
	"github.com/samzong/mdctl/internal/translator"
	"github.com/spf13/cobra"
)

var (
	fromPath string
	toPath   string
	locale   string
	force    bool
	format   bool
)

// 生成目标文件路径
func generateTargetPath(sourcePath, lang string) string {
	dir := filepath.Dir(sourcePath)
	base := filepath.Base(sourcePath)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)
	return filepath.Join(dir, nameWithoutExt+"_"+lang+ext)
}

var translateCmd = &cobra.Command{
	Use:   "translate",
	Short: "Translate markdown files using AI models",
	Long: `Translate markdown files or directories to specified language using AI models.

Supported AI Models:
  - OpenAI API (Current)
  - Ollama (Coming Soon)
  - Google Gemini (Coming Soon)
  - Anthropic Claude (Coming Soon)
	
Supported Languages:
  ar (العربية), de (Deutsch), en (English), es (Español), fr (Français),
  hi (हिन्दी), it (Italiano), ja (日本語), ko (한국어), pt (Português),
  ru (Русский), th (ไทย), vi (Tiếng Việt), zh (中文)

Examples:
  # Translate a single file to Chinese
  mdctl translate -f README.md -l zh

  # Translate a directory to Japanese
  mdctl translate -f docs -l ja

  # Force translate an already translated file
  mdctl translate -f README.md -l ko -F

  # Format markdown content after translation
  mdctl translate -f README.md -l zh -m

  # Translate to a specific output path
  mdctl translate -f docs -l fr -t translated_docs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}

		// 验证语言选项
		if !translator.IsLanguageSupported(locale) {
			return fmt.Errorf("unsupported locale: %s\nSupported languages: %s",
				locale,
				translator.GetSupportedLanguages())
		}

		// 检查源路径是否存在
		if _, err := os.Stat(fromPath); os.IsNotExist(err) {
			return fmt.Errorf("source path does not exist: %s", fromPath)
		}

		// 获取源路径的绝对路径
		srcAbs, err := filepath.Abs(fromPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %v", err)
		}

		// 检查是文件还是目录
		fi, err := os.Stat(srcAbs)
		if err != nil {
			return fmt.Errorf("failed to get file info: %v", err)
		}

		if fi.IsDir() {
			// 如果是目录，且没有指定目标路径，则使用相同的目录结构
			if toPath == "" {
				return translator.ProcessDirectory(srcAbs, srcAbs, locale, cfg, force, format)
			}
			// 如果指定了目标路径，使用指定的路径
			dstAbs, err := filepath.Abs(toPath)
			if err != nil {
				return fmt.Errorf("failed to get absolute path: %v", err)
			}
			return translator.ProcessDirectory(srcAbs, dstAbs, locale, cfg, force, format)
		}

		// 处理单个文件
		var dstAbs string
		if toPath == "" {
			// 如果没有指定目标路径，在源文件同目录生成 name_lang.md
			dstAbs = generateTargetPath(srcAbs, locale)
		} else {
			// 如果指定了目标路径，使用指定的路径
			dstAbs, err = filepath.Abs(toPath)
			if err != nil {
				return fmt.Errorf("failed to get absolute path: %v", err)
			}
		}

		return translator.ProcessFile(srcAbs, dstAbs, locale, cfg, format, force)
	},
}

func init() {
	rootCmd.AddCommand(translateCmd)

	translateCmd.Flags().StringVarP(&fromPath, "from", "f", "", "Source file or directory path")
	translateCmd.Flags().StringVarP(&toPath, "to", "t", "", "Target file or directory path (optional, default: generate in same directory as source)")
	translateCmd.Flags().StringVarP(&locale, "locales", "l", "", "Target language code (e.g., zh, en, ja, ko, fr, de, es, etc.)")
	translateCmd.Flags().BoolVarP(&force, "force", "F", false, "Force translate even if already translated")
	translateCmd.Flags().BoolVarP(&format, "format", "m", false, "Format markdown content after translation")

	translateCmd.MarkFlagRequired("from")
	translateCmd.MarkFlagRequired("locales")
}
