[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# mdctl

A command-line tool for processing Markdown files. Currently, it supports automatically downloading remote images to local storage and updating the image references in Markdown files, as well as translating markdown files using AI models.

## Key Features

- Automatically downloads remote images to a specified local directory.
- Translates markdown files using AI models with support for multiple languages.
- Uploads local images in markdown files to cloud storage services and updates references.
- Exports markdown files to various document formats (DOCX, PDF, EPUB) with customization options.
- Generates llms.txt files from website sitemaps for training language models.

## Installation

Use Homebrew to install mdctl. Follow the [Homebrew Installation Guide](https://brew.sh/) to install Homebrew.

```bash
brew tap samzong/tap
brew install samzong/tap/mdctl
```

Or use go to install mdctl.

```bash
go install github.com/samzong/mdctl@latest
```

## Usage

Quick examples for common tasks:

### Downloading Images

```bash
# Process a single file
mdctl download -f path/to/your/file.md

# Process a directory
mdctl download -d path/to/your/directory
```

### Translating I18n

```bash
# Translate to Chinese
mdctl translate -f README.md -l zh

# Translate a directory to Japanese
mdctl translate -d docs/ -l ja
```

### Uploading Images to Cloud Storage

```bash
# Upload images from a file
mdctl upload -f post.md

# Upload images from a directory
mdctl upload -d docs/
```

### Exporting Documents to `.docx`

```bash
# Export to DOCX
mdctl export -f README.md -o output.docx

# Export to PDF with table of contents
mdctl export -d docs/ -o documentation.pdf -F pdf --toc
```

### Generating `llms.txt` from `sitemap.xml`

```bash
# Standard mode (titles and descriptions)
mdctl llmstxt https://example.com/sitemap.xml > llms.txt

# Full-content mode
mdctl llmstxt -f https://example.com/sitemap.xml > llms-full.txt
```

## Command Reference

### `download`

Downloads and localizes remote images in Markdown files.

Parameters:

- `-f, --file`: Specifies the Markdown file to process.
- `-d, --dir`: Specifies the directory to process (recursively processes all Markdown files).
- `-o, --output`: Specifies the directory for saving images (optional).
  - Default: `images` subdirectory within the file's directory (file mode).
  - Default: `images` subdirectory within the specified directory (directory mode).

### `translate`

The `translate` command allows you to translate markdown files or directories to a specified language using AI models.

#### Supported AI Models

- OpenAI API (Current)
- Ollama (Coming Soon)
- Google Gemini (Coming Soon)
- Anthropic Claude (Coming Soon)

#### Supported Languages

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

#### Configuration

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

Features:

- Supports translating single files or entire directories
- Maintains directory structure when translating directories
- Adds front matter to track translation status
- Supports force translation with `-F` flag
- Supports translation between multiple languages
- Shows translation progress with detailed status information

### `upload`

The `upload` command uploads local images in markdown files to cloud storage and rewrites URLs to reference the uploaded versions.

Features:

- Supports multiple cloud storage providers with S3-compatible APIs
- Preserves relative link structure
- Configurable upload path prefixes
- Custom domain support for generated URLs
- Conflict resolution policies
- Concurrent uploads for better performance
- Dry-run mode for previewing changes

Parameters:

- `-f, --file`: Source markdown file to process
- `-d, --dir`: Source directory containing markdown files to process
- `-p, --provider`: Cloud storage provider (s3, r2, minio)
- `-b, --bucket`: Cloud storage bucket name
- `-c, --custom-domain`: Custom domain for generated URLs
- `--prefix`: Path prefix for uploaded files
- `--dry-run`: Preview changes without uploading
- `--concurrency`: Number of concurrent uploads (default: 5)
- `-F, --force`: Force upload even if file exists
- `--skip-verify`: Skip SSL verification
- `--ca-cert`: Path to CA certificate
- `--conflict`: Conflict policy (rename, version, overwrite) (default: rename)
- `--cache-dir`: Cache directory path
- `--include`: Comma-separated list of file extensions to include
- `--storage`: Storage name to use from configuration
- `-v, --verbose`: Enable verbose output

### `export`

The `export` command converts markdown files to various document formats like DOCX, PDF, and EPUB using Pandoc as the underlying conversion tool.

Features:

- Supports exporting single files or entire directories
- Generates professional documents from markdown content
- Supports various document formats (DOCX, PDF, EPUB)
- Table of contents generation
- Custom document templates
- Heading level adjustment
- Support for different site structures (basic, MkDocs, Hugo, Docusaurus)

Parameters:

- `-f, --file`: Source markdown file to export
- `-d, --dir`: Source directory containing markdown files to export
- `-o, --output`: Output file path (required)
- `-F, --format`: Output format (docx, pdf, epub) (default: docx)
- `-t, --template`: Word template file path
- `-s, --site-type`: Site type (basic, mkdocs, hugo, docusaurus) (default: basic)
- `--toc`: Generate table of contents
- `--toc-depth`: Depth of table of contents (default: 3)
- `--shift-heading-level-by`: Shift heading level by N
- `--file-as-title`: Use filename as section title
- `-n, --nav-path`: Specify the navigation path to export
- `-v, --verbose`: Enable verbose output

### `llmstxt`

The `llmstxt` command generates a markdown-formatted text file from a website's sitemap.xml, which is ideal for training or fine-tuning language models.

Features:

- Extracts website content in a structured format
- Supports standard mode (title and description only) and full-content mode
- Allows filtering pages with include/exclude patterns
- Configurable concurrency and timeout settings
- Output to file or stdout

Parameters:

- `[url]`: The URL to the sitemap.xml file (required)
- `-o, --output`: Output file path (default: stdout)
- `-i, --include-path`: Glob patterns for paths to include (can be specified multiple times)
- `-e, --exclude-path`: Glob patterns for paths to exclude (can be specified multiple times)
- `-f, --full`: Enable full-content mode (extract page content)
- `-c, --concurrency`: Number of concurrent requests (default: 5)
- `--timeout`: Request timeout in seconds (default: 30)
- `-v, --verbose`: Enable verbose output

## Notes

1. `-f` and `-d` parameters cannot be used together
2. If no output directory is specified, the tool will automatically create a default `images` directory
3. Only remote images (http/https) will be processed, local image references will not be modified
4. Image filenames will retain their original names, if duplicates occur, a hash of the URL will be added as a suffix
