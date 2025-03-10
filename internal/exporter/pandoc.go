package exporter

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// PandocExporter 使用 Pandoc 工具实现导出功能
type PandocExporter struct {
	PandocPath string
	Logger     *log.Logger
}

// Export 使用 Pandoc 将 Markdown 导出为其他格式
func (e *PandocExporter) Export(input, output string, options ExportOptions) error {
	// 如果没有设置日志记录器，创建一个默认的
	if e.Logger == nil {
		if options.Verbose {
			e.Logger = log.New(os.Stdout, "[PANDOC] ", log.LstdFlags)
		} else {
			e.Logger = log.New(io.Discard, "", 0)
		}
	}

	e.Logger.Printf("Starting Pandoc export: %s -> %s", input, output)

	// 确保输出路径是绝对路径
	absOutput, err := filepath.Abs(output)
	if err != nil {
		e.Logger.Printf("Failed to get absolute path for output: %s", err)
		return fmt.Errorf("failed to get absolute path for output: %s", err)
	}
	e.Logger.Printf("Using absolute output path: %s", absOutput)

	// 创建一个临时文件，用于处理后的内容
	e.Logger.Println("Creating sanitized copy of input file...")
	tempFile, err := createSanitizedCopy(input, e.Logger)
	if err != nil {
		e.Logger.Printf("Failed to create sanitized copy: %s", err)
		return fmt.Errorf("failed to create sanitized copy: %s", err)
	}
	defer os.Remove(tempFile)
	e.Logger.Printf("Sanitized copy created: %s", tempFile)

	// 构建 Pandoc 命令参数
	e.Logger.Println("Building Pandoc command arguments...")
	args := []string{
		tempFile,
		"-o", absOutput,
		"--standalone",
		"--pdf-engine=xelatex",
		"-V", "mainfont=SimSun", // 使用宋体作为主要字体
		"--wrap=preserve",
		"--embed-resources", // 嵌入资源到输出文件
	}

	// 添加资源路径参数，帮助 Pandoc 找到图片
	// 收集所有可能的资源路径
	resourcePaths := make(map[string]bool)

	// 添加输入文件所在目录
	inputDir := filepath.Dir(input)
	resourcePaths[inputDir] = true
	e.Logger.Printf("添加输入文件目录到资源路径: %s", inputDir)

	// 添加当前工作目录
	workingDir, err := os.Getwd()
	if err == nil {
		resourcePaths[workingDir] = true
		e.Logger.Printf("添加当前工作目录到资源路径: %s", workingDir)
	}

	// 添加输出文件所在目录
	outputDir := filepath.Dir(absOutput)
	resourcePaths[outputDir] = true
	e.Logger.Printf("添加输出文件目录到资源路径: %s", outputDir)

	// 添加源文件目录到资源路径
	if len(options.SourceDirs) > 0 {
		for _, dir := range options.SourceDirs {
			resourcePaths[dir] = true
			e.Logger.Printf("添加源文件目录到资源路径: %s", dir)
		}
	}

	// 将所有资源路径添加到 Pandoc 参数中
	for path := range resourcePaths {
		args = append(args, "--resource-path", path)
	}

	// 添加模板参数
	if options.Template != "" {
		e.Logger.Printf("Using template: %s", options.Template)
		args = append(args, "--reference-doc", options.Template)
	}

	// 添加目录参数
	if options.GenerateToc {
		e.Logger.Println("Generating table of contents")
		args = append(args, "--toc")

		// 添加目录深度参数
		if options.TocDepth > 0 {
			e.Logger.Printf("Setting table of contents depth to: %d", options.TocDepth)
			args = append(args, "--toc-depth", fmt.Sprintf("%d", options.TocDepth))
		}
	}

	// 添加标题层级偏移参数
	if options.ShiftHeadingLevelBy != 0 {
		e.Logger.Printf("Shifting heading levels by: %d", options.ShiftHeadingLevelBy)
		args = append(args, "--shift-heading-level-by", fmt.Sprintf("%d", options.ShiftHeadingLevelBy))
	}

	// 根据输出格式添加特定参数
	e.Logger.Printf("Using output format: %s", options.Format)
	switch options.Format {
	case "pdf":
		// PDF 格式需要特殊处理中文
		e.Logger.Println("Adding PDF-specific parameters for CJK support")
		args = append(args,
			"-V", "CJKmainfont=SimSun", // 中日韩字体设置
			"-V", "documentclass=article",
			"-V", "geometry=margin=1in")
	case "epub":
		// EPUB 格式的特殊参数
		e.Logger.Println("Adding EPUB-specific parameters")
		args = append(args, "--epub-chapter-level=1")
	}

	// 执行 Pandoc 命令
	e.Logger.Printf("Executing Pandoc command: %s %s", e.PandocPath, strings.Join(args, " "))
	cmd := exec.Command(e.PandocPath, args...)

	// 设置工作目录为输入文件所在目录，这有助于 Pandoc 找到相对路径的图片
	cmd.Dir = inputDir

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		// 如果执行失败，尝试查看输入文件内容以便调试
		e.Logger.Printf("Pandoc execution failed: %s", err)
		e.Logger.Printf("Pandoc output: %s", string(outputBytes))

		inputContent, readErr := os.ReadFile(tempFile)
		if readErr == nil {
			// 只显示前 500 个字符，避免输出过多
			contentPreview := string(inputContent)
			if len(contentPreview) > 500 {
				contentPreview = contentPreview[:500] + "..."
			}
			e.Logger.Printf("Input file preview:\n%s", contentPreview)
			return fmt.Errorf("pandoc execution failed: %s\nOutput: %s\nCommand: %s\nInput file preview:\n%s",
				err, string(outputBytes), strings.Join(cmd.Args, " "), contentPreview)
		}

		return fmt.Errorf("pandoc execution failed: %s\nOutput: %s\nCommand: %s",
			err, string(outputBytes), strings.Join(cmd.Args, " "))
	}

	e.Logger.Printf("Pandoc export completed successfully: %s", output)
	return nil
}

// createSanitizedCopy 创建一个经过清理的临时文件副本
func createSanitizedCopy(inputFile string, logger *log.Logger) (string, error) {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}

	// 读取输入文件内容
	logger.Printf("Reading input file: %s", inputFile)
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return "", fmt.Errorf("failed to read input file: %s", err)
	}

	// 将内容转换为字符串
	contentStr := string(content)

	// 移除 YAML front matter
	logger.Println("Removing YAML front matter...")
	yamlFrontMatterRegex := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n`)
	if yamlFrontMatterRegex.MatchString(contentStr) {
		logger.Println("YAML front matter found, removing it")
		contentStr = yamlFrontMatterRegex.ReplaceAllString(contentStr, "")
	}

	// 修复可能导致 YAML 解析错误的行
	logger.Println("Fixing potential YAML parsing issues...")
	lines := strings.Split(contentStr, "\n")
	var cleanedLines []string
	fixedLines := 0

	for _, line := range lines {
		// 跳过可能导致 YAML 解析错误的行
		if strings.Contains(line, ":") && !strings.Contains(line, ": ") && !strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "\t") {
			// 这种情况下冒号后面应该有空格，但没有，可能会导致 YAML 解析错误
			// 尝试修复它
			fixedLine := strings.Replace(line, ":", ": ", 1)
			cleanedLines = append(cleanedLines, fixedLine)
			fixedLines++
			logger.Printf("Fixed line with missing space after colon: %s -> %s", line, fixedLine)
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "- ") && len(line) > 1 {
			// 这种情况下破折号后面应该有空格，但没有，可能会导致 YAML 解析错误
			// 尝试修复它
			fixedLine := strings.Replace(line, "-", "- ", 1)
			cleanedLines = append(cleanedLines, fixedLine)
			fixedLines++
			logger.Printf("Fixed line with missing space after dash: %s -> %s", line, fixedLine)
		} else {
			cleanedLines = append(cleanedLines, line)
		}
	}

	logger.Printf("Fixed %d lines with potential YAML issues", fixedLines)

	// 创建临时文件
	tempDir := os.TempDir()
	tempFilePath := filepath.Join(tempDir, "mdctl-sanitized-"+filepath.Base(inputFile))

	// 写入清理后的内容
	logger.Printf("Writing sanitized content to temporary file: %s", tempFilePath)
	err = os.WriteFile(tempFilePath, []byte(strings.Join(cleanedLines, "\n")), 0644)
	if err != nil {
		return "", err
	}

	return tempFilePath, nil
}

// preprocessInputFile 预处理输入文件，移除可能导致 Pandoc 解析错误的内容
func preprocessInputFile(filePath string) error {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// 检查是否有不规范的 YAML front matter
	yamlFrontMatterRegex := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n`)
	if yamlFrontMatterRegex.MatchString(contentStr) {
		// 提取 YAML front matter 内容
		matches := yamlFrontMatterRegex.FindStringSubmatch(contentStr)
		if len(matches) > 1 {
			yamlContent := matches[1]

			// 检查 YAML 内容是否有格式问题
			if strings.Contains(yamlContent, "\n-") && !strings.Contains(yamlContent, "\n- ") {
				// 修复格式问题：确保 - 后面有空格
				fixedYaml := strings.ReplaceAll(yamlContent, "\n-", "\n- ")
				fixedContent := strings.Replace(contentStr, yamlContent, fixedYaml, 1)

				// 写回文件
				return os.WriteFile(filePath, []byte(fixedContent), 0644)
			}
		}

		// 如果 YAML 格式可能有其他问题，直接移除整个 front matter
		processedContent := yamlFrontMatterRegex.ReplaceAllString(contentStr, "")
		return os.WriteFile(filePath, []byte(processedContent), 0644)
	}

	return nil
}

// CheckPandocAvailability 检查 Pandoc 是否可用
func CheckPandocAvailability() error {
	cmd := exec.Command("pandoc", "--version")
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pandoc is not available: %s\n\nPlease install Pandoc to use the export feature:\n\n"+
			"macOS: brew install pandoc\n"+
			"Ubuntu/Debian: sudo apt-get install pandoc\n"+
			"Windows: choco install pandoc\n\n"+
			"For more information, visit: https://pandoc.org/installing.html", err)
	}

	// 检查版本
	versionStr := string(outputBytes)
	if !strings.Contains(versionStr, "pandoc") {
		return fmt.Errorf("unexpected pandoc version output: %s", versionStr)
	}

	return nil
}
