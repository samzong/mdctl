package exporter

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// Merger 负责合并多个 Markdown 文件
type Merger struct {
	ShiftHeadingLevelBy int
	FileAsTitle         bool
	Logger              *log.Logger
	// 存储所有源文件的目录，用于后续设置 Pandoc 的资源路径
	SourceDirs []string
	// 是否启用详细日志
	Verbose bool
}

// Merge 合并多个 Markdown 文件到一个目标文件
func (m *Merger) Merge(sources []string, target string) error {
	// 如果没有设置日志记录器，创建一个默认的
	if m.Logger == nil {
		if m.Verbose {
			m.Logger = log.New(os.Stdout, "[MERGER] ", log.LstdFlags)
		} else {
			m.Logger = log.New(io.Discard, "", 0)
		}
	}

	if len(sources) == 0 {
		m.Logger.Println("Error: no source files provided")
		return fmt.Errorf("no source files provided")
	}

	m.Logger.Printf("Merging %d files into: %s", len(sources), target)
	var mergedContent strings.Builder

	// 初始化源目录列表
	m.SourceDirs = make([]string, 0, len(sources))
	sourceDirsMap := make(map[string]bool) // 用于去重

	// 处理每个源文件
	for i, source := range sources {
		m.Logger.Printf("Processing file %d/%d: %s", i+1, len(sources), source)

		// 获取源文件所在目录并添加到列表中（去重）
		sourceDir := filepath.Dir(source)
		if !sourceDirsMap[sourceDir] {
			sourceDirsMap[sourceDir] = true
			m.SourceDirs = append(m.SourceDirs, sourceDir)
		}

		// 读取文件内容
		content, err := os.ReadFile(source)
		if err != nil {
			m.Logger.Printf("Error reading file %s: %s", source, err)
			return fmt.Errorf("failed to read file %s: %s", source, err)
		}

		// 处理内容
		processedContent := string(content)

		// 确保内容是有效的 UTF-8
		if !utf8.ValidString(processedContent) {
			m.Logger.Printf("File %s contains invalid UTF-8, attempting to convert from GBK", source)
			// 尝试将内容从 GBK 转换为 UTF-8
			reader := transform.NewReader(bytes.NewReader(content), simplifiedchinese.GBK.NewDecoder())
			decodedContent, err := io.ReadAll(reader)
			if err != nil {
				m.Logger.Printf("Failed to decode content from file %s: %s", source, err)
				return fmt.Errorf("failed to decode content from file %s: %s", source, err)
			}
			processedContent = string(decodedContent)
			m.Logger.Printf("Successfully converted content from GBK to UTF-8")
		}

		// 移除 YAML front matter
		m.Logger.Println("Removing YAML front matter...")
		processedContent = removeYAMLFrontMatter(processedContent)

		// 处理图片路径
		m.Logger.Println("Processing image paths...")
		processedContent, err = processImagePaths(processedContent, source, m.Logger, m.Verbose)
		if err != nil {
			m.Logger.Printf("Error processing image paths: %s", err)
			return fmt.Errorf("failed to process image paths: %s", err)
		}

		// 调整标题级别
		if m.ShiftHeadingLevelBy != 0 {
			m.Logger.Printf("Shifting heading levels by %d", m.ShiftHeadingLevelBy)
			processedContent = ShiftHeadings(processedContent, m.ShiftHeadingLevelBy)
		}

		// 添加文件名作为标题
		if m.FileAsTitle {
			filename := filepath.Base(source)
			m.Logger.Printf("Adding filename as title: %s", filename)
			processedContent = AddTitleFromFilename(processedContent, filename, 1+m.ShiftHeadingLevelBy)
		}

		// 添加到合并内容
		m.Logger.Printf("Adding processed content to merged result (length: %d bytes)", len(processedContent))
		mergedContent.WriteString(processedContent)

		// 如果不是最后一个文件，添加分隔符
		if i < len(sources)-1 {
			mergedContent.WriteString("\n\n")
		}
	}

	// 最终内容
	finalContent := mergedContent.String()

	// 再次检查是否有任何 YAML 相关的问题
	m.Logger.Println("Sanitizing final content...")
	finalContent = sanitizeContent(finalContent)

	// 写入目标文件，确保使用 UTF-8 编码
	m.Logger.Printf("Writing merged content to target file: %s (size: %d bytes)", target, len(finalContent))
	err := os.WriteFile(target, []byte(finalContent), 0644)
	if err != nil {
		m.Logger.Printf("Error writing merged content: %s", err)
		return fmt.Errorf("failed to write merged content to %s: %s", target, err)
	}

	m.Logger.Printf("Successfully merged %d files into: %s", len(sources), target)
	return nil
}

// processImagePaths 处理 Markdown 中的图片路径，将相对路径转换为相对于执行命令位置的路径
func processImagePaths(content, sourcePath string, logger *log.Logger, verbose bool) (string, error) {
	// 如果没有设置日志记录器，创建一个默认的
	if logger == nil {
		if verbose {
			logger = log.New(os.Stdout, "[IMAGE] ", log.LstdFlags)
		} else {
			logger = log.New(io.Discard, "", 0)
		}
	}

	// 获取源文件所在目录
	sourceDir := filepath.Dir(sourcePath)
	if verbose {
		logger.Printf("处理图片路径: 源文件目录 = %s", sourceDir)
	}

	// 获取当前工作目录（执行命令的位置）
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("无法获取当前工作目录: %v", err)
	}
	if verbose {
		logger.Printf("当前工作目录 = %s", workingDir)
	}

	// 获取源文件目录的绝对路径
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return "", fmt.Errorf("无法获取源文件目录的绝对路径: %v", err)
	}
	if verbose {
		logger.Printf("源文件目录的绝对路径 = %s", absSourceDir)
	}

	// 匹配 Markdown 图片语法: ![alt](path)
	imageRegex := regexp.MustCompile(`!\[(.*?)\]\((.*?)\)`)

	// 替换所有图片路径
	processedContent := imageRegex.ReplaceAllStringFunc(content, func(match string) string {
		// 提取图片路径
		submatches := imageRegex.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match // 如果匹配不正确，保持原样
		}

		altText := submatches[1]
		imagePath := submatches[2]
		if verbose {
			logger.Printf("找到图片: alt = %s, 路径 = %s", altText, imagePath)
		}

		// 如果是网络图片（以 http:// 或 https:// 开头），保持原样
		if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
			if verbose {
				logger.Printf("保留网络图片路径: %s", imagePath)
			}
			return match
		}

		// 解析图片的绝对路径
		var absoluteImagePath string
		if filepath.IsAbs(imagePath) {
			absoluteImagePath = imagePath
		} else {
			// 对于相对路径，先转换为绝对路径
			absoluteImagePath = filepath.Join(absSourceDir, imagePath)
		}
		if verbose {
			logger.Printf("解析图片路径: 相对路径 = %s, 绝对路径 = %s", imagePath, absoluteImagePath)
		}

		// 检查图片文件是否存在
		if _, err := os.Stat(absoluteImagePath); os.IsNotExist(err) {
			if verbose {
				logger.Printf("图片不存在: %s", absoluteImagePath)
			}
			// 图片不存在，尝试在相邻目录查找
			// 例如，如果路径是 ../images/image.png，尝试在源文件目录的上一级目录的 images 子目录中查找
			if strings.HasPrefix(imagePath, "../") {
				parentDir := filepath.Dir(absSourceDir)
				relPath := strings.TrimPrefix(imagePath, "../")
				alternativePath := filepath.Join(parentDir, relPath)
				if verbose {
					logger.Printf("尝试替代路径: %s", alternativePath)
				}
				if _, err := os.Stat(alternativePath); err == nil {
					absoluteImagePath = alternativePath
					if verbose {
						logger.Printf("找到图片在替代路径: %s", absoluteImagePath)
					}
				} else {
					// 仍然找不到，保持原样
					if verbose {
						logger.Printf("图片在替代路径也不存在: %s", alternativePath)
					}
					return match
				}
			} else {
				// 找不到图片，保持原样
				return match
			}
		}

		// 计算图片相对于当前工作目录的路径
		relPath, err := filepath.Rel(workingDir, absoluteImagePath)
		if err != nil {
			if verbose {
				logger.Printf("无法计算相对路径，保留原始路径: %s, 错误: %v", imagePath, err)
			}
			return match
		}

		// 使用相对于当前工作目录的路径更新图片引用
		newRef := fmt.Sprintf("![%s](%s)", altText, relPath)
		if verbose {
			logger.Printf("更新图片引用: %s -> %s", match, newRef)
		}
		return newRef
	})

	return processedContent, nil
}

// removeYAMLFrontMatter 移除 YAML front matter
func removeYAMLFrontMatter(content string) string {
	// 匹配 YAML front matter
	yamlFrontMatterRegex := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n`)
	return yamlFrontMatterRegex.ReplaceAllString(content, "")
}

// sanitizeContent 清理内容，移除可能导致 Pandoc 解析错误的内容
func sanitizeContent(content string) string {
	// 移除可能导致 YAML 解析错误的行
	lines := strings.Split(content, "\n")
	var cleanedLines []string

	for _, line := range lines {
		// 跳过可能导致 YAML 解析错误的行
		if strings.Contains(line, ":") && !strings.Contains(line, ": ") && !strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "\t") {
			// 这种情况下冒号后面应该有空格，但没有，可能会导致 YAML 解析错误
			// 尝试修复它
			fixedLine := strings.Replace(line, ":", ": ", 1)
			cleanedLines = append(cleanedLines, fixedLine)
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "- ") && len(line) > 1 {
			// 这种情况下破折号后面应该有空格，但没有，可能会导致 YAML 解析错误
			// 尝试修复它
			fixedLine := strings.Replace(line, "-", "- ", 1)
			cleanedLines = append(cleanedLines, fixedLine)
		} else {
			cleanedLines = append(cleanedLines, line)
		}
	}

	return strings.Join(cleanedLines, "\n")
}
