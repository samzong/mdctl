package exporter

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/samzong/mdctl/internal/exporter/sitereader"
)

// ExportOptions 定义导出选项
type ExportOptions struct {
	Template            string      // Word 模板文件路径
	GenerateToc         bool        // 是否生成目录
	ShiftHeadingLevelBy int         // 标题层级偏移量
	FileAsTitle         bool        // 是否使用文件名作为章节标题
	Format              string      // 输出格式 (docx, pdf, epub)
	SiteType            string      // 站点类型 (mkdocs, hugo, docusaurus)
	Verbose             bool        // 是否启用详细日志
	Logger              *log.Logger // 日志记录器
	SourceDirs          []string    // 源文件所在的目录列表，用于处理图片路径
	TocDepth            int         // 目录深度，默认为 3
	NavPath             string      // 指定要导出的导航路径
}

// Exporter 定义导出器接口
type Exporter interface {
	Export(input string, output string, options ExportOptions) error
}

// DefaultExporter 是默认的导出器实现
type DefaultExporter struct {
	pandocPath string
	logger     *log.Logger
}

// NewExporter 创建一个新的导出器
func NewExporter() *DefaultExporter {
	return &DefaultExporter{
		pandocPath: "pandoc", // 默认使用系统 PATH 中的 pandoc
		logger:     log.New(os.Stdout, "[EXPORTER] ", log.LstdFlags),
	}
}

// ExportFile 导出单个 Markdown 文件
func (e *DefaultExporter) ExportFile(input, output string, options ExportOptions) error {
	// 设置日志记录器
	if options.Logger != nil {
		e.logger = options.Logger
	} else if !options.Verbose {
		e.logger = log.New(io.Discard, "", 0)
	}

	e.logger.Printf("Exporting file: %s -> %s", input, output)

	// 检查文件是否存在
	if _, err := os.Stat(input); os.IsNotExist(err) {
		e.logger.Printf("Error: input file does not exist: %s", input)
		return fmt.Errorf("input file does not exist: %s", input)
	}
	e.logger.Printf("Input file exists: %s", input)

	// 创建输出目录（如果不存在）
	outputDir := filepath.Dir(output)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		e.logger.Printf("Error: failed to create output directory: %s", err)
		return fmt.Errorf("failed to create output directory: %s", err)
	}
	e.logger.Printf("Output directory created/verified: %s", outputDir)

	// 添加源文件目录到 SourceDirs
	sourceDir := filepath.Dir(input)
	if options.SourceDirs == nil {
		options.SourceDirs = []string{sourceDir}
	} else {
		// 检查是否已经存在
		found := false
		for _, dir := range options.SourceDirs {
			if dir == sourceDir {
				found = true
				break
			}
		}
		if !found {
			options.SourceDirs = append(options.SourceDirs, sourceDir)
		}
	}
	e.logger.Printf("Added source directory to resource paths: %s", sourceDir)

	// 使用 Pandoc 导出
	e.logger.Println("Starting Pandoc export process...")
	pandocExporter := &PandocExporter{
		PandocPath: e.pandocPath,
		Logger:     e.logger,
	}
	err := pandocExporter.Export(input, output, options)
	if err != nil {
		e.logger.Printf("Pandoc export failed: %s", err)
		return err
	}

	e.logger.Printf("File export completed successfully: %s", output)
	return nil
}

// ExportDirectory 导出目录中的 Markdown 文件
func (e *DefaultExporter) ExportDirectory(inputDir, output string, options ExportOptions) error {
	// 设置日志记录器
	if options.Logger != nil {
		e.logger = options.Logger
	} else if !options.Verbose {
		e.logger = log.New(io.Discard, "", 0)
	}

	e.logger.Printf("Exporting directory: %s -> %s", inputDir, output)

	// 检查目录是否存在
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		e.logger.Printf("Error: input directory does not exist: %s", inputDir)
		return fmt.Errorf("input directory does not exist: %s", inputDir)
	}
	e.logger.Printf("Input directory exists: %s", inputDir)

	// 创建输出目录（如果不存在）
	outputDir := filepath.Dir(output)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		e.logger.Printf("Error: failed to create output directory: %s", err)
		return fmt.Errorf("failed to create output directory: %s", err)
	}
	e.logger.Printf("Output directory created/verified: %s", outputDir)

	// 初始化 SourceDirs（如果为 nil）
	if options.SourceDirs == nil {
		options.SourceDirs = []string{inputDir}
	} else {
		// 检查是否已经存在
		found := false
		for _, dir := range options.SourceDirs {
			if dir == inputDir {
				found = true
				break
			}
		}
		if !found {
			options.SourceDirs = append(options.SourceDirs, inputDir)
		}
	}
	e.logger.Printf("Added input directory to resource paths: %s", inputDir)

	// 根据站点类型选择不同的处理方式
	var files []string
	var err error

	if options.SiteType != "" && options.SiteType != "basic" {
		// 使用站点读取器获取文件列表
		e.logger.Printf("Using site reader for site type: %s", options.SiteType)
		reader, err := sitereader.GetSiteReader(options.SiteType, options.Verbose, e.logger)
		if err != nil {
			e.logger.Printf("Error getting site reader: %s", err)
			return err
		}

		// 检测是否为指定类型的站点
		e.logger.Printf("Detecting if directory is a %s site...", options.SiteType)
		if !reader.Detect(inputDir) {
			e.logger.Printf("Error: directory %s does not appear to be a %s site", inputDir, options.SiteType)
			return fmt.Errorf("directory %s does not appear to be a %s site", inputDir, options.SiteType)
		}
		e.logger.Printf("Directory confirmed as %s site", options.SiteType)

		e.logger.Println("Reading site structure...")
		files, err = reader.ReadStructure(inputDir, "", options.NavPath)
		if err != nil {
			e.logger.Printf("Error reading site structure: %s", err)
			return err
		}
		e.logger.Printf("Found %d files in site structure", len(files))
	} else {
		// 基础目录模式：按文件名排序
		e.logger.Println("Using basic directory mode, sorting files by name")
		files, err = GetMarkdownFilesInDir(inputDir)
		if err != nil {
			e.logger.Printf("Error getting markdown files: %s", err)
			return err
		}
		e.logger.Printf("Found %d markdown files in directory", len(files))
	}

	if len(files) == 0 {
		e.logger.Printf("Error: no markdown files found in directory: %s", inputDir)
		return fmt.Errorf("no markdown files found in directory: %s", inputDir)
	}

	// 如果只有一个文件，直接导出
	if len(files) == 1 {
		e.logger.Printf("Only one file found, exporting directly: %s", files[0])
		return e.ExportFile(files[0], output, options)
	}

	// 合并多个文件
	e.logger.Printf("Merging %d files...", len(files))
	merger := &Merger{
		ShiftHeadingLevelBy: options.ShiftHeadingLevelBy,
		FileAsTitle:         options.FileAsTitle,
		Logger:              e.logger,
		SourceDirs:          make([]string, 0),
		Verbose:             options.Verbose,
	}

	// 创建临时文件
	e.logger.Println("Creating temporary file for merged content...")
	tempFile, err := os.CreateTemp("", "mdctl-merged-*.md")
	if err != nil {
		e.logger.Printf("Error creating temporary file: %s", err)
		return fmt.Errorf("failed to create temporary file: %s", err)
	}
	tempFilePath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempFilePath)
	e.logger.Printf("Temporary file created: %s", tempFilePath)

	// 合并文件
	e.logger.Println("Merging files...")
	if err := merger.Merge(files, tempFilePath); err != nil {
		e.logger.Printf("Error merging files: %s", err)
		return fmt.Errorf("failed to merge files: %s", err)
	}
	e.logger.Println("Files merged successfully")

	// 将合并器收集的源目录添加到选项中
	if merger.SourceDirs != nil && len(merger.SourceDirs) > 0 {
		e.logger.Printf("Adding %d source directories from merger", len(merger.SourceDirs))
		for _, dir := range merger.SourceDirs {
			// 检查是否已经存在
			found := false
			for _, existingDir := range options.SourceDirs {
				if existingDir == dir {
					found = true
					break
				}
			}
			if !found {
				options.SourceDirs = append(options.SourceDirs, dir)
				e.logger.Printf("Added source directory: %s", dir)
			}
		}
	}

	// 导出合并后的文件
	e.logger.Println("Starting Pandoc export process...")
	pandocExporter := &PandocExporter{
		PandocPath: e.pandocPath,
		Logger:     e.logger,
	}
	err = pandocExporter.Export(tempFilePath, output, options)
	if err != nil {
		e.logger.Printf("Pandoc export failed: %s", err)
		return err
	}

	e.logger.Printf("Directory export completed successfully: %s", output)
	return nil
}

// SiteReader 定义站点读取器接口
type SiteReader interface {
	// 检测给定目录是否为此类型的站点
	Detect(dir string) bool
	// 读取站点结构，返回按顺序排列的文件列表
	ReadStructure(dir string, configPath string) ([]string, error)
}

// GetMarkdownFilesInDir 获取目录中的所有 Markdown 文件并按文件名排序
func GetMarkdownFilesInDir(dir string) ([]string, error) {
	// 检查目录是否存在
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}

	// 递归查找所有 Markdown 文件
	var files []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".md" || ext == ".markdown" {
				files = append(files, path)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %s", dir, err)
	}

	// 按文件名排序
	sort.Strings(files)

	return files, nil
}
