# mdctl 开发者指南

## 项目介绍

mdctl 是一个用于处理 Markdown 文件的命令行工具，主要功能包括：

1. **下载功能**：自动下载 Markdown 文件中的远程图片到本地，并更新引用路径
2. **翻译功能**：使用 AI 模型将 Markdown 文件翻译成多种语言
3. **上传功能**：将本地图片上传到云存储，并更新 Markdown 文件中的引用
4. **配置管理**：管理工具的配置信息
5. **其他功能**：如导出为其他格式、生成 llms.txt 文件等

## 项目结构

```bash
../mdctl
├── cmd
│   ├── config.go
│   ├── download.go
│   ├── export.go
│   ├── llmstxt.go
│   ├── root.go
│   ├── translate.go
│   └── upload.go
├── internal
│   ├── cache
│   ├── config
│   ├── exporter
│   ├── llmstxt
│   ├── markdownfmt
│   ├── processor
│   ├── storage
│   ├── translator
│   └── uploader
├── main.go
├── go.mod
├── go.sum
```

## 核心模块说明

### 命令行模块 (cmd/)

使用 [Cobra](https://github.com/spf13/cobra) 库实现命令行界面，主要命令包括：

- **root**: 根命令，定义基本信息和版本
- **download**: 下载远程图片到本地
- **translate**: 翻译 Markdown 文件
- **upload**: 上传本地图片到云存储
- **config**: 管理配置信息

### 处理器模块 (internal/processor/)

负责处理 Markdown 文件中的远程图片下载，主要功能：

- 解析 Markdown 文件中的图片链接
- 下载远程图片到本地
- 更新 Markdown 文件中的图片引用路径

### 翻译模块 (internal/translator/)

负责翻译 Markdown 文件，主要功能：

- 支持多种语言翻译
- 保持 Markdown 格式和 front matter 不变
- 使用 AI 模型进行翻译
- 支持目录结构的翻译

### 上传模块 (internal/uploader/)

负责上传本地图片到云存储，主要功能：

- 解析 Markdown 文件中的本地图片链接
- 上传图片到云存储
- 更新 Markdown 文件中的图片引用路径
- 支持多种冲突处理策略

### 存储模块 (internal/storage/)

定义存储提供者接口和实现，主要功能：

- 提供统一的存储接口
- 支持 S3 兼容的存储服务
- 处理文件上传和元数据管理

### llms.txt 生成模块 (internal/llmstxt/)

负责从网站的 sitemap.xml 生成 llms.txt 文件，主要功能：

- 解析 sitemap.xml 文件
- 访问每个 URL 并提取页面内容
- 生成格式化的 llms.txt 文档

### 配置模块 (internal/config/)

负责管理配置信息，主要功能：

- 加载和保存配置文件
- 管理 AI 模型配置
- 管理云存储配置

## 开发风格和约定

### 代码组织

1. **命令与实现分离**：命令行接口在 `cmd/` 目录，具体实现在 `internal/` 目录
2. **模块化设计**：每个功能都有独立的模块，如处理器、翻译器、上传器等
3. **接口定义**：使用接口定义模块间交互，如存储提供者接口

### 错误处理

错误处理采用 Go 语言的标准方式，通过返回错误值进行传递和处理。

### 配置管理

配置文件存储在 `~/.config/mdctl/config.json`，包含：

- AI 模型配置（端点、API 密钥、模型名称等）
- 云存储配置（提供者、区域、访问密钥等）

### 日志输出

使用标准输出进行日志记录，提供详细的处理信息和错误信息。

## 添加新功能的步骤

1. **定义命令**：在 `cmd/` 目录下创建新的命令文件，定义命令行接口
2. **实现功能**：在 `internal/` 目录下创建相应的实现模块
3. **注册命令**：在 `cmd/root.go` 的 `init()` 函数中注册新命令
4. **更新文档**：更新 README 文件，添加新功能的说明

## 构建和发布

项目使用 Makefile 和 GoReleaser 进行构建和发布：

- **构建**：使用 `make build` 命令构建项目
- **发布**：使用 `make release` 命令发布新版本

## 扩展点

### 添加新的存储提供者

1. 在 `internal/storage/` 目录下创建新的提供者实现
2. 实现 `Provider` 接口
3. 在初始化时注册提供者

### 添加新的 AI 模型支持

1. 在 `internal/translator/` 目录下扩展翻译器实现
2. 添加新模型的 API 调用
3. 更新配置模块以支持新模型的配置

### 添加新的 Markdown 处理功能

1. 创建新的处理器模块
2. 实现 Markdown 解析和处理逻辑
3. 添加新的命令行接口
