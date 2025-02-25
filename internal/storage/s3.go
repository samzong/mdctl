package storage

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/samzong/mdctl/internal/config"
)

// init registers the S3 provider
func init() {
	RegisterProvider("s3", func() Provider { return NewS3Provider() })
	RegisterProvider("r2", func() Provider { return NewS3Provider() })    // Cloudflare R2 (S3 compatible)
	RegisterProvider("minio", func() Provider { return NewS3Provider() }) // MinIO (S3 compatible)
}

// S3Provider implements the Provider interface for S3-compatible storage services
type S3Provider struct {
	client       *s3.S3
	bucket       string
	region       string
	endpoint     string
	customDomain string
	pathPrefix   string
	accountID    string // Add accountID field for R2
}

// NewS3Provider creates a new S3 provider
func NewS3Provider() *S3Provider {
	return &S3Provider{}
}

// Configure sets up the S3 provider with the given configuration
func (p *S3Provider) Configure(cfg config.CloudConfig) error {
	// Set provider configuration
	p.bucket = cfg.Bucket
	p.region = cfg.Region
	p.endpoint = cfg.Endpoint
	p.customDomain = cfg.CustomDomain
	p.pathPrefix = cfg.PathPrefix

	// Set accountID, prioritize AccountID from configuration
	p.accountID = cfg.AccountID

	// Try to extract accountID from endpoint URL: https://<account_id>.r2.cloudflarestorage.com
	if p.accountID == "" && strings.Contains(p.endpoint, "r2.cloudflarestorage.com") {
		p.accountID = extractR2AccountID(p.endpoint)
	}

	// If it's R2 but accountID not set, log a warning
	if strings.ToLower(cfg.Provider) == "r2" && p.accountID == "" {
		log.Printf("Warning: R2 account ID not set. r2.dev public URLs cannot be generated.")
	}

	// Create AWS configuration
	awsConfig := &aws.Config{
		Region:      aws.String(cfg.Region),
		Credentials: credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
	}

	// Set custom endpoint if provided
	if cfg.Endpoint != "" {
		awsConfig.Endpoint = aws.String(cfg.Endpoint)
		// Use path-style addressing for custom endpoints
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}

	// Configure TLS settings
	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}

	// Set up custom transport if needed
	if cfg.SkipVerify || cfg.CACertPath != "" {
		// Start with the default transport
		transport := &http.Transport{
			TLSHandshakeTimeout: 10 * time.Second,
		}

		// Configure TLS
		tlsConfig := &tls.Config{}

		// Skip certificate verification if requested
		if cfg.SkipVerify {
			tlsConfig.InsecureSkipVerify = true
		}

		// Load custom CA certificate if provided
		if cfg.CACertPath != "" {
			rootCAs, _ := x509.SystemCertPool()
			if rootCAs == nil {
				rootCAs = x509.NewCertPool()
			}

			certs, err := os.ReadFile(cfg.CACertPath)
			if err != nil {
				return fmt.Errorf("failed to read CA cert: %v", err)
			}

			if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
				return fmt.Errorf("failed to append CA cert")
			}

			tlsConfig.RootCAs = rootCAs
		}

		transport.TLSClientConfig = tlsConfig
		httpClient.Transport = transport
	}

	awsConfig.HTTPClient = httpClient

	// Create session
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %v", err)
	}

	// Create S3 client
	p.client = s3.New(sess)
	return nil
}

// Upload uploads a file to S3 storage
func (p *S3Provider) Upload(localPath, remotePath string, metadata map[string]string) (string, error) {
	// Ensure remotePath starts with prefix if set
	if p.pathPrefix != "" && !strings.HasPrefix(remotePath, p.pathPrefix) {
		remotePath = filepath.Join(p.pathPrefix, remotePath)
	}

	// Read file
	data, err := os.ReadFile(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Determine content type
	contentType := getContentType(localPath)

	// Prepare metadata
	s3Metadata := make(map[string]*string)
	for key, value := range metadata {
		s3Metadata[key] = aws.String(value)
	}

	// Upload to S3
	_, err = p.client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(p.bucket),
		Key:           aws.String(remotePath),
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(int64(len(data))),
		ContentType:   aws.String(contentType),
		Metadata:      s3Metadata,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}

	// Return public URL
	return p.GetPublicURL(remotePath), nil
}

// GetPublicURL returns the public URL for a remote path
func (p *S3Provider) GetPublicURL(remotePath string) string {
	if p.customDomain != "" {
		// Use custom domain
		return fmt.Sprintf("https://%s/%s", p.customDomain, remotePath)
	}

	// Generate r2.dev URL for Cloudflare R2
	if p.endpoint != "" && strings.Contains(p.endpoint, "r2.dev") {
		// First check if accountID is set
		accountID := p.accountID
		if accountID == "" {
			// Try to extract from endpoint again
			accountID = extractR2AccountID(p.endpoint)
		}

		if accountID != "" {
			// Use r2.dev public domain: https://pub-<bucket-name>.<account-id>.r2.dev
			return fmt.Sprintf("https://pub-%s.%s.r2.dev/%s", p.bucket, accountID, remotePath)
		}
	}

	// Use S3 URL format
	if p.endpoint != "" {
		// For custom endpoint (like MinIO or other S3-compatible)
		endpoint := strings.TrimRight(p.endpoint, "/")
		if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
			return fmt.Sprintf("%s/%s/%s", endpoint, p.bucket, remotePath)
		}
		return fmt.Sprintf("https://%s/%s/%s", endpoint, p.bucket, remotePath)
	}

	// Standard AWS S3
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", p.bucket, p.region, remotePath)
}

// ObjectExists checks if an object exists in the bucket
func (p *S3Provider) ObjectExists(remotePath string) (bool, error) {
	// Ensure remotePath starts with prefix if set
	if p.pathPrefix != "" && !strings.HasPrefix(remotePath, p.pathPrefix) {
		remotePath = filepath.Join(p.pathPrefix, remotePath)
	}

	_, err := p.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(remotePath),
	})
	if err != nil {
		// Check if error means object doesn't exist
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// CompareHash compares a local hash with a remote object's hash
func (p *S3Provider) CompareHash(remotePath, localHash string) (bool, error) {
	// Ensure remotePath starts with prefix if set
	if p.pathPrefix != "" && !strings.HasPrefix(remotePath, p.pathPrefix) {
		remotePath = filepath.Join(p.pathPrefix, remotePath)
	}

	headOutput, err := p.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(remotePath),
	})
	if err != nil {
		return false, err
	}

	// Check for hash in metadata
	if headOutput.Metadata != nil {
		if hash, ok := headOutput.Metadata["Hash"]; ok && hash != nil {
			return *hash == localHash, nil
		}
	}

	// Try ETag (might be MD5 in simple cases)
	if headOutput.ETag != nil {
		etag := *headOutput.ETag
		// Remove quotes from ETag
		etag = strings.Trim(etag, "\"")
		return etag == localHash, nil
	}

	return false, nil
}

// SetObjectMetadata sets metadata for an object
func (p *S3Provider) SetObjectMetadata(remotePath string, metadata map[string]string) error {
	// Ensure remotePath starts with prefix if set
	if p.pathPrefix != "" && !strings.HasPrefix(remotePath, p.pathPrefix) {
		remotePath = filepath.Join(p.pathPrefix, remotePath)
	}

	// Get the current object
	getObjectOutput, err := p.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(remotePath),
	})
	if err != nil {
		return err
	}

	// Read the object data
	data, err := io.ReadAll(getObjectOutput.Body)
	if err != nil {
		return err
	}
	defer getObjectOutput.Body.Close()

	// Prepare metadata
	s3Metadata := make(map[string]*string)
	for key, value := range metadata {
		s3Metadata[key] = aws.String(value)
	}

	// Upload the object with new metadata
	_, err = p.client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(p.bucket),
		Key:           aws.String(remotePath),
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(int64(len(data))),
		ContentType:   getObjectOutput.ContentType,
		Metadata:      s3Metadata,
	})

	return err
}

// GetObjectMetadata retrieves metadata for an object
func (p *S3Provider) GetObjectMetadata(remotePath string) (map[string]string, error) {
	// Ensure remotePath starts with prefix if set
	if p.pathPrefix != "" && !strings.HasPrefix(remotePath, p.pathPrefix) {
		remotePath = filepath.Join(p.pathPrefix, remotePath)
	}

	headOutput, err := p.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(remotePath),
	})
	if err != nil {
		return nil, err
	}

	metadata := make(map[string]string)
	for key, value := range headOutput.Metadata {
		if value != nil {
			metadata[key] = *value
		}
	}

	return metadata, nil
}

// Helper function to determine content type from file extension
func getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".bmp":
		return "image/bmp"
	case ".tiff", ".tif":
		return "image/tiff"
	default:
		return "application/octet-stream"
	}
}

// Add extractR2AccountID function
func extractR2AccountID(endpoint string) string {
	// Remove prefix http:// or https://
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	// Parse R2 endpoint format: <account_id>.r2.cloudflarestorage.com
	parts := strings.Split(endpoint, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
