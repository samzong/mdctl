# Design Document: Image Upload Feature for mdctl

## Overview

Add a new feature to mdctl that uploads local images in markdown files to cloud storage services (S3-compatible APIs like Cloudflare R2, AWS S3, etc.) and rewrites the URLs in the markdown content.

## Goals

1. Upload local images to cloud storage services
2. Support multiple storage providers with S3-compatible APIs
3. Rewrite image URLs in markdown files to point to the cloud storage
4. Maintain the existing design patterns and code structure
5. Implement idempotent operations with content verification
6. Support concurrent uploads for performance optimization
7. Handle custom SSL certificates for various cloud providers

## Architecture

Following the existing architecture pattern of mdctl, the upload feature will be implemented with these components:

### 1. Command Layer (`cmd/upload.go`)

- Define CLI parameters:
  - Source file/directory (`-f/--file` or `-d/--dir`)
  - Cloud provider (`-p/--provider`)
  - Bucket name (`-b/--bucket`)
  - Custom domain (optional, `-c/--custom-domain`)
  - Path prefix (optional, `--prefix`)
  - File extensions to include (optional, `--include`)
  - Dry run mode (optional, `--dry-run`)
  - Concurrency level (optional, `--concurrency`)
  - Force upload (optional, `-F/--force`)
  - Skip SSL verification (optional, `--skip-verify`)
  - CA certificate path (optional, `--ca-cert`)
  - Conflict policy (optional, `--conflict=rename|version|overwrite`)
  - Cache directory (optional, `--cache-dir`)

- Validate input parameters
- Create and configure uploader component
- Add to the "core" command group alongside download and translate

### 2. Uploader Module (`internal/uploader/uploader.go`)

- Core business logic for uploading files
- Methods for:
  - Processing single files or directories recursively
  - Identifying local images in markdown
  - Uploading files to cloud storage
  - Rewriting URLs in markdown content
  - Generating appropriate cloud storage paths
  - Managing the worker pool for concurrent uploads
  - Tracking upload progress with statistics
  - Calculating and verifying content hashes
  - Handling conflict resolution
  - Managing the local cache of uploaded files

### 3. Storage Provider Interface (`internal/storage/provider.go`)

- Define a provider interface with methods:
  - `Upload(localPath, remotePath string, metadata map[string]string) (url string, err error)`
  - `Configure(config CloudConfig) error`
  - `GetPublicURL(remotePath string) string`
  - `ObjectExists(remotePath string) (bool, error)`
  - `CompareHash(remotePath, localHash string) (bool, error)`
  - `SetObjectMetadata(remotePath string, metadata map[string]string) error`
  - `GetObjectMetadata(remotePath string) (map[string]string, error)`

### 4. Storage Provider Implementations

- S3-compatible provider (`internal/storage/s3.go`):
  - Implementation for AWS S3, Cloudflare R2, Minio, etc.
  - Configure region, endpoint, credentials
  - Handle authentication and uploads
  - Support custom certificates and SSL verification options
  - Implement content verification with ETag/MD5 hash comparison
  - Support object tagging for metadata

### 5. Cache Management (`internal/cache/cache.go`)

- Maintain record of uploaded files with their hash values
- Cache structure with file path, remote URL, and hash
- Support for serializing/deserializing cache to disk
- Methods for lookup, update, and verification

### 6. Configuration Extensions (`internal/config/config.go`)

Add new configuration fields:
```go
type CloudConfig struct {
    Provider       string            `json:"provider"`
    Region         string            `json:"region"`
    Endpoint       string            `json:"endpoint"`
    AccessKey      string            `json:"access_key"`
    SecretKey      string            `json:"secret_key"`
    Bucket         string            `json:"bucket"`
    CustomDomain   string            `json:"custom_domain,omitempty"`
    PathPrefix     string            `json:"path_prefix,omitempty"`
    ProviderOpts   map[string]string `json:"provider_opts,omitempty"`
    Concurrency    int               `json:"concurrency"`
    SkipVerify     bool              `json:"skip_verify"`
    CACertPath     string            `json:"ca_cert_path,omitempty"`
    ConflictPolicy string            `json:"conflict_policy"`
    CacheDir       string            `json:"cache_dir,omitempty"`
}

// Add to Config struct
type Config struct {
    // Existing fields...
    CloudStorage CloudConfig `json:"cloud_storage"`
}
```

## Implementation Plan

1. Add cloud storage config section to config.go
2. Implement cache management module
3. Create storage provider interface 
4. Implement S3-compatible provider with SSL handling
5. Create worker pool for concurrent uploads
6. Create uploader module implementation with verification logic
7. Implement idempotency and conflict resolution strategies  
8. Add upload command to cmd package
9. Create comprehensive tests
10. Update help text and documentation
11. Add sample usage to README

## Command Usage Examples

```bash
# Upload images from a single file
mdctl upload -f path/to/file.md -p s3 -b my-bucket

# Upload images from a directory
mdctl upload -d path/to/dir -p r2 -b my-images --prefix blog/

# Use with a custom domain
mdctl upload -f post.md -p s3 -b media-bucket -c assets.example.com

# Use custom concurrency setting
mdctl upload -f blog-post.md -p s3 -b my-bucket --concurrency 10

# Force upload (bypass hash verification)
mdctl upload -f readme.md -p r2 -b my-images -F

# Specify conflict resolution strategy
mdctl upload -d docs/ -p s3 -b media --conflict=version

# Use custom SSL certificate
mdctl upload -f doc.md -p s3 -b media --ca-cert /path/to/cert.pem

# Skip SSL verification for self-signed certificates
mdctl upload -f doc.md -p minio -b local --skip-verify

# Configure cloud provider
mdctl config set -k cloud_storage.provider -v "r2"
mdctl config set -k cloud_storage.endpoint -v "https://xxxx.r2.cloudflarestorage.com"
mdctl config set -k cloud_storage.access_key -v "YOUR_ACCESS_KEY"
mdctl config set -k cloud_storage.secret_key -v "YOUR_SECRET_KEY"
mdctl config set -k cloud_storage.bucket -v "my-images"
mdctl config set -k cloud_storage.concurrency -v 5
mdctl config set -k cloud_storage.conflict_policy -v "rename"
```

## Technical Considerations

1. **S3 SDK**: Use the AWS SDK for Go to interact with S3-compatible APIs
2. **Image Processing**: Optional compression/resizing before upload
3. **Error Handling**: Provide detailed error messages for failed uploads
4. **URL Generation**:
   - Support both direct S3 URLs or custom domain URLs
   - Handle path prefixing correctly
5. **Idempotency & Verification**:
   - Calculate content hashes (MD5/SHA) for each file
   - Store metadata in the object tags for verification
   - Skip uploads for identical content (check hash before upload)
   - Optional force upload flag to override verification
   - Maintain a local cache of uploaded files with their hashes
6. **Concurrency & Reliability**:
   - Implement worker pool for parallel uploads
   - Configurable concurrency level (default: 5)
   - Progress tracking for concurrent operations
   - Built-in retry mechanism for failed uploads (hardcoded 3 retry attempts)
   - Exponential backoff between retries (starting at 1s, doubling each retry)
   - Standard timeout for upload operations
7. **SSL/Certificate Handling**:
   - Support custom CA certificates
   - Option to skip verification for self-signed certificates
   - Configurable TLS settings per provider
8. **Conflict Resolution**:
   - Strategies for handling name collisions (rename, version, overwrite)
   - Option to preserve original filenames or use hashed names
9. **Incremental Uploads**:
   - Track already uploaded files to avoid redundant operations
   - Support for resuming interrupted batch uploads

## Testing Strategy

1. Unit tests for URL parsing and rewriting
2. Mocked storage provider for testing upload logic
3. Verification tests for hash calculation and comparison
4. Concurrency tests to ensure worker pool functions correctly
5. SSL/TLS configuration tests with mock certificates
6. Cache management tests for serialization/deserialization
7. Conflict resolution strategy tests
8. Integration tests with a local MinIO server
9. End-to-end tests with actual markdown files
10. Idempotency tests to verify repeated executions
11. Performance benchmarks for concurrent uploads