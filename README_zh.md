[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# mdctl

[English Version](README.md) | [中文版](README_zh.md)

一个用于处理 Markdown 文件的命令行工具。目前支持自动下载远程图片到本地，更新 Markdown 文件中的图片引用路径，使用 AI 模型翻译 Markdown 文件，以及将本地图片上传到云存储服务。

## 功能特点

- 支持处理单个 Markdown 文件或整个目录（包括子目录）
- 自动下载远程图片到本地指定目录
- 上传本地图片到云存储服务（S3、Cloudflare R2、MinIO 等）
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

### 上传本地图片到云存储

将单个文件中的本地图片上传到 S3：
```bash
mdctl upload -f path/to/your/file.md -p s3 -b your-bucket-name
```

将目录中的本地图片上传到 Cloudflare R2，并设置路径前缀：
```bash
mdctl upload -d path/to/your/directory -p r2 -b your-bucket-name --prefix blog/
```

使用自定义域名生成 URL：
```bash
mdctl upload -f path/to/your/file.md -p s3 -b your-bucket-name -c assets.yourdomain.com
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

### upload 命令

将 Markdown 文件中的本地图片上传到云存储并重写 URL。

参数：
- `-f, --file`: 指定要处理的 Markdown 文件
- `-d, --dir`: 指定要处理的目录（将递归处理所有 Markdown 文件）
- `-p, --provider`: 云存储提供商（s3、r2、minio）
- `-b, --bucket`: 云存储桶名称
- `-c, --custom-domain`: 生成 URL 的自定义域名（可选）
- `--prefix`: 上传文件的路径前缀（可选）
- `--dry-run`: 预览更改而不上传（可选）
- `--concurrency`: 并发上传数量（默认：5）
- `-F, --force`: 即使文件已存在也强制上传（可选）
- `--skip-verify`: 跳过自签名证书的 SSL 验证（可选）
- `--ca-cert`: CA 证书路径（可选）
- `--conflict`: 冲突策略（rename、version、overwrite）（默认：rename）
- `--cache-dir`: 缓存目录路径（可选）
- `--include`: 要包含的文件扩展名列表，以逗号分隔（可选）

### 配置

配置云存储设置：

```bash
# 设置云存储提供商
mdctl config set -k cloud_storage.provider -v "r2"

# 设置端点 URL
mdctl config set -k cloud_storage.endpoint -v "https://xxxx.r2.cloudflarestorage.com"

# 设置 Cloudflare 账户 ID（对 R2 可选，如果未提供会从endpoint中提取）
mdctl config set -k cloud_storage.account_id -v "your-account-id"

# 设置访问密钥和密钥
mdctl config set -k cloud_storage.access_key -v "YOUR_ACCESS_KEY"
mdctl config set -k cloud_storage.secret_key -v "YOUR_SECRET_KEY"

# 设置桶名称
mdctl config set -k cloud_storage.bucket -v "my-images"

# 设置其他选项
mdctl config set -k cloud_storage.custom_domain -v "assets.yourdomain.com"
mdctl config set -k cloud_storage.path_prefix -v "blog"
mdctl config set -k cloud_storage.concurrency -v "10"
mdctl config set -k cloud_storage.conflict_policy -v "rename"
```

### 多云存储配置

mdctl 支持管理多个云存储配置并在它们之间切换：

```bash
# 列出所有存储配置
mdctl config list-storages

# 配置 AWS S3 存储
mdctl config set -k cloud_storages.my-s3.provider -v "s3"
mdctl config set -k cloud_storages.my-s3.region -v "us-east-1"
mdctl config set -k cloud_storages.my-s3.access_key -v "YOUR_AWS_ACCESS_KEY"
mdctl config set -k cloud_storages.my-s3.secret_key -v "YOUR_AWS_SECRET_KEY"
mdctl config set -k cloud_storages.my-s3.bucket -v "my-s3-bucket"

# 配置 Cloudflare R2 存储
mdctl config set -k cloud_storages.my-r2.provider -v "r2"
mdctl config set -k cloud_storages.my-r2.endpoint -v "https://YOUR_ACCOUNT_ID.r2.cloudflarestorage.com"
mdctl config set -k cloud_storages.my-r2.account_id -v "YOUR_ACCOUNT_ID"
mdctl config set -k cloud_storages.my-r2.access_key -v "YOUR_R2_ACCESS_KEY"
mdctl config set -k cloud_storages.my-r2.secret_key -v "YOUR_R2_SECRET_KEY"
mdctl config set -k cloud_storages.my-r2.bucket -v "my-r2-bucket"

# 配置 MinIO 存储
mdctl config set -k cloud_storages.my-minio.provider -v "minio"
mdctl config set -k cloud_storages.my-minio.endpoint -v "http://localhost:9000"
mdctl config set -k cloud_storages.my-minio.region -v "auto"
mdctl config set -k cloud_storages.my-minio.access_key -v "minioadmin"
mdctl config set -k cloud_storages.my-minio.secret_key -v "minioadmin"
mdctl config set -k cloud_storages.my-minio.bucket -v "my-minio-bucket"

# 设置默认存储配置
mdctl config set-default-storage --name my-r2

# 上传时指定使用特定的存储配置
mdctl upload -f README.md --storage my-s3
mdctl upload -d docs/ --storage my-minio
```

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
mdctl config get -k cloud_storage.provider

# 设置配置值
mdctl config set -k api_key -v "你的API密钥"
mdctl config set -k model -v "gpt-4"
mdctl config set -k temperature -v "0.8"
mdctl config set -k cloud_storage.provider -v "s3"
mdctl config set -k cloud_storage.bucket -v "my-bucket"
```

### 可用的配置项

- `translate_prompt`: 翻译提示模板
- `endpoint`: AI 模型 API 端点 URL
- `api_key`: 你的 API 密钥
- `model`: 用于翻译的模型
- `temperature`: 模型温度设置（0.0 到 1.0）
- `top_p`: 模型 Top P 设置（0.0 到 1.0）
- `cloud_storage.provider`: 云存储提供商（s3、r2、minio）
- `cloud_storage.region`: 云提供商的区域
- `cloud_storage.endpoint`: 云提供商的端点 URL
- `cloud_storage.access_key`: 云提供商的访问密钥
- `cloud_storage.secret_key`: 云提供商的密钥
- `cloud_storage.bucket`: 桶名称
- `cloud_storage.custom_domain`: 生成 URL 的自定义域名
- `cloud_storage.path_prefix`: 上传文件的路径前缀
- `cloud_storage.concurrency`: 并发上传数量
- `cloud_storage.skip_verify`: 是否跳过 SSL 验证
- `cloud_storage.ca_cert_path`: CA 证书路径
- `cloud_storage.conflict_policy`: 冲突策略（rename、version、overwrite）
- `cloud_storage.cache_dir`: 缓存目录路径

## 注意事项

1. 不能同时使用 `-f` 和 `-d` 参数
2. 如果不指定输出目录，工具会自动创建默认的 `images` 目录
3. download 命令只会处理远程图片（http/https）
4. upload 命令只会处理本地图片
5. 图片文件名会保持原有名称，如果发生重名，会自动添加 URL 或文件的哈希值作为后缀 