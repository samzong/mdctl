package formats

import (
	"log"
)

// ExportOptions 定义导出选项
type ExportOptions struct {
	Template            string      // 模板文件路径
	GenerateToc         bool        // 是否生成目录
	ShiftHeadingLevelBy int         // 标题层级偏移量
	FileAsTitle         bool        // 是否使用文件名作为章节标题
	Verbose             bool        // 是否启用详细日志
	Logger              *log.Logger // 日志记录器
	SourceDirs          []string    // 源目录列表，用于处理图片路径
	TocDepth            int         // 目录深度，默认为 3
	NavPath             string      // 指定导航路径导出
}

// Formatter 定义格式化器接口
type Formatter interface {
	// Format 将输入文件转换为指定格式
	Format(input string, output string, options ExportOptions) error

	// ValidateOptions 验证导出选项是否有效
	ValidateOptions(options ExportOptions) error

	// GetFormatName 返回格式名称
	GetFormatName() string
}

// FormatFactory 创建格式化器的工厂函数类型
type FormatFactory func(logger *log.Logger) Formatter

// formatRegistry 存储已注册的格式化器工厂
var formatRegistry = make(map[string]FormatFactory)

// RegisterFormatter 注册格式化器工厂
func RegisterFormatter(format string, factory FormatFactory) {
	formatRegistry[format] = factory
}

// GetFormatter 获取指定格式的格式化器
func GetFormatter(format string, logger *log.Logger) (Formatter, error) {
	factory, ok := formatRegistry[format]
	if !ok {
		return nil, &UnsupportedFormatError{Format: format}
	}
	return factory(logger), nil
}

// UnsupportedFormatError 表示不支持的格式错误
type UnsupportedFormatError struct {
	Format string
}

// Error 实现 error 接口
func (e *UnsupportedFormatError) Error() string {
	return "unsupported format: " + e.Format
}
