package markdownfmt

import (
	"regexp"
	"strings"
)

// Formatter 用于格式化 markdown 内容
type Formatter struct {
	// 是否启用格式化
	enabled bool
}

// New 创建一个新的格式化器
func New(enabled bool) *Formatter {
	return &Formatter{
		enabled: enabled,
	}
}

// Format 格式化 markdown 内容
func (f *Formatter) Format(content string) string {
	if !f.enabled {
		return content
	}

	// 1. 分割内容为行
	lines := strings.Split(content, "\n")

	// 2. 处理每一行
	var formatted []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// 处理标题：确保标题前后有空行
		if isHeading(line) {
			// 如果不是第一行，且前一行不是空行，添加空行
			if i > 0 && len(strings.TrimSpace(lines[i-1])) > 0 {
				formatted = append(formatted, "")
			}
			// 规范化标题格式（# 后面有一个空格）
			line = formatHeading(line)
			formatted = append(formatted, line)
			// 如果不是最后一行，添加空行
			if i < len(lines)-1 {
				formatted = append(formatted, "")
			}
			continue
		}

		// 处理中英文之间的空格
		line = formatChineseEnglishSpace(line)

		formatted = append(formatted, line)
	}

	// 3. 处理连续空行
	formatted = removeConsecutiveBlankLines(formatted)

	// 4. 合并行
	result := strings.Join(formatted, "\n")

	return result
}

// isHeading 检查是否是标题行
func isHeading(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#")
}

// formatHeading 格式化标题行
func formatHeading(line string) string {
	// 移除开头的空格
	line = strings.TrimSpace(line)
	// 确保 # 和文本之间只有一个空格
	re := regexp.MustCompile(`^(#+)\s*`)
	return re.ReplaceAllString(line, "$1 ")
}

// formatChineseEnglishSpace 在中英文之间添加空格
func formatChineseEnglishSpace(line string) string {
	// 匹配中文和英文/数字之间的边界
	re := regexp.MustCompile(`([\\p{Han}])([A-Za-z0-9])`)
	line = re.ReplaceAllString(line, "$1 $2")

	re = regexp.MustCompile(`([A-Za-z0-9])([\\p{Han}])`)
	line = re.ReplaceAllString(line, "$1 $2")

	return line
}

// removeConsecutiveBlankLines 移除连续的空行
func removeConsecutiveBlankLines(lines []string) []string {
	var result []string
	isPrevLineBlank := false

	for _, line := range lines {
		isCurrentLineBlank := len(strings.TrimSpace(line)) == 0

		if !isCurrentLineBlank || !isPrevLineBlank {
			result = append(result, line)
		}

		isPrevLineBlank = isCurrentLineBlank
	}

	return result
}
