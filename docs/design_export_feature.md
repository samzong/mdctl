# Export 功能设计文档

## TODO 列表

以下是尚未完成的功能：

1. **站点读取器**：
   - [ ] 实现 Hugo 站点读取器
   - [ ] 实现 Docusaurus 站点读取器

2. **导出格式优化**：
   - [ ] 优化 PDF 导出格式
   - [ ] 优化 EPUB 导出格式

3. **高级功能**：
   - [ ] 实现自动生成封面功能
   - [ ] 实现自定义页眉页脚
   - [ ] 增强图表和公式渲染支持

## 功能概述

为 mdctl 工具增加 `export` 子命令，用于将 Markdown 文件导出为其他格式。第一版将优先支持导出为 Word 文档格式（docx），后续可扩展支持更多格式（如 PDF、EPUB 等）。

该功能将利用 Pandoc 作为底层导出工具，支持 Pandoc 的模板系统，允许用户配置自定义的导出模板。

## 用户需求

1. 支持将单个 Markdown 文件导出为 Word 格式 ✅
2. 支持将多个 Markdown 文件合并后导出为单个 Word 文档 ✅
3. 支持按照文件夹中的文件名顺序合并文件 ✅
4. 支持多种文档系统（MkDocs ✅、Hugo ❌、Docusaurus ❌）的文件读取方式
5. 在合并过程中智能调整标题层级，保持文档结构的清晰性 ✅
6. 支持自定义 Word 模板，使最终文档具有一致的样式 ✅

## 命令设计

```
mdctl export [flags]
```

### 参数设计

- `-f, --file`: 指定单个 Markdown 文件进行导出 ✅
- `-d, --dir`: 指定包含多个 Markdown 文件的目录 ✅
- `-s, --site-type`: 指定文档站点类型，可选值：mkdocs, hugo, docusaurus（默认：mkdocs）✅
- `-o, --output`: 指定输出文件路径 ✅
- `-t, --template`: 指定 Word 模板文件路径 ✅
- `-F, --format`: 指定输出格式，可选值：docx, pdf, epub（默认：docx）✅
- `--toc`: 是否生成目录（默认：false）✅
- `--shift-heading-level-by`: 标题层级偏移量（默认：0）✅
- `--file-as-title`: 是否使用文件名作为章节标题（默认：false）✅
- `--verbose, -v`: 启用详细日志记录（默认：false）✅ [新增]
- `--toc-depth`: 目录深度（默认：3）✅ [新增]
- `--nav-path`: 指定导航路径导出（例如：'Section1/Subsection2'）✅ [新增]

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

# 生成目录并指定目录深度
mdctl export -d docs/ -o documentation.docx --toc --toc-depth 4

# 导出特定导航路径
mdctl export -d docs/ -s mkdocs -o section_docs.docx --nav-path "Section1/Subsection2"

# 启用详细日志
mdctl export -d docs/ -o documentation.docx -v
```

## 实现设计

### 整体架构

按照项目的现有结构，我们在 `cmd/` 目录下创建了 `export.go` 文件定义命令接口，在 `internal/` 目录下创建了 `exporter/` 模块实现具体功能。

```
mdctl/
├── cmd/
│   └── export.go        # ✅ 已实现：export 命令定义
├── internal/
│   └── exporter/        # ✅ 已实现：导出功能实现
│       ├── exporter.go  # ✅ 已实现：导出器接口定义
│       ├── pandoc.go    # ✅ 已实现：Pandoc 导出实现
│       ├── merger.go    # ✅ 已实现：Markdown 合并实现
│       ├── sitereader/  # 部分实现：不同文档系统的站点结构读取
│       │   ├── reader.go    # ✅ 已实现：站点读取器接口
│       │   ├── mkdocs.go    # ✅ 已实现：MkDocs 站点读取
│       │   ├── hugo.go      # ❌ 未实现：Hugo 站点读取
│       │   └── docusaurus.go # ❌ 未实现：Docusaurus 站点读取
│       └── heading.go   # ✅ 已实现：标题处理实现
```

### 核心组件

#### 1. 命令处理器 (cmd/export.go)

负责解析命令行参数并调用导出功能。✅ 已实现

```go
var (
    exportFile          string
    exportDir           string
    siteType            string
    exportOutput        string
    exportTemplate      string
    exportFormat        string
    generateToc         bool
    shiftHeadingLevelBy int
    fileAsTitle         bool
    verbose             bool
    tocDepth            int
    navPath             string
    logger              *log.Logger

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

定义导出功能的通用接口，支持扩展其他格式。✅ 已实现

```go
type ExportOptions struct {
    Template            string      // Word 模板文件路径
    GenerateToc         bool        // 是否生成目录
    ShiftHeadingLevelBy int         // 标题层级偏移量
    FileAsTitle         bool        // 是否使用文件名作为章节标题
    Format              string      // 输出格式 (docx, pdf, epub)
    SiteType            string      // 站点类型 (mkdocs, hugo, docusaurus)
    Verbose             bool        // 是否启用详细日志
    Logger              *log.Logger // 日志记录器
    SourceDirs          []string    // 源目录列表，用于处理图片路径
    TocDepth            int         // 目录深度，默认为 3
    NavPath             string      // 指定导航路径导出
}

// Exporter 定义导出器接口
type Exporter interface {
    Export(input string, output string, options ExportOptions) error
}

// DefaultExporter 是默认的导出器实现
type DefaultExporter struct {
    pandocPath string
    logger     *log.Logger
}
```

#### 3. Pandoc 导出实现 (internal/exporter/pandoc.go)

使用 Pandoc 工具实现导出功能。✅ 已实现

```go
type PandocExporter struct {
    PandocPath string
    Logger     *log.Logger
}

func (e *PandocExporter) Export(input, output string, options ExportOptions) error {
    // 构建并执行 Pandoc 命令
    // 如果 pandoc 不可用，返回明确的错误提示
}

// CheckPandocAvailability 检查 Pandoc 是否可用
func CheckPandocAvailability() error {
    // 检查 Pandoc 是否已安装
}
```

#### 4. 站点结构读取器 (internal/exporter/sitereader/)

负责识别和解析不同文档系统的站点结构。✅ 部分实现（仅 MkDocs）

```go
// 站点读取器接口
type SiteReader interface {
    // 检测给定目录是否为此类型的站点
    Detect(dir string) bool
    
    // 读取站点结构，返回按顺序排列的文件列表
    ReadStructure(dir string, configPath string, navPath string) ([]string, error)
}

// GetSiteReader 根据站点类型返回相应的读取器
func GetSiteReader(siteType string, verbose bool, logger *log.Logger) (SiteReader, error) {
    switch siteType {
    case "mkdocs":
        return &MkDocsReader{Logger: logger}, nil
    case "hugo":
        return nil, fmt.Errorf("hugo site type is not yet implemented")
    case "docusaurus":
        return nil, fmt.Errorf("docusaurus site type is not yet implemented")
    default:
        return nil, fmt.Errorf("unsupported site type: %s", siteType)
    }
}
```

#### 5. Markdown 合并器 (internal/exporter/merger.go)

负责合并多个 Markdown 文件。✅ 已实现

```go
type Merger struct {
    ShiftHeadingLevelBy int
    FileAsTitle         bool
    Logger              *log.Logger
    SourceDirs          []string
    Verbose             bool
}

func (m *Merger) Merge(sources []string, target string) error {
    // 合并多个 Markdown 文件的逻辑
    // 自动处理标题层级
    // 处理图片路径
    // 处理 YAML Front Matter
}
```

#### 6. 标题处理器 (internal/exporter/heading.go)

处理 Markdown 文件中的标题层级。✅ 已实现

```go
// ShiftHeadings 调整 Markdown 文本中的标题级别
func ShiftHeadings(content string, shiftBy int) string {
    // 调整标题层级的逻辑
}

// AddTitleFromFilename 根据文件名添加标题
func AddTitleFromFilename(content, filename string, level int) string {
    // 从文件名生成标题
}
```

### 工作流程

1. **命令解析**：解析用户提供的命令行参数 ✅
2. **文件收集**：根据参数收集需要处理的 Markdown 文件 ✅
   - 单文件模式：直接使用指定文件 ✅
   - 目录模式：收集目录中的所有 Markdown 文件并按文件名排序 ✅
   - 站点模式：使用相应的站点读取器解析站点结构 ✅ (仅 MkDocs)
3. **文件合并**：如果有多个文件，将它们合并为一个临时 Markdown 文件 ✅
   - 自动调整每个文件的标题层级 ✅
   - 可选添加文件名作为章节标题 ✅
4. **格式转换**：使用 Pandoc 将 Markdown 转换为目标格式 ✅
   - 应用用户指定的模板（如果有）✅
   - 生成目录（如果启用）✅
5. **输出处理**：将最终结果输出到用户指定的路径 ✅

## 标题层级处理策略

为了解决多文件合并时标题层级的问题，系统自动处理标题层级：✅ 已实现

1. 每个文件的标题层级按照指定的偏移量调整：
   - H1 -> H(1+偏移量)
   - H2 -> H(2+偏移量)
   - ...
   - 如果调整后超过 H6，将转换为加粗文本 (**文本**)

2. 如果启用了文件名作为标题功能，会自动在每个文件内容前添加对应层级的标题

3. 系统会自动处理标题的相对层级关系，确保文档结构的逻辑性

## 依赖条件

**Pandoc**：需要系统中安装 Pandoc 工具 ✅ 已实现检查
- 在执行导出命令时检查 Pandoc 是否可用
- 如果找不到 Pandoc，提供明确的错误信息和安装指导

## 错误处理

1. Pandoc 不可用时提供明确的错误信息和安装指导 ✅
2. 文件不存在或无法访问时的错误处理 ✅
3. 合并过程中可能出现的格式问题处理 ✅
4. 模板文件异常的处理 ✅
5. 不支持的站点类型或配置文件处理 ✅

## 未来扩展

1. 支持更多的文档站点系统 ❌ (仅实现了 MkDocs)
2. 支持更复杂的文档结构处理，如自动生成封面、页眉页脚 ❌
3. 集成图表和公式渲染功能 ❌

## 当前实现的额外功能

1. **详细日志记录**：添加了 `--verbose` 选项，提供详细的导出过程日志 ✅
2. **目录深度控制**：添加了 `--toc-depth` 选项，控制目录的深度 ✅
3. **导航路径导出**：添加了 `--nav-path` 选项，支持导出特定导航路径的内容 ✅
4. **图片路径处理**：自动处理合并文件时的图片路径，确保在导出文档中正确显示 ✅
5. **YAML Front Matter 处理**：自动处理 Markdown 文件中的 YAML Front Matter ✅