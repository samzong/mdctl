package processor

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Processor struct {
	SourceFile     string
	SourceDir      string
	ImageOutputDir string
}

func New(sourceFile, sourceDir, imageOutputDir string) *Processor {
	return &Processor{
		SourceFile:     sourceFile,
		SourceDir:      sourceDir,
		ImageOutputDir: imageOutputDir,
	}
}

func (p *Processor) Process() error {
	if p.SourceFile != "" {
		return p.processFile(p.SourceFile)
	}
	return p.processDirectory(p.SourceDir)
}

func (p *Processor) processDirectory(dir string) error {
	fmt.Printf("Processing directory: %s\n", dir)
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".markdown")) {
			return p.processFile(path)
		}
		return nil
	})
}

func (p *Processor) processFile(filePath string) error {
	fmt.Printf("Processing file: %s\n", filePath)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	// Determine image output directory
	imgDir := p.determineImageDir(filePath)
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		return fmt.Errorf("failed to create image directory %s: %v", imgDir, err)
	}

	// Find all image links
	imgRegex := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	matches := imgRegex.FindAllStringSubmatch(string(content), -1)

	fmt.Printf("Found %d images in file %s\n", len(matches), filePath)

	newContent := string(content)
	for _, match := range matches {
		imgAlt := match[1]
		imgURL := match[2]

		// Replace image URL starting with "//" to "https://"
		if strings.HasPrefix(imgURL, "//") {
			imgURL = strings.Replace(imgURL, "//", "https://", 1)
		}
		// Skip local images
		if !strings.HasPrefix(imgURL, "http://") && !strings.HasPrefix(imgURL, "https://") {
			continue
		}

		// Download and save image
		localPath, err := p.downloadImage(imgURL, imgDir)
		if err != nil {
			fmt.Printf("Warning: Failed to download image %s: %v\n", imgURL, err)
			continue
		}

		// Calculate relative path
		relPath, err := filepath.Rel(filepath.Dir(filePath), localPath)
		if err != nil {
			fmt.Printf("Warning: Failed to calculate relative path: %v\n", err)
			continue
		}

		// Replace image link
		oldLink := fmt.Sprintf("![%s](%s)", match[1], match[2])
		newLink := fmt.Sprintf("![%s](%s)", imgAlt, relPath)
		newContent = strings.Replace(newContent, oldLink, newLink, 1)
	}

	// Write back to file
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %v", filePath, err)
	}

	return nil
}

func (p *Processor) determineImageDir(filePath string) string {
	if p.ImageOutputDir != "" {
		return p.ImageOutputDir
	}
	if p.SourceDir != "" {
		return filepath.Join(p.SourceDir, "images")
	}
	return filepath.Join(filepath.Dir(filePath), "images")
}

func (p *Processor) downloadImage(url string, destDir string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Get filename from URL or Content-Disposition
	filename := getFilenameFromURL(url, resp)

	// If no extension, try to get from Content-Type
	if filepath.Ext(filename) == "" {
		contentType := resp.Header.Get("Content-Type")
		ext := getExtensionFromContentType(contentType)
		if ext != "" {
			filename += ext
		}
	}

	// Ensure filename is unique
	hash := md5.New()
	io.WriteString(hash, url)
	urlHash := fmt.Sprintf("%x", hash.Sum(nil))[:8]

	ext := filepath.Ext(filename)
	basename := strings.TrimSuffix(filename, ext)
	filename = fmt.Sprintf("%s_%s%s", basename, urlHash, ext)

	localPath := filepath.Join(destDir, filename)

	// Create target file
	out, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Printf("Downloaded image to: %s\n", localPath)
	return localPath, nil
}

func getFilenameFromURL(url string, resp *http.Response) string {
	// First try to get from Content-Disposition
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if strings.Contains(cd, "filename=") {
			parts := strings.Split(cd, "filename=")
			if len(parts) > 1 {
				filename := strings.Trim(parts[1], `"'`)
				if filename != "" {
					return filename
				}
			}
		}
	}

	// Get from URL path
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		// Remove URL parameters
		if idx := strings.Index(filename, "?"); idx != -1 {
			filename = filename[:idx]
		}
		// Remove trailing "@" character
		if idx := strings.LastIndex(filename, "@"); idx != -1 {
			if idx > strings.LastIndex(filename, ".") {
				filename = filename[:idx]
			}
		}
		if filename != "" {
			return filename
		}
	}

	// Use default name
	return "image"
}

func getExtensionFromContentType(contentType string) string {
	switch contentType {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}
