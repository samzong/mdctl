package formats

import (
	"log"
)

// DocxFormatter 是DOCX格式的专用格式化器
type DocxFormatter struct {
	*PandocFormatter
}

// NewDocxFormatter 创建一个新的DOCX格式化器
func NewDocxFormatter(logger *log.Logger) Formatter {
	return &DocxFormatter{
		PandocFormatter: NewPandocFormatter("docx", logger),
	}
}

// ValidateOptions 验证DOCX导出选项是否有效
func (f *DocxFormatter) ValidateOptions(options ExportOptions) error {
	// 调用基类验证
	if err := f.PandocFormatter.ValidateOptions(options); err != nil {
		return err
	}

	// DOCX特定的验证逻辑可以在这里添加
	return nil
}

// addFormatSpecificArgs 添加DOCX格式特定的参数
func (f *DocxFormatter) addFormatSpecificArgs(args []string, options ExportOptions) []string {
	// DOCX格式不需要特殊参数，但如果将来需要，可以在这里添加
	return args
}

func init() {
	// 注册DOCX格式化器
	RegisterFormatter("docx", NewDocxFormatter)
}
