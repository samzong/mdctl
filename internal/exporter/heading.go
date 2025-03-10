package exporter

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

var (
	// 匹配 ATX 风格的标题（# 开头）
	atxHeadingRegex = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	// 匹配 Setext 风格的标题（下划线风格）
	setextHeading1Regex = regexp.MustCompile(`^=+\s*$`)
	setextHeading2Regex = regexp.MustCompile(`^-+\s*$`)
)

// ShiftHeadings 调整 Markdown 文本中的标题级别
func ShiftHeadings(content string, shiftBy int) string {
	if shiftBy == 0 {
		return content
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	var result []string
	var prevLine string
	var isPrevLineHeading bool

	for scanner.Scan() {
		line := scanner.Text()

		// 处理 ATX 风格的标题
		if matches := atxHeadingRegex.FindStringSubmatch(line); matches != nil {
			level := len(matches[1]) + shiftBy
			heading := matches[2]

			if level <= 6 {
				// 仍然是有效的标题级别
				result = append(result, fmt.Sprintf("%s %s", strings.Repeat("#", level), heading))
			} else {
				// 超过最大标题级别，转换为加粗文本
				result = append(result, fmt.Sprintf("**%s**", heading))
			}
			isPrevLineHeading = false
		} else if setextHeading1Regex.MatchString(line) && prevLine != "" {
			// 处理 Setext 风格的一级标题
			level := 1 + shiftBy
			if level <= 6 {
				result[len(result)-1] = fmt.Sprintf("%s %s", strings.Repeat("#", level), prevLine)
			} else {
				result[len(result)-1] = fmt.Sprintf("**%s**", prevLine)
			}
			isPrevLineHeading = true
		} else if setextHeading2Regex.MatchString(line) && prevLine != "" {
			// 处理 Setext 风格的二级标题
			level := 2 + shiftBy
			if level <= 6 {
				result[len(result)-1] = fmt.Sprintf("%s %s", strings.Repeat("#", level), prevLine)
			} else {
				result[len(result)-1] = fmt.Sprintf("**%s**", prevLine)
			}
			isPrevLineHeading = true
		} else {
			// 普通行
			result = append(result, line)
			isPrevLineHeading = false
		}

		if !isPrevLineHeading {
			prevLine = line
		}
	}

	return strings.Join(result, "\n")
}

// AddTitleFromFilename 根据文件名添加标题
func AddTitleFromFilename(content, filename string, level int) string {
	// 从文件名中提取标题（去除扩展名）
	title := strings.TrimSuffix(filename, ".md")
	title = strings.TrimSuffix(title, ".markdown")

	// 替换下划线和连字符为空格，使标题更易读
	title = strings.ReplaceAll(title, "_", " ")
	title = strings.ReplaceAll(title, "-", " ")

	// 标题首字母大写
	title = strings.Title(title)

	// 创建标题行
	var titleLine string
	if level <= 6 {
		titleLine = fmt.Sprintf("%s %s\n\n", strings.Repeat("#", level), title)
	} else {
		titleLine = fmt.Sprintf("**%s**\n\n", title)
	}

	return titleLine + content
}
