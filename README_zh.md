[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[English Version](README.md) | [中文版](README_zh.md)

一个用于处理 Markdown 文件的命令行工具。目前支持自动下载远程图片到本地，更新 Markdown 文件中的图片引用路径，以及使用 AI 模型翻译 Markdown 文件。

## 功能特点

- 支持处理单个 Markdown 文件或整个目录（包括子目录）
- 自动下载远程图片到本地指定目录
- 保持原有图片文件名（如果可用）
- 自动处理重名文件
- 使用相对路径更新 Markdown 文件中的图片引用
- 详细的处理日志输出
- 使用 AI 模型翻译 Markdown 文件，支持多种语言
- 从网站站点地图生成适用于训练语言模型的 llms.txt 文件
- 将 Markdown 文件导出为各种文档格式（DOCX、PDF、EPUB），支持多种自定义选项
- 将 Markdown 文件中的本地图片上传到云存储服务并更新引用

## 安装

使用 Homebrew 安装 mdctl。请参考 [Homebrew 安装指南](https://brew.sh/) 安装 Homebrew。

```bash
brew tap samzong/tap
brew install samzong/tap/mdctl
```

或者使用 go 安装 mdctl。

```bash
go install github.com/samzong/mdctl@latest
```

## 使用方法

常见任务的快速示例：

### 下载图片
```bash
# 处理单个文件
mdctl download -f path/to/your/file.md

# 处理整个目录
mdctl download -d path/to/your/directory
```

### 导出文档
```bash
# 导出为 DOCX
mdctl export -f README.md -o output.docx

# 导出为带目录的 PDF
mdctl export -d docs/ -o documentation.pdf -F pdf --toc
```

### 上传到云存储
```bash
# 上传单个文件中的图片
mdctl upload -f post.md

# 上传目录中的图片
mdctl upload -d docs/
```

### 翻译内容
```bash
# 翻译为中文
mdctl translate -f README.md -l zh

# 将目录翻译为日语
mdctl translate -d docs/ -l ja
```

### 生成 LLMS.txt
```bash
# 标准模式（标题和描述）
mdctl llmstxt https://example.com/sitemap.xml > llms.txt

# 全内容模式
mdctl llmstxt -f https://example.com/sitemap.xml > llms-full.txt
```

## 命令说明

### `download` 命令

下载并本地化 Markdown 文件中的远程图片。

参数：
- `-f, --file`: 指定要处理的 Markdown 文件
- `-d, --dir`: 指定要处理的目录（将递归处理所有 Markdown 文件）
- `-o, --output`: 指定图片保存的目录（可选）
  - 文件模式下默认保存在文件所在目录的 `images` 子目录中
  - 目录模式下默认保存在目录下的 `images` 子目录中

### `export` 命令

`export` 命令使用 Pandoc 作为底层转换工具，将 Markdown 文件转换为各种文档格式，如 DOCX、PDF 和 EPUB。

功能特点：
- 支持导出单个文件或整个目录
- 从 Markdown 内容生成专业文档
- 支持多种文档格式（DOCX、PDF、EPUB）
- 生成目录
- 自定义文档模板
- 标题级别调整
- 支持不同的站点结构（basic、MkDocs、Hugo、Docusaurus）

使用方法：
```bash
# 将单个文件导出为 DOCX
mdctl export -f README.md -o output.docx

# 将目录导出为 PDF
mdctl export -d docs/ -o documentation.pdf -F pdf

# 使用模板并生成目录
mdctl export -d docs/ -o report.docx -t templates/corporate.docx --toc

# 调整标题级别导出
mdctl export -d docs/ -o documentation.docx --shift-heading-level-by 2

# 从特定类型的站点导出（MkDocs、Hugo 等）
mdctl export -d docs/ -s mkdocs -o site_docs.docx
```

参数：
- `-f, --file`: 要导出的源 Markdown 文件
- `-d, --dir`: 包含要导出的 Markdown 文件的源目录
- `-o, --output`: 输出文件路径（必需）
- `-F, --format`: 输出格式（docx、pdf、epub）（默认：docx）
- `-t, --template`: Word 模板文件路径
- `-s, --site-type`: 站点类型（basic、mkdocs、hugo、docusaurus）（默认：basic）
- `--toc`: 生成目录
- `--toc-depth`: 目录深度（默认：3）
- `--shift-heading-level-by`: 标题级别偏移量
- `--file-as-title`: 使用文件名作为章节标题
- `-n, --nav-path`: 指定要导出的导航路径
- `-v, --verbose`: 启用详细输出

### `upload` 命令

`upload` 命令将 Markdown 文件中的本地图片上传到云存储，并重写 URL 以引用上传后的版本。

功能特点：
- 支持具有 S3 兼容 API 的多种云存储提供商
- 保持相对链接结构
- 可配置的上传路径前缀
- 生成 URL 的自定义域名支持
- 冲突解决策略
- 并行上传提高性能
- 预览模式以查看更改而不实际上传

使用方法：
```bash
# 上传单个文件中的图片
mdctl upload -f post.md

# 上传目录中的图片
mdctl upload -d docs/

# 使用特定提供商和存储桶上传
mdctl upload -f post.md -p s3 -b my-bucket

# 使用自定义域名生成 URL
mdctl upload -f post.md -c images.example.com

# 预览更改而不上传
mdctl upload -f post.md --dry-run

# 强制上传即使文件已存在
mdctl upload -f post.md --force
```

参数：
- `-f, --file`: 要处理的源 Markdown 文件
- `-d, --dir`: 包含要处理的 Markdown 文件的源目录
- `-p, --provider`: 云存储提供商（s3、r2、minio）
- `-b, --bucket`: 云存储桶名称
- `-c, --custom-domain`: 生成 URL 的自定义域名
- `--prefix`: 上传文件的路径前缀
- `--dry-run`: 预览更改而不上传
- `--concurrency`: 并行上传数量（默认：5）
- `-F, --force`: 即使文件存在也强制上传
- `--skip-verify`: 跳过 SSL 验证
- `--ca-cert`: CA 证书路径
- `--conflict`: 冲突策略（rename、version、overwrite）（默认：rename）
- `--cache-dir`: 缓存目录路径
- `--include`: 要包含的文件扩展名（逗号分隔列表）
- `--storage`: 使用的存储配置名称
- `-v, --verbose`: 启用详细输出

### `llmstxt` 命令

`llmstxt` 命令从网站的 sitemap.xml 生成一个适用于训练或微调语言模型的 llms.txt 文件。

功能特点：
- 以结构化格式提取网站内容
- 支持标准模式（仅标题和描述）和全内容模式
- 允许使用包含/排除模式过滤页面
- 可配置的并发和超时设置
- 输出到文件或标准输出

使用方法：
```bash
# 标准模式 - 仅提取标题和描述
mdctl llmstxt https://example.com/sitemap.xml > llms.txt

# 全内容模式 - 提取完整页面内容
mdctl llmstxt -f https://example.com/sitemap.xml > llms-full.txt

# 指定输出文件
mdctl llmstxt -o output.txt https://example.com/sitemap.xml

# 包含/排除特定路径
mdctl llmstxt -i "/blog/*" -e "/blog/draft/*" https://example.com/sitemap.xml

# 配置并发和超时
mdctl llmstxt -c 10 --timeout 60 https://example.com/sitemap.xml
```

参数：
- `[url]`: sitemap.xml 文件的 URL（必需）
- `-o, --output`: 输出文件路径（默认：标准输出）
- `-i, --include-path`: 要包含的路径的 Glob 模式（可多次指定）
- `-e, --exclude-path`: 要排除的路径的 Glob 模式（可多次指定）
- `-f, --full`: 启用全内容模式（提取页面内容）
- `-c, --concurrency`: 并发请求数（默认：5）
- `--timeout`: 请求超时（秒）（默认：30）
- `-v, --verbose`: 启用详细输出

### `translate` 命令

`translate` 命令允许你使用 AI 模型将 Markdown 文件或目录翻译成指定的语言。

#### 支持的 AI 模型

- OpenAI API（当前支持）
- Ollama（即将支持）
- Google Gemini（即将支持）
- Anthropic Claude（即将支持）

#### 支持的语言

- 阿拉伯语 (العربية)
- 中文 (中文)
- 英语 (English)
- 法语 (Français)
- 德语 (Deutsch)
- 印地语 (हिन्दी)
- 意大利语 (Italiano)
- 日语 (日本語)
- 韩语 (한국어)
- 葡萄牙语 (Português)
- 俄语 (Русский)
- 西班牙语 (Español)
- 泰语 (ไทย)
- 越南语 (Tiếng Việt)

#### 配置

在 `~/.config/mdctl/config.json` 创建配置文件，内容如下：

```json
{
  "translate_prompt": "Please translate the following markdown content to {TARGET_LANG}, keep the markdown format and front matter unchanged:",
  "endpoint": "你的API端点",
  "api_key": "你的API密钥",
  "model": "gpt-3.5-turbo",
  "temperature": 0.0,
  "top_p": 1.0
}
```

或者使用配置命令进行设置：

```bash
# 设置 API 端点
mdctl config set -k endpoint -v "你的API端点"

# 设置 API 密钥
mdctl config set -k api_key -v "你的API密钥"

# 设置模型
mdctl config set -k model -v "gpt-3.5-turbo"
```

#### 使用方法

```bash
# 将单个文件翻译成中文
mdctl translate -f README.md -l zh

# 将目录翻译成日语
mdctl translate -f docs -l ja

# 强制翻译已翻译过的文件
mdctl translate -f README.md -l ko -F

# 翻译到指定的输出路径
mdctl translate -f docs -l fr -t translated_docs
```

功能特点：
- 支持翻译单个文件或整个目录
- 翻译目录时保持目录结构
- 添加 front matter 以跟踪翻译状态
- 支持使用 `-F` 标志强制翻译
- 支持多种语言之间的翻译
- 显示详细的翻译进度和状态信息

### `config` 命令

`config` 命令允许你管理配置设置。

使用方法：

```bash
# 列出所有配置设置
mdctl config list

# 获取特定配置值
mdctl config get -k api_key
mdctl config get -k model

# 设置配置值
mdctl config set -k api_key -v "你的API密钥"
mdctl config set -k model -v "gpt-4"
mdctl config set -k temperature -v "0.8"
```

可用的配置项：

- `translate_prompt`: 翻译提示模板
- `endpoint`: AI 模型 API 端点 URL
- `api_key`: 你的 API 密钥
- `model`: 用于翻译的模型
- `temperature`: 模型温度设置（0.0 到 1.0）
- `top_p`: 模型 Top P 设置（0.0 到 1.0）

## 注意事项

1. 不能同时使用 `-f` 和 `-d` 参数
2. 如果不指定输出目录，工具会自动创建默认的 `images` 目录
3. 工具只会处理远程图片（http/https），本地图片引用不会被修改
4. 图片文件名会保持原有名称，如果发生重名，会自动添加 URL 的哈希值作为后缀 