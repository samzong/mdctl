package llmstxt

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gobwas/glob"
)

// Sitemap XML结构
type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []struct {
		Loc        string `xml:"loc"`
		LastMod    string `xml:"lastmod,omitempty"`
		ChangeFreq string `xml:"changefreq,omitempty"`
		Priority   string `xml:"priority,omitempty"`
	} `xml:"url"`
}

// SitemapIndex XML结构
type SitemapIndex struct {
	XMLName  xml.Name `xml:"sitemapindex"`
	Sitemaps []struct {
		Loc     string `xml:"loc"`
		LastMod string `xml:"lastmod,omitempty"`
	} `xml:"sitemap"`
}

// 解析sitemap.xml文件并返回所有URL
func (g *Generator) parseSitemap() ([]string, error) {
	g.logger.Printf("Parsing sitemap from %s", g.config.SitemapURL)

	// 设置HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(g.config.Timeout) * time.Second,
	}

	// 构建请求
	req, err := http.NewRequest("GET", g.config.SitemapURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", g.config.UserAgent)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sitemap: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch sitemap, status code: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sitemap content: %w", err)
	}

	// 尝试解析为标准sitemap
	var sitemap Sitemap
	if err := xml.Unmarshal(body, &sitemap); err == nil && len(sitemap.URLs) > 0 {
		g.logger.Println("Parsed standard sitemap")
		return g.extractURLsFromSitemap(sitemap), nil
	}

	// 尝试解析为sitemap索引
	var sitemapIndex SitemapIndex
	if err := xml.Unmarshal(body, &sitemapIndex); err == nil && len(sitemapIndex.Sitemaps) > 0 {
		g.logger.Println("Parsed sitemap index, fetching child sitemaps")
		return g.fetchSitemapIndex(sitemapIndex, client)
	}

	// 如果都解析失败，尝试按文本格式处理（一行一个URL）
	lines := string(body)
	if len(lines) > 0 {
		g.logger.Println("Parsing as text sitemap")
		return g.parseTextSitemap(lines), nil
	}

	return nil, fmt.Errorf("could not parse sitemap, unknown format")
}

// 从标准sitemap提取URL
func (g *Generator) extractURLsFromSitemap(sitemap Sitemap) []string {
	urls := make([]string, 0, len(sitemap.URLs))
	for _, urlEntry := range sitemap.URLs {
		if urlEntry.Loc != "" {
			urls = append(urls, urlEntry.Loc)
		}
	}
	return urls
}

// 从sitemap索引获取所有子sitemap的URL
func (g *Generator) fetchSitemapIndex(index SitemapIndex, client *http.Client) ([]string, error) {
	var allURLs []string

	for _, sitemapEntry := range index.Sitemaps {
		if sitemapEntry.Loc == "" {
			continue
		}

		g.logger.Printf("Fetching child sitemap: %s", sitemapEntry.Loc)

		// 构建请求
		req, err := http.NewRequest("GET", sitemapEntry.Loc, nil)
		if err != nil {
			g.logger.Printf("Warning: failed to create request for child sitemap %s: %v", sitemapEntry.Loc, err)
			continue
		}

		// 设置User-Agent
		req.Header.Set("User-Agent", g.config.UserAgent)

		// 发送请求
		resp, err := client.Do(req)
		if err != nil {
			g.logger.Printf("Warning: failed to fetch child sitemap %s: %v", sitemapEntry.Loc, err)
			continue
		}

		// 读取响应体
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			g.logger.Printf("Warning: failed to read child sitemap %s: %v", sitemapEntry.Loc, err)
			continue
		}

		// 解析子sitemap
		var childSitemap Sitemap
		if err := xml.Unmarshal(body, &childSitemap); err != nil {
			g.logger.Printf("Warning: failed to parse child sitemap %s: %v", sitemapEntry.Loc, err)
			continue
		}

		// 提取URL
		childURLs := g.extractURLsFromSitemap(childSitemap)
		g.logger.Printf("Found %d URLs in child sitemap %s", len(childURLs), sitemapEntry.Loc)
		allURLs = append(allURLs, childURLs...)
	}

	return allURLs, nil
}

// 解析文本格式的sitemap（一行一个URL）
func (g *Generator) parseTextSitemap(content string) []string {
	lines := splitLines(content)
	var urls []string

	for _, line := range lines {
		line = normalizeURL(line)
		if isValidURL(line) {
			urls = append(urls, line)
		}
	}

	return urls
}

// 基于include/exclude模式过滤URL列表
func (g *Generator) filterURLs(urls []string) []string {
	if len(g.config.IncludePaths) == 0 && len(g.config.ExcludePaths) == 0 {
		return urls // 没有过滤规则，直接返回
	}

	// 编译include/exclude模式
	var includeMatchers, excludeMatchers []glob.Glob
	for _, pattern := range g.config.IncludePaths {
		matcher, err := glob.Compile(pattern)
		if err != nil {
			g.logger.Printf("Warning: invalid include pattern '%s': %v", pattern, err)
			continue
		}
		includeMatchers = append(includeMatchers, matcher)
	}

	for _, pattern := range g.config.ExcludePaths {
		matcher, err := glob.Compile(pattern)
		if err != nil {
			g.logger.Printf("Warning: invalid exclude pattern '%s': %v", pattern, err)
			continue
		}
		excludeMatchers = append(excludeMatchers, matcher)
	}

	var filteredURLs []string
	for _, url := range urls {
		// 如果有include规则，必须匹配其中一个
		if len(includeMatchers) > 0 {
			matched := false
			for _, matcher := range includeMatchers {
				if matcher.Match(url) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// 如果匹配任何exclude规则，则排除
		excluded := false
		for _, matcher := range excludeMatchers {
			if matcher.Match(url) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		filteredURLs = append(filteredURLs, url)
	}

	return filteredURLs
}

// 辅助函数：按行分割文本
func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

// 辅助函数：标准化URL（去除空格等）
func normalizeURL(url string) string {
	return url
}

// 辅助函数：检查URL是否有效
func isValidURL(url string) bool {
	return url != ""
}
