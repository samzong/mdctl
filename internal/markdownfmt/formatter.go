package markdownfmt

import (
	"fmt"
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

		// 处理链接中的空格
		line = formatMarkdownLinks(line)

		// 处理括号内容
		line = formatParentheses(line)

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

// formatParentheses 处理括号内的格式
func formatParentheses(line string) string {
	// 先处理 http/https 链接，将它们临时替换掉
	linkPattern := regexp.MustCompile(`\([^)]*https?://[^)]+\)`)
	links := linkPattern.FindAllString(line, -1)
	for i, link := range links {
		line = strings.Replace(line, link, fmt.Sprintf("__LINK_PLACEHOLDER_%d__", i), 1)
	}

	// 处理普通括号内容
	re := regexp.MustCompile(`\(([^)]+)\)`)
	line = re.ReplaceAllStringFunc(line, func(match string) string {
		// 提取括号内的内容
		content := match[1 : len(match)-1]
		// 清理首尾空格
		content = strings.TrimSpace(content)
		// 将连续的空格替换为单个空格
		content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
		return fmt.Sprintf("(%s)", content)
	})

	// 恢复链接
	for i, link := range links {
		line = strings.Replace(line, fmt.Sprintf("__LINK_PLACEHOLDER_%d__", i), link, 1)
	}

	return line
}

// formatMarkdownLinks 处理 markdown 链接中的空格
func formatMarkdownLinks(line string) string {
	// 匹配 markdown 链接格式 [text](url)，包括可能的空格
	linkPattern := regexp.MustCompile(`\[(.*?)\]\(\s*(.*?)\s*\)`)

	// 处理链接文本和 URL 中的空格
	line = linkPattern.ReplaceAllStringFunc(line, func(match string) string {
		// 提取链接文本和 URL
		parts := linkPattern.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}

		text := parts[1]
		url := parts[2]

		// 清理 URL 中的空格
		url = strings.TrimSpace(url)
		// 移除 URL 中的所有空格和不可见字符
		url = regexp.MustCompile(`[\s\p{Zs}\p{C}]+`).ReplaceAllString(url, "")

		// 保持链接文本中的空格，但清理首尾空格和连续空格
		text = strings.TrimSpace(text)
		text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

		// 重新组装链接
		return fmt.Sprintf("[%s](%s)", text, url)
	})

	// 处理标题链接中的空格
	headingLinkPattern := regexp.MustCompile(`\]\(#(.*?)\)`)
	line = headingLinkPattern.ReplaceAllStringFunc(line, func(match string) string {
		parts := headingLinkPattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}

		anchor := parts[1]
		// 移除所有空格
		anchor = regexp.MustCompile(`\s+`).ReplaceAllString(anchor, "")
		return fmt.Sprintf("](#%s)", anchor)
	})

	return line
}

// formatChineseEnglishSpace 在中英文之间添加空格
func formatChineseEnglishSpace(line string) string {
	// 匹配中文和英文/数字之间的边界
	re := regexp.MustCompile(`([\p{Han}])([A-Za-z0-9])`)
	line = re.ReplaceAllString(line, "$1 $2")

	re = regexp.MustCompile(`([A-Za-z0-9])([\p{Han}])`)
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
