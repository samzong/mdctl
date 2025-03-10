# Export 功能设计文档

## 功能概述

为 mdctl 工具增加 `export` 子命令，用于将 Markdown 文件导出为其他格式。第一版将优先支持导出为 Word 文档格式（docx），后续可扩展支持更多格式（如 PDF、EPUB 等）。

该功能将利用 Pandoc 作为底层导出工具，支持 Pandoc 的模板系统，允许用户配置自定义的导出模板。

## 用户需求

1. 支持将单个 Markdown 文件导出为 Word 格式
2. 支持将多个 Markdown 文件合并后导出为单个 Word 文档
3. 支持按照文件夹中的文件名顺序合并文件
4. 支持多种文档系统（MkDocs 第一期、Hugo、Docusaurus coming soon）的文件读取方式
5. 在合并过程中智能调整标题层级，保持文档结构的清晰性
6. 支持自定义 Word 模板，使最终文档具有一致的样式

## 命令设计

```
mdctl export [flags]
```

### 参数设计

- `-f, --file`: 指定单个 Markdown 文件进行导出
- `-d, --dir`: 指定包含多个 Markdown 文件的目录
- `-s, --site-type`: 指定文档站点类型，可选值：mkdocs, hugo, docusaurus（默认：mkdocs）
- `-o, --output`: 指定输出文件路径
- `-t, --template`: 指定 Word 模板文件路径
- `-F, --format`: 指定输出格式，可选值：docx, pdf, epub（默认：docx）
- `--toc`: 是否生成目录（默认：false）
- `--shift-heading-level-by`: 标题层级偏移量（默认：0）
- `--file-as-title`: 是否使用文件名作为章节标题（默认：false）

### 使用示例

```bash
# 导出单个文件
mdctl export -f README.md -o output.docx

# 导出整个目录
mdctl export -d docs/ -o documentation.docx

# 导出 MkDocs 站点
mdctl export -d docs/ -s mkdocs -o site_docs.docx

# 导出 Hugo 站点
mdctl export -d content/ -s hugo -o hugo_docs.docx

# 使用自定义模板
mdctl export -d docs/ -o report.docx -t templates/corporate.docx

# 指定标题层级偏移量
mdctl export -d docs/ -o documentation.docx --shift-heading-level-by 2

# 导出为 PDF 格式
mdctl export -d docs/ -o documentation.pdf -F pdf
```

## 实现设计

### 整体架构

按照项目的现有结构，我们将在 `cmd/` 目录下创建 `export.go` 文件定义命令接口，在 `internal/` 目录下创建 `exporter/` 模块实现具体功能。

```
mdctl/
├── cmd/
│   └── export.go        # 新增：export 命令定义
├── internal/
│   └── exporter/        # 新增：导出功能实现
│       ├── exporter.go  # 导出器接口定义
│       ├── pandoc.go    # Pandoc 导出实现
│       ├── merger.go    # Markdown 合并实现
│       ├── sitereader/  # 新增：不同文档系统的站点结构读取
│       │   ├── reader.go    # 站点读取器接口
│       │   ├── mkdocs.go    # MkDocs 站点读取
│       │   ├── hugo.go      # Hugo 站点读取
│       │   └── docusaurus.go # Docusaurus 站点读取
│       └── heading.go   # 标题处理实现
```

### 核心组件

#### 1. 命令处理器 (cmd/export.go)

负责解析命令行参数并调用导出功能。

```go
var (
    exportFile         string
    exportDir          string
    siteType           string
    configFile         string
    exportOutput       string
    exportTemplate     string
    exportFormat       string
    pandocPath         string
    generateToc        bool
    shiftHeadingLevelBy int
    fileAsTitle        bool

    exportCmd = &cobra.Command{
        Use:   "export",
        Short: "Export markdown files to other formats",
        Long:  `...`,
        RunE: func(cmd *cobra.Command, args []string) error {
            // 参数验证和处理逻辑
            // 调用 internal/exporter 的功能
        },
    }
)
```

#### 2. 导出器接口 (internal/exporter/exporter.go)

定义导出功能的通用接口，支持扩展其他格式。

```go
type Exporter interface {
    Export(input string, output string, options ExportOptions) error
}

type ExportOptions struct {
    Template            string
    GenerateToc         bool
    ShiftHeadingLevelBy int
    FileAsTitle         bool
    Format              string
    // 其他选项
}
```

#### 3. Pandoc 导出实现 (internal/exporter/pandoc.go)

使用 Pandoc 工具实现导出功能。

```go
type PandocExporter struct {
    PandocPath string
}

func (e *PandocExporter) Export(input, output string, options ExportOptions) error {
    // 构建并执行 Pandoc 命令
    // 如果 pandoc 不可用，返回明确的错误提示
}
```

#### 4. 站点结构读取器 (internal/exporter/sitereader/)

负责识别和解析不同文档系统的站点结构。

```go
// 站点读取器接口
type SiteReader interface {
    // 检测给定目录是否为此类型的站点
    Detect(dir string) bool
    
    // 读取站点结构，返回按顺序排列的文件列表
    ReadStructure(dir string, configPath string) ([]string, error)
}

// 工厂函数，根据站点类型返回相应的读取器
func GetSiteReader(siteType string) (SiteReader, error) {
    // 返回对应类型的读取器实现
}
```

#### 5. Markdown 合并器 (internal/exporter/merger.go)

负责合并多个 Markdown 文件。

```go
type Merger struct {
    ShiftHeadingLevelBy int
    FileAsTitle         bool
}

func (m *Merger) Merge(sources []string, target string) error {
    // 合并多个 Markdown 文件的逻辑
    // 自动处理标题层级
}
```

#### 6. 标题处理器 (internal/exporter/heading.go)

处理 Markdown 文件中的标题层级。

```go
func ShiftHeadings(content string, levels int) string {
    // 调整标题层级的逻辑
}
```

### 工作流程

1. **命令解析**：解析用户提供的命令行参数
2. **文件收集**：根据参数收集需要处理的 Markdown 文件
   - 单文件模式：直接使用指定文件
   - 目录模式：收集目录中的所有 Markdown 文件并按文件名排序
   - 站点模式：使用相应的站点读取器解析站点结构
3. **文件合并**：如果有多个文件，将它们合并为一个临时 Markdown 文件
   - 自动调整每个文件的标题层级
   - 可选添加文件名作为章节标题
4. **格式转换**：使用 Pandoc 将 Markdown 转换为目标格式
   - 应用用户指定的模板（如果有）
   - 生成目录（如果启用）
5. **输出处理**：将最终结果输出到用户指定的路径

## 标题层级处理策略

为了解决多文件合并时标题层级的问题，系统将自动处理标题层级：

1. 每个文件的标题层级将按照指定的偏移量调整：
   - H1 -> H(1+偏移量)
   - H2 -> H(2+偏移量)
   - ...
   - 如果调整后超过 H6，将转换为加粗文本 (**文本**)

2. 如果启用了文件名作为标题功能，会自动在每个文件内容前添加对应层级的标题

3. 系统会自动处理标题的相对层级关系，确保文档结构的逻辑性

## 依赖条件

**Pandoc**：需要系统中安装 Pandoc 工具
- 在执行导出命令时检查 Pandoc 是否可用
- 如果找不到 Pandoc，提供明确的错误信息和安装指导

## 错误处理

1. Pandoc 不可用时提供明确的错误信息和安装指导
2. 文件不存在或无法访问时的错误处理
3. 合并过程中可能出现的格式问题处理
4. 模板文件异常的处理
5. 不支持的站点类型或配置文件处理

## 未来扩展

1. 增强模板管理功能，支持模板下载和更新
2. 支持更多的文档站点系统
3. 支持更复杂的文档结构处理，如自动生成封面、页眉页脚
4. 集成图表和公式渲染功能