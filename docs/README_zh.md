---
translated: true
---


[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# mdctl

[English](README.md) | [中文版](README_zh.md)

一个用于处理Markdown文件的命令行工具。目前支持自动将远程图片下载到本地存储，并更新Markdown文件中的图片引用。

## 功能

- 支持处理单个Markdown文件或整个目录（包括子目录）。
- 自动将远程图片下载到指定的本地目录。
- 保留原始图片文件名（如果可用）。
- 自动处理重复文件名。
- 使用相对路径更新Markdown文件中的图片引用。
- 提供详细的处理日志。

## 安装

使用Homebrew安装mdctl。按照[Homebrew安装指南](https://brew.sh/)安装Homebrew。

```bash
brew tap samzong/tap
brew install samzong/tap/mdctl
```

或使用go安装mdctl。

```bash
go install github.com/samzong/mdctl@latest
```

## 使用

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

## 命令参考

### `download` 命令

下载并本地化Markdown文件中的远程图片。

参数：
- `-f, --file`: 指定要处理的Markdown文件。
- `-d, --dir`: 指定要处理的目录（递归处理所有Markdown文件）。
- `-o, --output`: 指定保存图片的目录（可选）。
  - 默认：文件所在目录下的`images`子目录（文件模式）。
  - 默认：指定目录下的`images`子目录（目录模式）。

## 翻译命令

`translate`命令允许你将Markdown文件或目录翻译成指定语言。

### 配置

在`~/.config/mdctl/config.json`创建配置文件，内容如下：

```json
{
  "translate_prompt": "Please translate the following markdown content to {TARGET_LANG}, keep the markdown format and front matter unchanged:",
  "endpoint": "https://api.openai.com/v1",
  "api_key": "your-api-key",
  "model": "gpt-3.5-turbo",
  "temperature": 0.0,
  "top_p": 1.0
}
```

### 使用

```bash
# 翻译单个文件
mdctl translate --from path/to/source.md --to path/to/target.md --locales en

# 翻译整个目录
mdctl translate --from path/to/source/dir --to path/to/target/dir --locales zh

# 强制翻译即使已翻译
mdctl translate --from path/to/source.md --locales en -f
```

### 功能

- 支持翻译单个文件或整个目录
- 翻译目录时保持目录结构
- 添加front matter以跟踪翻译状态
- 支持使用`-f`标志强制翻译
- 支持在英语（en）和中文（zh）之间翻译

## 配置命令

`config`命令允许你管理配置设置。

### 使用

```bash
# 列出所有配置设置
mdctl config list

# 获取特定配置值
mdctl config get -k api_key
mdctl config get -k model

# 设置配置值
mdctl config set -k api_key -v "your-api-key"
mdctl config set -k model -v "gpt-4"
mdctl config set -k temperature -v "0.8"
```

### 可用配置键

- `translate_prompt`: 翻译提示模板
- `endpoint`: OpenAI API端点URL
- `api_key`: 你的OpenAI API密钥
- `model`: 用于翻译的模型（例如，gpt-3.5-turbo, gpt-4）
- `temperature`: 模型的温度设置（0.0到1.0）
- `top_p`: 模型的Top P设置（0.0到1.0）

## 注意事项

1. `-f`和`