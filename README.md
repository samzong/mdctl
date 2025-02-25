[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fsamzong%2Fmdctl.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fsamzong%2Fmdctl?ref=badge_shield)

# mdctl

![logo](https://mdctl.samzong.me/blog/images/images/logo_5dde22d2.png)

[English](README.md) | [中文版](README_zh.md)

A command-line tool for processing Markdown files. Currently, it supports automatically downloading remote images to local storage and updating the image references in Markdown files, translating markdown files using AI models, and uploading local images to cloud storage.

## Features

- Supports processing individual Markdown files or entire directories (including subdirectories).
- Automatically downloads remote images to a specified local directory.
- Uploads local images to cloud storage (S3, Cloudflare R2, MinIO, etc.).
- Preserves original image filenames (where available).
- Handles duplicate filenames automatically.
- Updates image references in Markdown files using relative paths.
- Provides detailed processing logs.
- Translates markdown files using AI models with support for multiple languages.

## Installation

use Homebrew to install mdctl. Follow the [Homebrew Installation Guide](https://brew.sh/) to install Homebrew.

```bash
brew tap samzong/tap
brew install samzong/tap/mdctl
```

or use go to install mdctl.

```bash
go install github.com/samzong/mdctl@latest
```

## Usage

### Downloading Remote Images

Processing a single file:
```bash
mdctl download -f path/to/your/file.md
```

Processing an entire directory:
```bash
mdctl download -d path/to/your/directory
```

Specifying an output directory for images:
```bash
mdctl download -f path/to/your/file.md -o path/to/images
```

### Uploading Local Images to Cloud Storage

Upload images from a single file to S3:
```bash
mdctl upload -f path/to/your/file.md -p s3 -b your-bucket-name
```

Upload images from a directory to Cloudflare R2 with a path prefix:
```bash
mdctl upload -d path/to/your/directory -p r2 -b your-bucket-name --prefix blog/
```

Use a custom domain for generated URLs:
```bash
mdctl upload -f path/to/your/file.md -p s3 -b your-bucket-name -c assets.yourdomain.com
```

## Command Reference

### `download` Command

Downloads and localizes remote images in Markdown files.

Parameters:
- `-f, --file`: Specifies the Markdown file to process.
- `-d, --dir`: Specifies the directory to process (recursively processes all Markdown files).
- `-o, --output`: Specifies the directory for saving images (optional).
  - Default: `images` subdirectory within the file's directory (file mode).
  - Default: `images` subdirectory within the specified directory (directory mode).

### `upload` Command

Uploads local images in Markdown files to cloud storage and rewrites the URLs.

Parameters:
- `-f, --file`: Specifies the Markdown file to process.
- `-d, --dir`: Specifies the directory to process (recursively processes all Markdown files).
- `-p, --provider`: Cloud storage provider (s3, r2, minio).
- `-b, --bucket`: Cloud storage bucket name.
- `-c, --custom-domain`: Custom domain for generated URLs (optional).
- `--prefix`: Path prefix for uploaded files (optional).
- `--dry-run`: Preview changes without uploading (optional).
- `--concurrency`: Number of concurrent uploads (default: 5).
- `-F, --force`: Force upload even if file exists (optional).
- `--skip-verify`: Skip SSL verification for self-signed certificates (optional).
- `--ca-cert`: Path to CA certificate (optional).
- `--conflict`: Conflict policy (rename, version, overwrite) (default: rename).
- `--cache-dir`: Cache directory path (optional).
- `--include`: Comma-separated list of file extensions to include (optional).

### Configuration

To configure cloud storage settings:

```bash
# Set cloud storage provider
mdctl config set -k cloud_storage.provider -v "r2"

# Set endpoint URL
mdctl config set -k cloud_storage.endpoint -v "https://xxxx.r2.cloudflarestorage.com"

# Set Cloudflare account ID (optional for R2, will be extracted from endpoint if not provided)
mdctl config set -k cloud_storage.account_id -v "your-account-id"

# Set access and secret keys
mdctl config set -k cloud_storage.access_key -v "YOUR_ACCESS_KEY"
mdctl config set -k cloud_storage.secret_key -v "YOUR_SECRET_KEY"

# Set bucket name
mdctl config set -k cloud_storage.bucket -v "my-images"

# Set additional options
mdctl config set -k cloud_storage.custom_domain -v "assets.yourdomain.com"
mdctl config set -k cloud_storage.path_prefix -v "blog"
mdctl config set -k cloud_storage.concurrency -v "10"
mdctl config set -k cloud_storage.conflict_policy -v "rename"
```

### Multiple Cloud Storage Configurations

mdctl supports managing multiple cloud storage configurations and switching between them:

```bash
# List all storage configurations
mdctl config list-storages

# Configure AWS S3 storage
mdctl config set -k cloud_storages.my-s3.provider -v "s3"
mdctl config set -k cloud_storages.my-s3.region -v "us-east-1"
mdctl config set -k cloud_storages.my-s3.access_key -v "YOUR_AWS_ACCESS_KEY"
mdctl config set -k cloud_storages.my-s3.secret_key -v "YOUR_AWS_SECRET_KEY"
mdctl config set -k cloud_storages.my-s3.bucket -v "my-s3-bucket"

# Configure Cloudflare R2 storage
mdctl config set -k cloud_storages.my-r2.provider -v "r2"
mdctl config set -k cloud_storages.my-r2.endpoint -v "https://YOUR_ACCOUNT_ID.r2.cloudflarestorage.com"
mdctl config set -k cloud_storages.my-r2.account_id -v "YOUR_ACCOUNT_ID"
mdctl config set -k cloud_storages.my-r2.access_key -v "YOUR_R2_ACCESS_KEY"
mdctl config set -k cloud_storages.my-r2.secret_key -v "YOUR_R2_SECRET_KEY"
mdctl config set -k cloud_storages.my-r2.bucket -v "my-r2-bucket"

# Configure MinIO storage
mdctl config set -k cloud_storages.my-minio.provider -v "minio"
mdctl config set -k cloud_storages.my-minio.endpoint -v "http://localhost:9000"
mdctl config set -k cloud_storages.my-minio.region -v "auto"
mdctl config set -k cloud_storages.my-minio.access_key -v "minioadmin"
mdctl config set -k cloud_storages.my-minio.secret_key -v "minioadmin"
mdctl config set -k cloud_storages.my-minio.bucket -v "my-minio-bucket"

# Set the default storage configuration
mdctl config set-default-storage --name my-r2

# Use a specific storage configuration for upload
mdctl upload -f README.md --storage my-s3
mdctl upload -d docs/ --storage my-minio
```

## Translate Command

The `translate` command allows you to translate markdown files or directories to a specified language using AI models.

### Supported AI Models

- OpenAI API (Current)
- Ollama (Coming Soon)
- Google Gemini (Coming Soon)
- Anthropic Claude (Coming Soon)

### Supported Languages

- Arabic (العربية)
- Chinese (中文)
- English (English)
- French (Français)
- German (Deutsch)
- Hindi (हिन्दी)
- Italian (Italiano)
- Japanese (日本語)
- Korean (한국어)
- Portuguese (Português)
- Russian (Русский)
- Spanish (Español)
- Thai (ไทย)
- Vietnamese (Tiếng Việt)

### Configuration

Create a configuration file at `~/.config/mdctl/config.json` with the following content:

```json
{
  "translate_prompt": "Please translate the following markdown content to {TARGET_LANG}, keep the markdown format and front matter unchanged:",
  "endpoint": "your-api-endpoint",
  "api_key": "your-api-key",
  "model": "gpt-3.5-turbo",
  "temperature": 0.0,
  "top_p": 1.0
}
```

Or use the config command to set up:

```bash
# Set API endpoint
mdctl config set -k endpoint -v "your-api-endpoint"

# Set API key
mdctl config set -k api_key -v "your-api-key"

# Set model
mdctl config set -k model -v "gpt-3.5-turbo"
```

### Usage

```bash
# Translate a single file to Chinese
mdctl translate -f README.md -l zh

# Translate a directory to Japanese
mdctl translate -f docs -l ja

# Force translate an already translated file
mdctl translate -f README.md -l ko -F

# Translate to a specific output path
mdctl translate -f docs -l fr -t translated_docs
```

### Features

- Supports translating single files or entire directories
- Maintains directory structure when translating directories
- Adds front matter to track translation status
- Supports force translation with `-F` flag
- Supports translation between multiple languages
- Shows translation progress with detailed status information

## Config Command

The `config` command allows you to manage your configuration settings.

### Usage

```bash
# List all configuration settings
mdctl config list

# Get a specific configuration value
mdctl config get -k api_key
mdctl config get -k model
mdctl config get -k cloud_storage.provider

# Set a configuration value
mdctl config set -k api_key -v "your-api-key"
mdctl config set -k model -v "gpt-4"
mdctl config set -k temperature -v "0.8"
mdctl config set -k cloud_storage.provider -v "s3"
mdctl config set -k cloud_storage.bucket -v "my-bucket"
```

### Available Configuration Keys

- `translate_prompt`: The prompt template for translation
- `endpoint`: AI model API endpoint URL
- `api_key`: Your API key
- `model`: The model to use for translation
- `temperature`: Temperature setting for the model (0.0 to 1.0)
- `top_p`: Top P setting for the model (0.0 to 1.0)
- `cloud_storage.provider`: Cloud storage provider (s3, r2, minio)
- `cloud_storage.region`: Region for the cloud provider
- `cloud_storage.endpoint`: Endpoint URL for the cloud provider
- `cloud_storage.access_key`: Access key for the cloud provider
- `cloud_storage.secret_key`: Secret key for the cloud provider
- `cloud_storage.bucket`: Bucket name
- `cloud_storage.custom_domain`: Custom domain for generated URLs
- `cloud_storage.path_prefix`: Path prefix for uploaded files
- `cloud_storage.concurrency`: Number of concurrent uploads
- `cloud_storage.skip_verify`: Whether to skip SSL verification
- `cloud_storage.ca_cert_path`: Path to CA certificate
- `cloud_storage.conflict_policy`: Conflict policy (rename, version, overwrite)
- `cloud_storage.cache_dir`: Cache directory path

## Notes

1. `-f` and `-d` parameters cannot be used together
2. If no output directory is specified, the tool will automatically create a default `images` directory
3. Only remote images (http/https) will be processed by the download command
4. Only local images will be processed by the upload command
5. Image filenames will retain their original names, if duplicates occur, a hash of the URL will be added as a suffix

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fsamzong%2Fmdctl.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fsamzong%2Fmdctl?ref=badge_large)