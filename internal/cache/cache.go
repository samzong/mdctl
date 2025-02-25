package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CacheItem represents a single cached file information
type CacheItem struct {
	LocalPath  string    `json:"local_path"`
	RemotePath string    `json:"remote_path"`
	URL        string    `json:"url"`
	Hash       string    `json:"hash"`
	UploadTime time.Time `json:"upload_time"`
}

// Cache manages information about uploaded files
type Cache struct {
	Items    map[string]CacheItem `json:"items"`
	Version  string               `json:"version"`
	CacheDir string               `json:"cache_dir,omitempty"`
	mutex    sync.RWMutex
}

// New creates a new cache instance
func New(cacheDir string) *Cache {
	if cacheDir == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			cacheDir = filepath.Join(homeDir, ".cache", "mdctl")
		} else {
			// Fallback to temp directory
			cacheDir = filepath.Join(os.TempDir(), "mdctl-cache")
		}
	}

	return &Cache{
		Items:    make(map[string]CacheItem),
		Version:  "1.0",
		CacheDir: cacheDir,
	}
}

// saveWithoutLock writes cache to disk without acquiring the lock
// This should only be called from methods that already hold a lock
func (c *Cache) saveWithoutLock() error {
	// Ensure cache directory exists
	if err := os.MkdirAll(c.CacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	cacheFile := filepath.Join(c.CacheDir, "upload-cache.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %v", err)
	}

	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %v", err)
	}

	return nil
}

// Load reads cache from disk
func (c *Cache) Load() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Ensure cache directory exists
	if err := os.MkdirAll(c.CacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	cacheFile := filepath.Join(c.CacheDir, "upload-cache.json")
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		// Cache file doesn't exist yet, create a new one
		c.Items = make(map[string]CacheItem)
		return c.saveWithoutLock() // 使用无锁版本，避免死锁
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return fmt.Errorf("failed to read cache file: %v", err)
	}

	if err := json.Unmarshal(data, c); err != nil {
		// If cache is corrupt, start with a fresh one
		c.Items = make(map[string]CacheItem)
		return nil
	}

	return nil
}

// Save persists the cache to disk
func (c *Cache) Save() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.saveWithoutLock() // Use the lockless version to avoid deadlock
}

// AddItem adds or updates a cache item
func (c *Cache) AddItem(localPath, remotePath, url, hash string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.Items[localPath] = CacheItem{
		LocalPath:  localPath,
		RemotePath: remotePath,
		URL:        url,
		Hash:       hash,
		UploadTime: time.Now(),
	}
}

// GetItem retrieves a cache item by local path
func (c *Cache) GetItem(localPath string) (CacheItem, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.Items[localPath]
	return item, exists
}

// HasItemWithHash checks if an item with the same hash exists
func (c *Cache) HasItemWithHash(hash string) (CacheItem, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, item := range c.Items {
		if item.Hash == hash {
			return item, true
		}
	}
	return CacheItem{}, false
}

// RemoveItem removes an item from the cache
func (c *Cache) RemoveItem(localPath string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.Items, localPath)
}
