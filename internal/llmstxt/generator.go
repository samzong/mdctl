package llmstxt

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"time"
)

// GeneratorConfig 包含生成llms.txt所需的配置
type GeneratorConfig struct {
	SitemapURL   string
	IncludePaths []string
	ExcludePaths []string
	FullMode     bool
	Concurrency  int
	Timeout      int
	UserAgent    string
	Verbose      bool
}

// PageInfo 存储页面的信息
type PageInfo struct {
	Title       string
	URL         string
	Description string
	Content     string // 页面正文内容，仅在全文模式下填充
	Section     string // 从URL路径提取的第一段作为章节
}

// Generator 是llms.txt生成器
type Generator struct {
	config GeneratorConfig
	logger *log.Logger
}

// NewGenerator 创建一个新的生成器实例
func NewGenerator(config GeneratorConfig) *Generator {
	var logger *log.Logger
	if config.Verbose {
		logger = log.New(os.Stdout, "[LLMSTXT] ", log.LstdFlags)
	} else {
		logger = log.New(io.Discard, "", 0)
	}

	return &Generator{
		config: config,
		logger: logger,
	}
}

// Generate 执行生成过程并返回生成的内容
func (g *Generator) Generate() (string, error) {
	startTime := time.Now()
	g.logger.Printf("Starting generation for sitemap: %s", g.config.SitemapURL)
	if g.config.FullMode {
		g.logger.Println("Full-content mode enabled")
	}

	// 1. 解析sitemap.xml获取URL列表
	urls, err := g.parseSitemap()
	if err != nil {
		return "", fmt.Errorf("failed to parse sitemap: %w", err)
	}
	g.logger.Printf("Found %d URLs in sitemap", len(urls))

	// 2. 过滤URL（基于include/exclude模式）
	urls = g.filterURLs(urls)
	g.logger.Printf("%d URLs after filtering", len(urls))

	// 3. 创建工作池并获取页面信息
	pages, err := g.fetchPages(urls)
	if err != nil {
		return "", fmt.Errorf("failed to fetch pages: %w", err)
	}

	// 4. 按章节分组页面信息
	sections := g.groupBySections(pages)

	// 5. 格式化为Markdown内容
	content := g.formatContent(sections)

	elapsedTime := time.Since(startTime).Round(time.Millisecond)
	g.logger.Printf("Generation completed successfully in %v", elapsedTime)
	return content, nil
}

// 按章节分组页面信息
func (g *Generator) groupBySections(pages []PageInfo) map[string][]PageInfo {
	sections := make(map[string][]PageInfo)

	for _, page := range pages {
		sections[page.Section] = append(sections[page.Section], page)
	}

	// 对每个章节内的页面按URL路径长度排序
	for section, sectionPages := range sections {
		sort.Slice(sectionPages, func(i, j int) bool {
			return len(sectionPages[i].URL) < len(sectionPages[j].URL)
		})
		sections[section] = sectionPages
	}

	return sections
}

// 获取排序后的章节名称列表，确保ROOT章节始终在最前面
func (g *Generator) getSortedSections(sections map[string][]PageInfo) []string {
	sectionNames := make([]string, 0, len(sections))

	// 首先添加ROOT章节（如果存在）
	if _, hasRoot := sections["ROOT"]; hasRoot {
		sectionNames = append(sectionNames, "ROOT")
	}

	// 添加其他章节并按字母排序
	for section := range sections {
		if section != "ROOT" {
			sectionNames = append(sectionNames, section)
		}
	}

	// 只有当有非ROOT章节时才进行排序
	if len(sectionNames) > 1 {
		// 只对非ROOT章节部分进行排序
		nonRootSections := sectionNames[1:]
		sort.Strings(nonRootSections)
	}

	return sectionNames
}
