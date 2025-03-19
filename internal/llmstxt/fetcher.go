package llmstxt

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// 使用工作池并发获取页面信息
func (g *Generator) fetchPages(urls []string) ([]PageInfo, error) {
	g.logger.Printf("Starting to fetch %d pages with concurrency %d", len(urls), g.config.Concurrency)

	// 创建结果通道和错误通道
	resultChan := make(chan PageInfo, len(urls))
	errorChan := make(chan error, len(urls))

	// 创建工作通道，控制并发数量
	workChan := make(chan string, len(urls))

	// 启动工作池
	var wg sync.WaitGroup
	for i := 0; i < g.config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for urlStr := range workChan {
				pageInfo, err := g.fetchPageContent(urlStr)
				if err != nil {
					g.logger.Printf("Warning: failed to fetch page %s: %v", urlStr, err)
					errorChan <- fmt.Errorf("failed to fetch page %s: %w", urlStr, err)
					continue
				}
				resultChan <- pageInfo
			}
		}()
	}

	// 发送所有URL到工作通道
	for _, urlStr := range urls {
		workChan <- urlStr
	}
	close(workChan)

	// 等待所有工作完成
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// 收集结果
	var results []PageInfo
	for result := range resultChan {
		results = append(results, result)
		g.logger.Printf("Fetched page: %s", result.URL)
	}

	// 检查错误（不中断处理，只记录警告）
	for err := range errorChan {
		g.logger.Printf("Warning: %v", err)
	}

	g.logger.Printf("Successfully fetched %d/%d pages", len(results), len(urls))

	return results, nil
}

// 获取单个页面的内容
func (g *Generator) fetchPageContent(urlStr string) (PageInfo, error) {
	// 设置HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(g.config.Timeout) * time.Second,
	}

	// 构建请求
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return PageInfo{}, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", g.config.UserAgent)

	// 发送请求
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return PageInfo{}, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PageInfo{}, fmt.Errorf("failed to fetch page, status code: %d", resp.StatusCode)
	}

	// 提取页面信息
	pageInfo, err := g.extractPageInfo(urlStr, resp)
	if err != nil {
		return PageInfo{}, fmt.Errorf("failed to extract page info: %w", err)
	}

	// 记录耗时信息
	elapsed := time.Since(start).Round(time.Millisecond)
	g.logger.Printf("Fetched %s in %v", urlStr, elapsed)

	return pageInfo, nil
}
