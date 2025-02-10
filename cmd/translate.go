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

// Generate target file path
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

		// Validate language option
		if !translator.IsLanguageSupported(locale) {
			return fmt.Errorf("unsupported locale: %s\nSupported languages: %s",
				locale,
				translator.GetSupportedLanguages())
		}

		// Check if source path exists
		if _, err := os.Stat(fromPath); os.IsNotExist(err) {
			return fmt.Errorf("source path does not exist: %s", fromPath)
		}

		// Get absolute path of source path
		srcAbs, err := filepath.Abs(fromPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %v", err)
		}

		// Check if it's a file or directory
		fi, err := os.Stat(srcAbs)
		if err != nil {
			return fmt.Errorf("failed to get file info: %v", err)
		}

		if fi.IsDir() {
			// If it's a directory and no target path specified, use the same directory structure
			if toPath == "" {
				return translator.ProcessDirectory(srcAbs, srcAbs, locale, cfg, force, format)
			}
			// If target path is specified, use the specified path
			dstAbs, err := filepath.Abs(toPath)
			if err != nil {
				return fmt.Errorf("failed to get absolute path: %v", err)
			}
			return translator.ProcessDirectory(srcAbs, dstAbs, locale, cfg, force, format)
		}

		// Process single file
		var dstAbs string
		if toPath == "" {
			// If no target path specified, generate name_lang.md in the same directory as source
			dstAbs = generateTargetPath(srcAbs, locale)
		} else {
			// If target path specified, use the specified path
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
