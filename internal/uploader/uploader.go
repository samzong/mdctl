package uploader

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/samzong/mdctl/internal/cache"
	"github.com/samzong/mdctl/internal/config"
	"github.com/samzong/mdctl/internal/storage"
)

// FileStats holds statistics about processed files
type FileStats struct {
	TotalFiles     int
	ProcessedFiles int
	UploadedImages int
	SkippedImages  int
	FailedImages   int
	ChangedFiles   int
}

// ConflictPolicy defines how to handle naming conflicts
type ConflictPolicy string

const (
	// ConflictPolicyRename adds a unique suffix to the filename
	ConflictPolicyRename ConflictPolicy = "rename"
	// ConflictPolicyVersion adds a version number to the filename
	ConflictPolicyVersion ConflictPolicy = "version"
	// ConflictPolicyOverwrite replaces the existing file
	ConflictPolicyOverwrite ConflictPolicy = "overwrite"
)

// UploaderConfig holds configuration for the uploader
type UploaderConfig struct {
	SourceFile     string
	SourceDir      string
	Provider       string
	Bucket         string
	CustomDomain   string
	PathPrefix     string
	DryRun         bool
	Concurrency    int
	ForceUpload    bool
	SkipVerify     bool
	CACertPath     string
	ConflictPolicy ConflictPolicy
	CacheDir       string
	FileExtensions []string
}

// Uploader handles uploading images and rewriting markdown
type Uploader struct {
	Config         UploaderConfig
	provider       storage.Provider
	stats          FileStats
	cache          *cache.Cache
	workerWg       sync.WaitGroup
	taskChan       chan uploadTask
	resultChan     chan uploadResult
	errorChan      chan error
	doneProcessing bool
	pendingFiles   map[string][]pendingReplace // Map to track pending link updates for each file
	fileMutex      sync.Mutex                  // Mutex to protect pendingFiles
}

// Define a struct to track pending replacements
type pendingReplace struct {
	LocalPath  string
	OldLink    string
	ImgAlt     string
	RemotePath string // Add remote path to match during result processing
}

type uploadTask struct {
	LocalPath   string
	RemotePath  string
	Filename    string
	OriginalURL string
	AltText     string
}

type uploadResult struct {
	Task     uploadTask
	URL      string
	Uploaded bool
	Err      error
}

// New creates a new uploader
func New(uploaderConfig UploaderConfig) (*Uploader, error) {
	// Create cache
	cacheManager := cache.New(uploaderConfig.CacheDir)
	if err := cacheManager.Load(); err != nil {
		return nil, fmt.Errorf("failed to load cache: %v", err)
	}

	// Normalize conflict policy
	if uploaderConfig.ConflictPolicy == "" {
		uploaderConfig.ConflictPolicy = ConflictPolicyRename
	}

	// Set default concurrency
	if uploaderConfig.Concurrency <= 0 {
		uploaderConfig.Concurrency = 5
	}

	// Get config from file
	appConfig, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Get active cloud storage configuration
	var activeConfig config.CloudConfig

	// If Provider is specified in command line, user wants to use command line parameters
	// Otherwise, get active storage configuration from config
	if uploaderConfig.Provider == "" {
		// Check if --storage parameter was specified in command line
		for _, arg := range os.Args {
			if arg == "--storage" || arg == "-s" {
				// If --storage specified, it will be handled later in cmd/upload.go
				break
			}
		}

		// Get active cloud storage configuration
		activeConfig = appConfig.GetActiveCloudConfig("")
	} else {
		// When using command line parameters, start with default config and then override
		activeConfig = appConfig.GetActiveCloudConfig("")
	}

	// Override with command line arguments if provided
	if uploaderConfig.Bucket != "" {
		activeConfig.Bucket = uploaderConfig.Bucket
	}
	if uploaderConfig.CustomDomain != "" {
		activeConfig.CustomDomain = uploaderConfig.CustomDomain
	}
	if uploaderConfig.PathPrefix != "" {
		activeConfig.PathPrefix = uploaderConfig.PathPrefix
	}
	if uploaderConfig.SkipVerify {
		activeConfig.SkipVerify = true
	}
	if uploaderConfig.CACertPath != "" {
		activeConfig.CACertPath = uploaderConfig.CACertPath
	}
	if string(uploaderConfig.ConflictPolicy) != "" {
		activeConfig.ConflictPolicy = string(uploaderConfig.ConflictPolicy)
	}
	if uploaderConfig.Provider != "" {
		activeConfig.Provider = uploaderConfig.Provider
	}

	// Ensure Provider specified from command line takes precedence over Provider in config
	providerName := strings.ToLower(activeConfig.Provider)
	if providerName == "" {
		providerName = strings.ToLower(uploaderConfig.Provider)
	}

	// Get provider and configure it
	if providerName == "" {
		return nil, errors.New("provider must be specified")
	}

	provider, exists := storage.GetProvider(providerName)
	if !exists {
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	// Configure provider
	if err := provider.Configure(activeConfig); err != nil {
		return nil, fmt.Errorf("failed to configure provider: %v", err)
	}

	return &Uploader{
		Config:       uploaderConfig,
		provider:     provider,
		cache:        cacheManager,
		pendingFiles: make(map[string][]pendingReplace), // Initialize pendingFiles
	}, nil
}

// Process starts the upload process
func (u *Uploader) Process() (*FileStats, error) {
	// Initialize channels for worker pool
	u.taskChan = make(chan uploadTask, u.Config.Concurrency*2)
	u.resultChan = make(chan uploadResult, u.Config.Concurrency*2)
	u.errorChan = make(chan error, 10)

	// Start worker pool
	for i := 0; i < u.Config.Concurrency; i++ {
		u.workerWg.Add(1)
		go u.uploadWorker()
	}

	// Start result processor
	resultWg := sync.WaitGroup{}
	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		u.processResults()
	}()

	// Process files
	var err error
	if u.Config.SourceFile != "" {
		err = u.processFile(u.Config.SourceFile)
	} else if u.Config.SourceDir != "" {
		err = u.processDirectory(u.Config.SourceDir)
	} else {
		err = errors.New("either source file or source directory must be specified")
	}

	// Signal that all files have been processed
	u.doneProcessing = true
	close(u.taskChan)

	// Wait for all uploads to complete
	u.workerWg.Wait()
	close(u.resultChan)

	// Wait for result processor to complete
	resultWg.Wait()

	// Save cache
	if err := u.cache.Save(); err != nil {
		fmt.Printf("Warning: Failed to save cache: %v\n", err)
	}

	return &u.stats, err
}

// processDirectory processes all markdown files in a directory
func (u *Uploader) processDirectory(dir string) error {
	fmt.Printf("Processing directory: %s\n", dir)
	u.stats.TotalFiles = 0

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".markdown")) {
			u.stats.TotalFiles++
			return u.processFile(path)
		}
		return nil
	})
}

// processFile processes a single markdown file
func (u *Uploader) processFile(filePath string) error {
	fmt.Printf("Processing file: %s\n", filePath)
	u.stats.ProcessedFiles++

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	// Find all image links
	imgRegex := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	matches := imgRegex.FindAllStringSubmatch(string(content), -1)

	if len(matches) == 0 {
		fmt.Printf("No images found in file %s\n", filePath)
		return nil
	}

	fmt.Printf("Found %d images in file %s\n", len(matches), filePath)

	// Track changes to the file
	newContent := string(content)
	var contentChanged bool

	for _, match := range matches {
		imgAlt := match[1]
		imgURL := match[2]

		// Skip remote images
		if strings.HasPrefix(imgURL, "http://") || strings.HasPrefix(imgURL, "https://") || strings.HasPrefix(imgURL, "//") {
			continue
		}

		// Get absolute path for local image
		var imgPath string
		if filepath.IsAbs(imgURL) {
			imgPath = imgURL
		} else {
			// Resolve relative to the markdown file
			imgPath = filepath.Join(filepath.Dir(filePath), imgURL)
		}

		// Check if file exists
		if _, err := os.Stat(imgPath); os.IsNotExist(err) {
			fmt.Printf("Warning: Image does not exist: %s\n", imgPath)
			continue
		}

		// Calculate hash for the file
		hash, err := u.calculateFileHash(imgPath)
		if err != nil {
			fmt.Printf("Warning: Failed to calculate hash for %s: %v\n", imgPath, err)
			continue
		}

		// Check if file is already in cache
		if !u.Config.ForceUpload {
			if item, exists := u.cache.GetItem(imgPath); exists {
				// Use cached URL
				oldLink := fmt.Sprintf("![%s](%s)", imgAlt, imgURL)
				newLink := fmt.Sprintf("![%s](%s)", imgAlt, item.URL)
				if oldLink != newLink {
					newContent = strings.Replace(newContent, oldLink, newLink, 1)
					contentChanged = true
				}
				fmt.Printf("Using cached URL for image: %s → %s\n", imgPath, item.URL)
				u.stats.SkippedImages++
				continue
			}
		}

		// Generate remote path
		ext := filepath.Ext(imgPath)
		filename := filepath.Base(imgPath)
		nameWithoutExt := strings.TrimSuffix(filename, ext)

		// Clean filename
		nameWithoutExt = cleanFileName(nameWithoutExt)
		remotePath := fmt.Sprintf("images/%s_%s%s", nameWithoutExt, hash[:8], ext)

		// Record link replacement information
		oldLink := fmt.Sprintf("![%s](%s)", imgAlt, imgURL)
		u.fileMutex.Lock()
		u.pendingFiles[filePath] = append(u.pendingFiles[filePath], pendingReplace{
			LocalPath:  imgPath,
			OldLink:    oldLink,
			ImgAlt:     imgAlt,
			RemotePath: remotePath,
		})
		u.fileMutex.Unlock()

		// Add to upload queue
		u.taskChan <- uploadTask{
			LocalPath:  imgPath,
			RemotePath: remotePath,
			Filename:   filename,
		}
	}

	// 仅处理来自缓存的链接更新，非缓存的会在所有上传完成后处理
	if contentChanged && !u.Config.DryRun {
		if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %v", filePath, err)
		}
		u.stats.ChangedFiles++
	}

	return nil
}

// uploadWorker processes upload tasks
func (u *Uploader) uploadWorker() {
	defer u.workerWg.Done()

	for task := range u.taskChan {
		// Calculate hash for file
		hash, err := u.calculateFileHash(task.LocalPath)
		if err != nil {
			u.resultChan <- uploadResult{
				Task: task,
				Err:  fmt.Errorf("failed to calculate hash: %v", err),
			}
			continue
		}

		// Skip upload in dry run mode
		if u.Config.DryRun {
			u.resultChan <- uploadResult{
				Task:     task,
				URL:      u.provider.GetPublicURL(task.RemotePath),
				Uploaded: false,
			}
			continue
		}

		// Handle conflict according to policy
		remotePath := task.RemotePath
		exists, err := u.provider.ObjectExists(remotePath)
		if err != nil {
			u.resultChan <- uploadResult{
				Task: task,
				Err:  fmt.Errorf("failed to check if object exists: %v", err),
			}
			continue
		}

		if exists && !u.Config.ForceUpload {
			// Check if hash matches
			hashMatches, err := u.provider.CompareHash(remotePath, hash)
			if err == nil && hashMatches {
				// File already exists with same content, just return the URL
				u.resultChan <- uploadResult{
					Task:     task,
					URL:      u.provider.GetPublicURL(remotePath),
					Uploaded: false,
				}
				continue
			}

			// Handle conflict based on policy
			switch u.Config.ConflictPolicy {
			case ConflictPolicyRename:
				// Generate new name with timestamp
				ext := filepath.Ext(remotePath)
				base := strings.TrimSuffix(remotePath, ext)
				timestamp := time.Now().UnixNano()
				remotePath = fmt.Sprintf("%s_%d%s", base, timestamp, ext)
			case ConflictPolicyVersion:
				// Find next available version number
				ext := filepath.Ext(remotePath)
				base := strings.TrimSuffix(remotePath, ext)
				version := 1
				for {
					newPath := fmt.Sprintf("%s_v%d%s", base, version, ext)
					exists, _ := u.provider.ObjectExists(newPath)
					if !exists {
						remotePath = newPath
						break
					}
					version++
				}
			case ConflictPolicyOverwrite:
				// Keep the same path, will overwrite
			}
		}

		// Upload file
		metadata := map[string]string{
			"Hash":       hash,
			"Original":   task.Filename,
			"UploadTime": time.Now().Format(time.RFC3339),
		}

		url, err := u.provider.Upload(task.LocalPath, remotePath, metadata)
		if err != nil {
			u.resultChan <- uploadResult{
				Task: task,
				Err:  fmt.Errorf("failed to upload file: %v", err),
			}
			continue
		}

		u.resultChan <- uploadResult{
			Task:     task,
			URL:      url,
			Uploaded: true,
		}
	}
}

// processResults handles results from the upload workers
func (u *Uploader) processResults() {
	uploadedURLs := make(map[string]string)

	for result := range u.resultChan {
		if result.Err != nil {
			fmt.Printf("Error uploading %s: %v\n", result.Task.LocalPath, result.Err)
			u.stats.FailedImages++
			continue
		}

		// Store URL for later use in content replacement
		uploadedURLs[result.Task.LocalPath] = result.URL

		if result.Uploaded {
			fmt.Printf("Uploaded image: %s → %s\n", result.Task.LocalPath, result.URL)
			u.stats.UploadedImages++

			// Add to cache
			hash, _ := u.calculateFileHash(result.Task.LocalPath)
			u.cache.AddItem(result.Task.LocalPath, result.Task.RemotePath, result.URL, hash)
		} else {
			fmt.Printf("Skipped upload (already exists): %s → %s\n", result.Task.LocalPath, result.URL)
			u.stats.SkippedImages++
		}
	}

	// After all uploads complete, update file contents
	u.fileMutex.Lock()
	defer u.fileMutex.Unlock()

	for filePath, replaces := range u.pendingFiles {
		if len(replaces) == 0 {
			continue
		}

		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading file %s for update: %v", filePath, err)
			continue
		}

		// Apply all replacements
		newContent := string(content)
		contentChanged := false

		for _, replace := range replaces {
			if newURL, exists := uploadedURLs[replace.LocalPath]; exists {
				newLink := fmt.Sprintf("![%s](%s)", replace.ImgAlt, newURL)
				oldNewContent := newContent
				newContent = strings.Replace(newContent, replace.OldLink, newLink, 1)
				if oldNewContent != newContent {
					contentChanged = true
					fmt.Printf("Updated link in %s: %s -> %s\n", filePath, replace.OldLink, newLink)
				}
			}
		}

		// Save updated file
		if contentChanged && !u.Config.DryRun {
			if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
				log.Printf("Error writing updated file: %v\n", err)
			} else {
				u.stats.ChangedFiles++
			}
		}
	}
}

// calculateFileHash computes MD5 hash of a file
func (u *Uploader) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// cleanFileName removes special characters from filename
func cleanFileName(name string) string {
	// Replace spaces and special characters with underscores
	re := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	name = re.ReplaceAllString(name, "_")

	// Collapse multiple underscores
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}

	// Trim underscores from beginning and end
	name = strings.Trim(name, "_")

	// Limit length
	if len(name) > 50 {
		name = name[:50]
	}

	return name
}
