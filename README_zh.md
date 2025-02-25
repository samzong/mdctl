[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# mdctl

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

### 下载远程图片

处理单个文件：
```bash
mdctl download -f path/to/your/file.md
```

处理整个目录：
```bash
mdctl download -d path/to/your/directory
```

指定图片输出目录：
```bash
mdctl download -f path/to/your/file.md -o path/to/images
```

## 命令说明

### download 命令

下载并本地化 Markdown 文件中的远程图片。

参数：
- `-f, --file`: 指定要处理的 Markdown 文件
- `-d, --dir`: 指定要处理的目录（将递归处理所有 Markdown 文件）
- `-o, --output`: 指定图片保存的目录（可选）
  - 文件模式下默认保存在文件所在目录的 `images` 子目录中
  - 目录模式下默认保存在目录下的 `images` 子目录中

## 翻译命令

`translate` 命令允许你使用 AI 模型将 Markdown 文件或目录翻译成指定的语言。

### 支持的 AI 模型

- OpenAI API（当前支持）
- Ollama（即将支持）
- Google Gemini（即将支持）
- Anthropic Claude（即将支持）

### 支持的语言

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

### 配置

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

### 使用方法

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

### 功能特点

- 支持翻译单个文件或整个目录
- 翻译目录时保持目录结构
- 添加 front matter 以跟踪翻译状态
- 支持使用 `-F` 标志强制翻译
- 支持多种语言之间的翻译
- 显示详细的翻译进度和状态信息

## 配置命令

`config` 命令允许你管理配置设置。

### 使用方法

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

### 可用的配置项

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