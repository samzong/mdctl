package storage

import (
	"github.com/samzong/mdctl/internal/config"
)

// Provider defines the interface for storage providers
type Provider interface {
	// Upload uploads a file to cloud storage
	Upload(localPath, remotePath string, metadata map[string]string) (string, error)

	// Configure sets up the provider with the given configuration
	Configure(config config.CloudConfig) error

	// GetPublicURL returns the public URL for a remote path
	GetPublicURL(remotePath string) string

	// ObjectExists checks if an object exists in the storage
	ObjectExists(remotePath string) (bool, error)

	// CompareHash compares a local hash with a remote object's hash
	CompareHash(remotePath, localHash string) (bool, error)

	// SetObjectMetadata sets metadata for an object
	SetObjectMetadata(remotePath string, metadata map[string]string) error

	// GetObjectMetadata retrieves metadata for an object
	GetObjectMetadata(remotePath string) (map[string]string, error)
}

// ProviderFactory is a function that creates a new storage provider
type ProviderFactory func() Provider

var providers = make(map[string]ProviderFactory)

// RegisterProvider registers a storage provider factory
func RegisterProvider(name string, factory ProviderFactory) {
	providers[name] = factory
}

// GetProvider returns a storage provider by name
func GetProvider(name string) (Provider, bool) {
	factory, exists := providers[name]
	if !exists {
		return nil, false
	}
	return factory(), true
}

// ListProviders returns a list of available provider names
func ListProviders() []string {
	var names []string
	for name := range providers {
		names = append(names, name)
	}
	return names
}
