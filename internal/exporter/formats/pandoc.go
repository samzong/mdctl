package formats

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

// PandocFormatter 是基于Pandoc的格式化器基类
type PandocFormatter struct {
	PandocPath string
	Logger     *log.Logger
	FormatName string
}

// NewPandocFormatter 创建一个新的Pandoc格式化器
func NewPandocFormatter(formatName string, logger *log.Logger) *PandocFormatter {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}

	return &PandocFormatter{
		PandocPath: "pandoc", // 默认使用系统PATH中的pandoc
		Logger:     logger,
		FormatName: formatName,
	}
}

// GetFormatName 返回格式名称
func (f *PandocFormatter) GetFormatName() string {
	return f.FormatName
}

// ValidateOptions 验证导出选项是否有效
func (f *PandocFormatter) ValidateOptions(options ExportOptions) error {
	// 基本验证，子类可以覆盖此方法进行更具体的验证
	return nil
}

// Format 将输入文件转换为指定格式
func (f *PandocFormatter) Format(input, output string, options ExportOptions) error {
	// 如果没有设置日志记录器，创建一个默认的
	if f.Logger == nil {
		if options.Verbose {
			f.Logger = log.New(os.Stdout, "[PANDOC] ", log.LstdFlags)
		} else {
			f.Logger = log.New(io.Discard, "", 0)
		}
	}

	f.Logger.Printf("Starting Pandoc export: %s -> %s", input, output)

	// 确保输出路径是绝对路径
	absOutput, err := filepath.Abs(output)
	if err != nil {
		f.Logger.Printf("Failed to get absolute path for output: %s", err)
		return fmt.Errorf("failed to get absolute path for output: %s", err)
	}
	f.Logger.Printf("Using absolute output path: %s", absOutput)

	// 创建一个临时文件，用于处理后的内容
	f.Logger.Println("Creating sanitized copy of input file...")
	tempFile, err := createSanitizedCopy(input, f.Logger)
	if err != nil {
		f.Logger.Printf("Failed to create sanitized copy: %s", err)
		return fmt.Errorf("failed to create sanitized copy: %s", err)
	}
	defer os.Remove(tempFile)
	f.Logger.Printf("Sanitized copy created: %s", tempFile)

	// 构建基本的Pandoc命令参数
	args := f.buildBasicPandocArgs(tempFile, absOutput, options)

	// 允许子类添加格式特定的参数
	args = f.addFormatSpecificArgs(args, options)

	// 执行 Pandoc 命令
	f.Logger.Printf("Executing Pandoc command: %s %s", f.PandocPath, strings.Join(args, " "))
	cmd := exec.Command(f.PandocPath, args...)

	// 设置工作目录为输入文件所在目录，这有助于 Pandoc 找到相对路径的图片
	cmd.Dir = filepath.Dir(input)

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		// 如果执行失败，尝试查看输入文件内容以便调试
		f.Logger.Printf("Pandoc execution failed: %s", err)
		f.Logger.Printf("Pandoc output: %s", string(outputBytes))

		inputContent, readErr := os.ReadFile(tempFile)
		if readErr == nil {
			// 只显示前 500 个字符，避免输出过多
			contentPreview := string(inputContent)
			if len(contentPreview) > 500 {
				contentPreview = contentPreview[:500] + "..."
			}
			f.Logger.Printf("Input file preview:\n%s", contentPreview)
			return fmt.Errorf("pandoc execution failed: %s\nOutput: %s\nCommand: %s\nInput file preview:\n%s",
				err, string(outputBytes), strings.Join(cmd.Args, " "), contentPreview)
		}

		return fmt.Errorf("pandoc execution failed: %s\nOutput: %s\nCommand: %s",
			err, string(outputBytes), strings.Join(cmd.Args, " "))
	}

	f.Logger.Printf("Pandoc export completed successfully: %s", output)
	return nil
}

// buildBasicPandocArgs 构建基本的Pandoc命令参数
func (f *PandocFormatter) buildBasicPandocArgs(input, output string, options ExportOptions) []string {
	args := []string{
		input,
		"-o", output,
		"--standalone",
		"--wrap=preserve",
		"--embed-resources", // 嵌入资源到输出文件
	}

	// 添加资源路径参数，帮助 Pandoc 找到图片
	// 收集所有可能的资源路径
	resourcePaths := make(map[string]bool)

	// 添加输入文件所在目录
	inputDir := filepath.Dir(input)
	resourcePaths[inputDir] = true
	f.Logger.Printf("添加输入文件目录到资源路径: %s", inputDir)

	// 添加当前工作目录
	workingDir, err := os.Getwd()
	if err == nil {
		resourcePaths[workingDir] = true
		f.Logger.Printf("添加当前工作目录到资源路径: %s", workingDir)
	}

	// 添加输出文件所在目录
	outputDir := filepath.Dir(output)
	resourcePaths[outputDir] = true
	f.Logger.Printf("添加输出文件目录到资源路径: %s", outputDir)

	// 添加源文件目录到资源路径
	if len(options.SourceDirs) > 0 {
		for _, dir := range options.SourceDirs {
			resourcePaths[dir] = true
			f.Logger.Printf("添加源文件目录到资源路径: %s", dir)
		}
	}

	// 将所有资源路径添加到 Pandoc 参数中
	for path := range resourcePaths {
		args = append(args, "--resource-path", path)
	}

	// 添加模板参数
	if options.Template != "" {
		f.Logger.Printf("Using template: %s", options.Template)
		args = append(args, "--reference-doc", options.Template)
	}

	// 添加目录参数
	if options.GenerateToc {
		f.Logger.Println("Generating table of contents")
		args = append(args, "--toc")

		// 添加目录深度参数
		if options.TocDepth > 0 {
			f.Logger.Printf("Setting table of contents depth to: %d", options.TocDepth)
			args = append(args, "--toc-depth", fmt.Sprintf("%d", options.TocDepth))
		}
	}

	// 添加标题层级偏移参数
	if options.ShiftHeadingLevelBy != 0 {
		f.Logger.Printf("Shifting heading levels by: %d", options.ShiftHeadingLevelBy)
		args = append(args, "--shift-heading-level-by", fmt.Sprintf("%d", options.ShiftHeadingLevelBy))
	}

	return args
}

// addFormatSpecificArgs 添加格式特定的参数，子类可以覆盖此方法
func (f *PandocFormatter) addFormatSpecificArgs(args []string, options ExportOptions) []string {
	// 基类不添加任何特定参数，由子类实现
	return args
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

// CheckPandocAvailability 检查 Pandoc 是否可用
func CheckPandocAvailability() error {
	cmd := exec.Command("pandoc", "--version")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("pandoc is not available: %s\nPlease install Pandoc from https://pandoc.org/installing.html", err)
	}
	return nil
}
