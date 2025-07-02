# mdctl 的 llmstxt 功能设计文档

## 功能概述

为 mdctl 工具添加`llmstxt`子命令，用于将网站的`sitemap.xml`转换为`llms.txt`文件。`llms.txt`是一个包含网站页面列表的 Markdown 格式文件，专为训练或微调大语言模型(LLMs)准备内容而设计。

该功能支持两种模式：

1. **标准模式**：仅提取页面标题和描述信息，生成`llms.txt`
2. **全文模式**：除了标题和描述外，还提取页面正文内容，生成`llms-full.txt`

通过提供网站的 sitemap.xml URL，自动爬取页面内容，并生成结构化的 Markdown 文档，使其易于被用于 AI 模型的训练或微调。

## 用户需求

1. 支持从网站的 sitemap.xml 中获取所有 URL
2. 访问每个 URL 并提取页面标题和描述
3. 可选择提取页面正文内容（全文模式）
4. 按页面路径的第一段分组内容，形成章节结构
5. 生成格式化的 Markdown 文档
6. 支持包含/排除特定路径
7. 提供详细的处理日志和错误信息

## 命令设计

```
mdctl llmstxt [flags] <url>
```

### 参数设计

- `<url>`：sitemap.xml 的 URL (必填参数)
- `-i, --include-path`：包含的路径模式，支持 glob 语法 (可多次指定)
- `-e, --exclude-path`：排除的路径模式，支持 glob 语法 (可多次指定)
- `--ignore`：忽略的路径模式，支持 glob 语法 (可多次指定，与 exclude-path 功能相同)
- `-o, --output`：输出文件路径 (默认：标准输出)
- `-f, --full`：启用全文模式，提取页面正文内容 (默认：false)
- `-c, --concurrency`：并发请求数量 (默认：5)
- `--timeout`：请求超时时间，单位秒 (默认：30)
- `--verbose`：启用详细日志输出 (默认：false)

### 使用示例

```bash
# 基本用法（标准模式）
mdctl llmstxt https://example.com/sitemap.xml > llms.txt

# 全文模式
mdctl llmstxt -f https://example.com/sitemap.xml > llms-full.txt

# 指定输出文件
mdctl llmstxt https://example.com/sitemap.xml -o llms.txt

# 排除特定路径
mdctl llmstxt https://example.com/sitemap.xml -e "**/blog/**" -e "**/privacy**"

# 使用 ignore 模式过滤不需要的页面 (与 exclude-path 功能相同)
mdctl llmstxt https://example.com/sitemap.xml --ignore "**/admin/**" --ignore "**/private/**"

# 只包含特定路径
mdctl llmstxt https://example.com/sitemap.xml -i "**/docs/**"

# 调整并发请求数量
mdctl llmstxt https://example.com/sitemap.xml -c 10

# 启用详细日志
mdctl llmstxt https://example.com/sitemap.xml --verbose
```

## 实现设计

### 整体架构

按照 mdctl 的现有结构，在`cmd/`目录下创建`llmstxt.go`文件定义命令接口，在`internal/`目录下创建`llmstxt/`模块实现具体功能。

```
mdctl/
├── cmd/
│   └── llmstxt.go       # 新增：llmstxt命令定义
├── internal/
│   └── llmstxt/         # 新增：llmstxt功能实现
│       ├── generator.go  # 生成器核心实现
│       ├── sitemap.go    # Sitemap解析器
│       ├── fetcher.go    # 网页内容获取
│       ├── extractor.go  # 页面标题、描述和正文提取器
│       └── formatter.go  # Markdown格式生成
```

### 核心组件

#### 1. 命令处理器 (cmd/llmstxt.go)

负责解析命令行参数并调用 llmstxt 功能。

```go
var (
    includePaths     []string
    excludePaths     []string
    outputPath       string
    fullMode         bool
    concurrency      int
    timeout          int
    verbose          bool

    llmstxtCmd = &cobra.Command{
        Use:   "llmstxt [url]",
        Short: "Generate llms.txt from sitemap.xml",
        Long:  `Generate a llms.txt file from a website's sitemap.xml. This file is a curated list of the website's pages in markdown format, perfect for training or fine-tuning language models.

In standard mode, only title and description are extracted. In full mode (-f flag), the content of each page is also extracted.`,
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            sitemapURL := args[0]

            // 创建生成器并配置选项
            config := llmstxt.GeneratorConfig{
                SitemapURL:    sitemapURL,
                IncludePaths:  includePaths,
                ExcludePaths:  excludePaths,
                FullMode:      fullMode,
                Concurrency:   concurrency,
                Timeout:       timeout,
                UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
                Verbose:       verbose,
            }

            generator := llmstxt.NewGenerator(config)

            // 执行生成
            content, err := generator.Generate()
            if err != nil {
                return err
            }

            // 输出内容
            if outputPath == "" {
                // 输出到标准输出
                fmt.Println(content)
            } else {
                // 输出到文件
                return os.WriteFile(outputPath, []byte(content), 0644)
            }

            return nil
        },
    }
)

func init() {
    llmstxtCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: stdout)")
    llmstxtCmd.Flags().StringSliceVarP(&includePaths, "include-path", "i", []string{}, "Glob patterns for paths to include (can be specified multiple times)")
    llmstxtCmd.Flags().StringSliceVarP(&excludePaths, "exclude-path", "e", []string{}, "Glob patterns for paths to exclude (can be specified multiple times)")
    llmstxtCmd.Flags().BoolVarP(&fullMode, "full", "f", false, "Enable full-content mode (extract page content)")
    llmstxtCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 5, "Number of concurrent requests")
    llmstxtCmd.Flags().IntVar(&timeout, "timeout", 30, "Request timeout in seconds")
    llmstxtCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")

    // 将命令添加到核心命令组
    llmstxtCmd.GroupID = "core"
}
```

#### 2. 生成器 (internal/llmstxt/generator.go)

负责协调整个生成过程的核心组件。

```go
// GeneratorConfig 包含生成llms.txt所需的配置
type GeneratorConfig struct {
    SitemapURL    string
    IncludePaths  []string
    ExcludePaths  []string
    FullMode      bool
    Concurrency   int
    Timeout       int
    UserAgent     string
    Verbose       bool
}

// PageInfo 存储页面的信息
type PageInfo struct {
    Title       string
    URL         string
    Description string
    Content     string    // 页面正文内容，仅在全文模式下填充
    Section     string    // 从URL路径提取的第一段作为章节
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

    g.logger.Println("Generation completed successfully")
    return content, nil
}

// 其他私有方法实现具体功能...
```

#### 3. Sitemap 解析器 (internal/llmstxt/sitemap.go)

负责解析 sitemap.xml 文件，获取所有 URL。

```go
// 解析sitemap.xml文件并返回所有URL
func (g *Generator) parseSitemap() ([]string, error) {
    g.logger.Printf("Parsing sitemap from %s", g.config.SitemapURL)

    // 使用适当的HTTP客户端配置（超时、自定义UA等）
    client := &http.Client{
        Timeout: time.Duration(g.config.Timeout) * time.Second,
    }

    // 构建请求
    req, err := http.NewRequest("GET", g.config.SitemapURL, nil)
    if err != nil {
        return nil, err
    }

    // 设置User-Agent
    req.Header.Set("User-Agent", g.config.UserAgent)

    // 实现sitemap解析逻辑
    // 支持标准的sitemap.xml格式和sitemap索引格式
    // 使用XML解析库解析响应内容

    // 返回解析出的URL列表
}

// 基于include/exclude模式过滤URL列表
func (g *Generator) filterURLs(urls []string) []string {
    // 实现URL过滤逻辑
    // 使用glob匹配实现include/exclude过滤
    // 返回过滤后的URL列表
}
```

#### 4. 页面获取器 (internal/llmstxt/fetcher.go)

负责并发获取页面内容。

```go
// 使用工作池并发获取页面信息
func (g *Generator) fetchPages(urls []string) ([]PageInfo, error) {
    g.logger.Printf("Starting to fetch %d pages with concurrency %d", len(urls), g.config.Concurrency)

    // 实现工作池进行并发请求
    // 使用信号量或工作池模式控制并发数量
    // 向每个页面发送HTTP请求并提取内容

    // 返回所有页面信息
}
```

#### 5. 内容提取器 (internal/llmstxt/extractor.go)

负责从 HTML 页面中提取标题、描述和正文信息。

```go
// 从HTML内容中提取页面信息
func (g *Generator) extractPageInfo(url string, html string) (PageInfo, error) {
    // 使用HTML解析库（如goquery）提取页面标题和描述

    var pageInfo PageInfo
    pageInfo.URL = url
    pageInfo.Section = parseSection(url)

    // 提取标题和描述
    // ...

    // 在全文模式下提取页面正文
    if g.config.FullMode {
        // 提取主要内容，去除导航、页眉页脚、广告等
        // 可使用启发式算法或基于DOM结构分析识别主要内容区域
        // 保留Markdown兼容的格式
        // ...
    }

    return pageInfo, nil
}

// 从URL中解析章节信息
func parseSection(url string) string {
    // 从URL中提取路径部分第一段作为章节
    // ...
}
```

#### 6. Markdown 格式化器 (internal/llmstxt/formatter.go)

负责生成最终的 Markdown 内容。

```go
// 按章节分组页面信息
func (g *Generator) groupBySections(pages []PageInfo) map[string][]PageInfo {
    // 实现章节分组逻辑
    // 返回章节到页面列表的映射
}

// 格式化为Markdown内容
func (g *Generator) formatContent(sections map[string][]PageInfo) string {
    var buf strings.Builder

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
    for section, pages := range sections {
        // 跳过ROOT章节，因为它已经用于标题和描述
        if section == "ROOT" {
            continue
        }

        // 添加章节标题
        buf.WriteString("## ")
        buf.WriteString(capitalizeString(section))
        buf.WriteString("\n\n")

        // 添加章节内每个页面的信息
        for _, page := range pages {
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

    return strings.ToUpper(str[:1]) + strings.ToLower(str[1:])
}
```

### 工作流程

1. **命令解析**：解析用户提供的命令行参数
2. **Sitemap 解析**：获取并解析网站的 sitemap.xml，提取所有 URL
3. **URL 筛选**：基于包含/排除规则过滤 URL 列表
4. **内容获取**：并发访问每个 URL，获取页面内容
5. **信息提取**：
   - 从每个页面提取标题和描述
   - 在全文模式下，额外提取页面正文内容
6. **内容分组**：按 URL 的第一段路径对内容进行分组
7. **格式化输出**：生成格式化的 Markdown 文档
   - 标准模式：仅包含标题和描述
   - 全文模式：包含标题、描述和正文
8. **结果处理**：将结果写入文件或输出到标准输出

### 路径处理和分组策略

1. **路径解析**：

   - 从 URL 中提取路径部分
   - 按照"/"分割路径，获取第一段作为章节名称
   - 根路径("/")被视为特殊章节"ROOT"

2. **章节处理**：
   - 章节名称会被格式化为首字母大写（其余小写）
   - 章节按字母顺序排序，但"ROOT"章节始终在最前面
   - 章节内的页面按 URL 路径长度排序（较短的优先）

### 正文提取策略（全文模式）

在全文模式下，系统将使用以下策略提取页面正文内容：

1. **内容识别**：

   - 使用启发式算法识别主要内容区域
   - 识别并排除导航栏、侧边栏、页眉页脚、广告等非主要内容
   - 优先提取带有特定 HTML 标签（如 article、main、content）的内容区域

2. **清理和格式化**：

   - 移除 HTML 标签，保留文本内容
   - 保持段落、标题、列表等结构
   - 保留图片引用的 Markdown 格式
   - 删除 JavaScript 代码块和其他非文本内容
   - 清理多余空白和换行

3. **内容限制**：
   - 限制正文长度，避免过长内容（默认限制为 10,000 字符）
   - 对超长内容进行智能截断，保持完整段落

### 错误处理

1. **Sitemap 解析错误**：

   - 提供清晰的错误消息，指出 sitemap.xml 无法访问或格式无效
   - 支持常见的 sitemap 格式（XML、TXT、索引 sitemap 等）

2. **HTTP 请求错误**：

   - 页面请求失败时记录警告，但继续处理其他页面
   - 设置合理的超时，避免单个页面阻塞整个处理过程

3. **内容提取错误**：
   - 对于无法提取标题或描述的页面，使用 URL 或路径作为备选
   - 正文提取失败时，提供警告但不中断处理
   - 记录详细的警告信息，但不中断处理

### 性能优化

1. **并发请求**：

   - 使用工作池模式控制并发数量
   - 默认并发数为 5，可通过参数调整

2. **超时控制**：

   - 为每个 HTTP 请求设置超时，避免长时间等待
   - 默认超时为 30 秒，可通过参数调整

3. **内存优化**：
   - 流式处理大型 sitemap，避免一次性加载所有 URL
   - 使用 strings.Builder 高效构建输出内容
   - 在处理大型站点时分批处理 URL，避免内存占用过高

### 浏览器模拟

为了更好地获取页面内容，系统默认使用模拟 Chrome 浏览器的 User-Agent：

```
Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36
```

这有助于：

1. 避免被网站检测为爬虫而阻止
2. 确保获取到与普通用户相同的页面内容
3. 提高内容提取的准确性

## 测试策略

1. **单元测试**：

   - URL 过滤功能测试
   - Markdown 格式化测试
   - 章节分组逻辑测试
   - 正文提取算法测试

2. **集成测试**：

   - 使用本地测试服务器进行 sitemap 解析测试
   - 页面内容提取测试（标准模式和全文模式）
   - 完整生成流程测试

3. **边界情况测试**：
   - 空 sitemap 处理
   - 无效 URL 处理
   - 超大 sitemap 处理
   - 各种特殊字符和 Unicode 字符测试
   - JavaScript 生成的动态内容测试

## 后续扩展

1. **缓存支持**：添加本地缓存，避免重复请求相同页面
2. **更多输出格式**：支持输出为其他格式（如 CSV、JSON 等）
3. **自定义模板**：支持用户自定义输出模板
4. **网站内容深度分析**：提取更多信息，如关键词、主题等
5. **内容优化**：添加内容过滤和优化功能，如删除广告、压缩内容等
6. **支持更多站点结构**：除了 sitemap.xml，支持从其他源（如 robots.txt、网站地图页面等）获取 URL

## 使用场景

1. **AI 开发者**：收集网站内容用于训练或微调语言模型
2. **文档管理**：生成网站内容的结构化摘要
3. **内容审计**：快速获取网站的页面标题和描述列表
4. **SEO 分析**：分析网站的标题和描述标签质量
5. **内容存档**：将网站内容保存为本地 Markdown 文档
