package llmstxt

import (
	"strings"
	"unicode"
)

// 格式化为Markdown内容
func (g *Generator) formatContent(sections map[string][]PageInfo) string {
	var buf strings.Builder

	// 获取排序后的章节列表
	sectionNames := g.getSortedSections(sections)

	// 找到根页面信息
	var rootPage PageInfo
	if rootPages, ok := sections["ROOT"]; ok && len(rootPages) > 0 {
		rootPage = rootPages[0]
	}

	// 添加文档标题
	buf.WriteString("# ")
	buf.WriteString(rootPage.Title)
	buf.WriteString("\n\n")

	// 添加文档描述
	buf.WriteString("> ")
	buf.WriteString(rootPage.Description)
	buf.WriteString("\n\n")

	// 处理每个章节
	for _, section := range sectionNames {
		// 跳过ROOT章节，因为它已经用于标题和描述
		if section == "ROOT" {
			continue
		}

		// 添加章节标题
		buf.WriteString("## ")
		buf.WriteString(capitalizeString(section))
		buf.WriteString("\n\n")

		// 添加章节内每个页面的信息
		for _, page := range sections[section] {
			buf.WriteString("- [")
			buf.WriteString(page.Title)
			buf.WriteString("](")
			buf.WriteString(page.URL)
			buf.WriteString("): ")
			buf.WriteString(page.Description)
			buf.WriteString("\n")

			// 在全文模式下添加页面正文内容
			if g.config.FullMode && page.Content != "" {
				buf.WriteString("\n")
				buf.WriteString(page.Content)
				buf.WriteString("\n")
			}

			buf.WriteString("\n")
		}
	}

	return buf.String()
}

// 首字母大写，其余小写
func capitalizeString(str string) string {
	if str == "" {
		return ""
	}

	runes := []rune(str)
	return string(unicode.ToUpper(runes[0])) + strings.ToLower(string(runes[1:]))
}
