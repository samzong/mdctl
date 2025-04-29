# mdctl - A CLI Tool for Markdown File Operations

<div align="center">
  <img src="./mdctl.png" alt="mdctl logo" width="200" />
  <br />
  <p>An AI-powered CLI tool to enhance your Markdown workflow, with auto-image downloading, translation, and more features coming soon!</p>
  <p>
    <a href="https://github.com/samzong/mdctl/actions/workflows/docker-build.yml"><img src="https://github.com/samzong/mdctl/actions/workflows/docker-build.yml/badge.svg" alt="Build Status"></a>
    <a href="https://github.com/samzong/mdctl/releases"><img src="https://img.shields.io/github/v/release/samzong/mdctl" alt="Release Version" /></a>
    <a href="https://github.com/samzong/mdctl/blob/main/LICENSE"><img src="https://img.shields.io/github/license/samzong/mdctl" alt="MIT License" /></a>
  </p>
</div>

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

## Developer's Guide

If you are interested in contributing, please refer to the [DEVELOPMENT.md](docs/DEVELOPMENT.md) file for a complete technical architecture, component design, and development guide.

## Contributing

Welcome to contribute code, report issues, or suggest features! Please follow these steps:

1. Fork this repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
