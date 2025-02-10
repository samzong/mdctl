[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# mdctl

[English](README.md) | [中文版](README_zh.md)

A command-line tool for processing Markdown files. Currently, it supports automatically downloading remote images to local storage and updating the image references in Markdown files.

## Features

- Supports processing individual Markdown files or entire directories (including subdirectories).
- Automatically downloads remote images to a specified local directory.
- Preserves original image filenames (where available).
- Handles duplicate filenames automatically.
- Updates image references in Markdown files using relative paths.
- Provides detailed processing logs.

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

## Command Reference

### `download` Command

Downloads and localizes remote images in Markdown files.

Parameters:
- `-f, --file`: Specifies the Markdown file to process.
- `-d, --dir`: Specifies the directory to process (recursively processes all Markdown files).
- `-o, --output`: Specifies the directory for saving images (optional).
  - Default: `images` subdirectory within the file's directory (file mode).
  - Default: `images` subdirectory within the specified directory (directory mode).

## Translate Command

The `translate` command allows you to translate markdown files or directories to a specified language.

### Configuration

Create a configuration file at `~/.config/mdctl/config.json` with the following content:

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

### Usage

```bash
# Translate a single file
mdctl translate --from path/to/source.md --to path/to/target.md --locales en

# Translate a directory
mdctl translate --from path/to/source/dir --to path/to/target/dir --locales zh

# Force translate even if already translated
mdctl translate --from path/to/source.md --locales en -f
```

### Features

- Supports translating single files or entire directories
- Maintains directory structure when translating directories
- Adds front matter to track translation status
- Supports force translation with `-f` flag
- Supports translation between English (en) and Chinese (zh)

## Config Command

The `config` command allows you to manage your configuration settings.

### Usage

```bash
# List all configuration settings
mdctl config list

# Get a specific configuration value
mdctl config get -k api_key
mdctl config get -k model

# Set a configuration value
mdctl config set -k api_key -v "your-api-key"
mdctl config set -k model -v "gpt-4"
mdctl config set -k temperature -v "0.8"
```

### Available Configuration Keys

- `translate_prompt`: The prompt template for translation
- `endpoint`: OpenAI API endpoint URL
- `api_key`: Your OpenAI API key
- `model`: The model to use for translation (e.g., gpt-3.5-turbo, gpt-4)
- `temperature`: Temperature setting for the model (0.0 to 1.0)
- `top_p`: Top P setting for the model (0.0 to 1.0)

## Notes

1. `-f` and `