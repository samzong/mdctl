# mdctl

[中文版](README_zh.md)

A command-line tool for processing Markdown files. Currently, it supports automatically downloading remote images to local storage and updating the image references in Markdown files.

## Features

- Supports processing individual Markdown files or entire directories (including subdirectories).
- Automatically downloads remote images to a specified local directory.
- Preserves original image filenames (where available).
- Handles duplicate filenames automatically.
- Updates image references in Markdown files using relative paths.
- Provides detailed processing logs.

## Installation

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

## Notes

1. `-f` and `-d` parameters cannot be specified simultaneously.
2. If no output directory is specified, the tool will automatically create a default `images` directory.
3. The tool processes only remote images (http/https); local image references are not modified.
4. Image filenames are preserved; if duplicates are encountered, the URL's hash value is appended as a suffix.