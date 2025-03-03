[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fsamzong%2Fmdctl.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fsamzong%2Fmdctl?ref=badge_shield)

[English](README.md) | [中文版](README_zh.md)

A command-line tool for processing Markdown files. Currently, it supports automatically downloading remote images to local storage and updating the image references in Markdown files, as well as translating markdown files using AI models.

## Features

- Supports processing individual Markdown files or entire directories (including subdirectories).
- Automatically downloads remote images to a specified local directory.
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

# Set a configuration value
mdctl config set -k api_key -v "your-api-key"
mdctl config set -k model -v "gpt-4"
mdctl config set -k temperature -v "0.8"
```

### Available Configuration Keys

- `translate_prompt`: The prompt template for translation
- `endpoint`: AI model API endpoint URL
- `api_key`: Your API key
- `model`: The model to use for translation
- `temperature`: Temperature setting for the model (0.0 to 1.0)
- `top_p`: Top P setting for the model (0.0 to 1.0)

## Notes

1. `-f` and `-d` parameters cannot be used together
2. If no output directory is specified, the tool will automatically create a default `images` directory
3. Only remote images (http/https) will be processed, local image references will not be modified
4. Image filenames will retain their original names, if duplicates occur, a hash of the URL will be added as a suffix

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fsamzong%2Fmdctl.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fsamzong%2Fmdctl?ref=badge_large)