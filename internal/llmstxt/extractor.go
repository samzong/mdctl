package llmstxt

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// 从HTML内容中提取页面信息
func (g *Generator) extractPageInfo(urlStr string, resp *http.Response) (PageInfo, error) {
	// 创建PageInfo对象
	pageInfo := PageInfo{
		URL:     urlStr,
		Section: parseSection(urlStr),
	}

	// 解析HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return pageInfo, err
	}

	// 提取标题
	pageInfo.Title = extractTitle(doc)
	if pageInfo.Title == "" {
		// 如果无法提取标题，使用URL的最后一段作为标题
		pageInfo.Title = extractTitleFromURL(urlStr)
	}

	// 提取描述
	pageInfo.Description = extractDescription(doc)

	// 在全文模式下提取正文内容
	if g.config.FullMode {
		pageInfo.Content = extractContent(doc)
	}

	return pageInfo, nil
}

// 从URL中提取章节信息
func parseSection(urlStr string) string {
	// 解析URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "ROOT"
	}

	// 分割路径
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")

	// 如果路径为空，返回ROOT
	if len(pathParts) == 0 || pathParts[0] == "" {
		return "ROOT"
	}

	// 返回第一段路径
	return pathParts[0]
}

// 从HTML文档中提取标题
func extractTitle(doc *goquery.Document) string {
	// 尝试从title标签提取
	title := doc.Find("title").First().Text()
	title = strings.TrimSpace(title)

	// 如果没有title标签，尝试从h1标签提取
	if title == "" {
		title = doc.Find("h1").First().Text()
		title = strings.TrimSpace(title)
	}

	return title
}

// 从URL中提取标题
func extractTitleFromURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	// 获取路径的最后一段
	basename := path.Base(parsedURL.Path)

	// 移除文件扩展名
	basename = strings.TrimSuffix(basename, path.Ext(basename))

	// 如果basename为空或者是"/"，使用主机名
	if basename == "" || basename == "." || basename == "/" {
		return parsedURL.Hostname()
	}

	// 替换连字符和下划线为空格，并首字母大写
	basename = strings.ReplaceAll(basename, "-", " ")
	basename = strings.ReplaceAll(basename, "_", " ")

	return strings.Title(basename)
}

// 从HTML文档中提取描述
func extractDescription(doc *goquery.Document) string {
	var description string

	// 尝试meta description
	description, _ = doc.Find("meta[name='description']").Attr("content")
	if description != "" {
		return strings.TrimSpace(description)
	}

	// 尝试og:description
	description, _ = doc.Find("meta[property='og:description']").Attr("content")
	if description != "" {
		return strings.TrimSpace(description)
	}

	// 尝试twitter:description
	description, _ = doc.Find("meta[name='twitter:description']").Attr("content")
	if description != "" {
		return strings.TrimSpace(description)
	}

	// 如果都没有找到，提取第一段文本
	description = doc.Find("p").First().Text()
	if description != "" {
		// 限制长度
		if len(description) > 200 {
			description = description[:197] + "..."
		}
		return strings.TrimSpace(description)
	}

	return "No description available"
}

// 从HTML文档中提取正文内容
func extractContent(doc *goquery.Document) string {
	var content strings.Builder

	// 尝试找到主要内容区域
	mainContent := doc.Find("article, main, #content, .content, .post-content").First()

	// 如果没有找到特定内容区域，使用body
	if mainContent.Length() == 0 {
		mainContent = doc.Find("body")
	}

	// 提取所有段落
	mainContent.Find("p, h1, h2, h3, h4, h5, h6, ul, ol, blockquote").Each(func(i int, s *goquery.Selection) {
		// 获取标签名
		tagName := goquery.NodeName(s)
		text := strings.TrimSpace(s.Text())

		if text == "" {
			return
		}

		// 根据标签类型格式化
		switch tagName {
		case "h1":
			content.WriteString("# " + text + "\n\n")
		case "h2":
			content.WriteString("## " + text + "\n\n")
		case "h3":
			content.WriteString("### " + text + "\n\n")
		case "h4":
			content.WriteString("#### " + text + "\n\n")
		case "h5":
			content.WriteString("##### " + text + "\n\n")
		case "h6":
			content.WriteString("###### " + text + "\n\n")
		case "p":
			content.WriteString(text + "\n\n")
		case "blockquote":
			content.WriteString("> " + text + "\n\n")
		case "ul", "ol":
			s.Find("li").Each(func(j int, li *goquery.Selection) {
				liText := strings.TrimSpace(li.Text())
				if liText != "" {
					if tagName == "ul" {
						content.WriteString("- " + liText + "\n")
					} else {
						content.WriteString(fmt.Sprintf("%d. %s\n", j+1, liText))
					}
				}
			})
			content.WriteString("\n")
		}
	})

	// 限制内容长度
	contentStr := content.String()
	if len(contentStr) > 10000 {
		// 找到最后一个段落结束位置
		lastParaEnd := strings.LastIndex(contentStr[:10000], "\n\n")
		if lastParaEnd == -1 {
			lastParaEnd = 10000
		}
		contentStr = contentStr[:lastParaEnd] + "\n\n... (content truncated)"
	}

	return contentStr
}
