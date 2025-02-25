package cmd

import (
	"fmt"
	"strings"

	"github.com/samzong/mdctl/internal/config"
	"github.com/samzong/mdctl/internal/uploader"
	"github.com/spf13/cobra"
)

var (
	// Upload command flags
	uploadSourceFile     string
	uploadSourceDir      string
	uploadProvider       string
	uploadBucket         string
	uploadCustomDomain   string
	uploadPathPrefix     string
	uploadDryRun         bool
	uploadConcurrency    int
	uploadForceUpload    bool
	uploadSkipVerify     bool
	uploadCACertPath     string
	uploadConflictPolicy string
	uploadCacheDir       string
	uploadIncludeExts    string
	uploadStorageName    string

	uploadCmd = &cobra.Command{
		Use:   "upload",
		Short: "Upload local images in markdown files to cloud storage",
		Long: `Upload local images in markdown files to cloud storage and rewrite URLs.
Supports multiple cloud storage providers with S3-compatible APIs.

Examples:
  mdctl upload -d docs/
  mdctl upload -f post.md
  mdctl upload -f post.md --storage my-s3`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if uploadSourceFile == "" && uploadSourceDir == "" {
				return fmt.Errorf("either source file (-f) or source directory (-d) must be specified")
			}
			if uploadSourceFile != "" && uploadSourceDir != "" {
				return fmt.Errorf("cannot specify both source file (-f) and source directory (-d)")
			}

			// Load configuration file first
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %v", err)
			}

			// Get active cloud storage configuration
			cloudConfig := cfg.GetActiveCloudConfig(uploadStorageName)

			// Command line parameters take precedence over configuration
			if uploadProvider == "" {
				uploadProvider = cloudConfig.Provider
			}

			if uploadBucket == "" {
				uploadBucket = cloudConfig.Bucket
			}

			// Check for empty values after using configuration file values
			if uploadProvider == "" {
				return fmt.Errorf("provider (-p) must be specified or set in configuration file")
			}

			if uploadBucket == "" {
				return fmt.Errorf("bucket (-b) must be specified or set in configuration file")
			}

			// Set default region for S3-compatible services
			// If region is not set or empty, set default region
			if cloudConfig.Region == "" {
				switch strings.ToLower(uploadProvider) {
				case "s3":
					// For AWS S3, default to us-east-1
					cloudConfig.Region = "us-east-1"
				case "r2", "minio", "b2":
					// For S3-compatible services, region can be any value but must be provided
					cloudConfig.Region = "auto"
				}
			}

			// If not specified in command line, get other configuration parameters
			if uploadCustomDomain == "" {
				uploadCustomDomain = cloudConfig.CustomDomain
			}

			if uploadPathPrefix == "" {
				uploadPathPrefix = cloudConfig.PathPrefix
			}

			if uploadConcurrency == 5 && cloudConfig.Concurrency != 0 { // 5 is default value
				uploadConcurrency = cloudConfig.Concurrency
			}

			if uploadCACertPath == "" {
				uploadCACertPath = cloudConfig.CACertPath
			}

			if uploadSkipVerify == false && cloudConfig.SkipVerify {
				uploadSkipVerify = true
			}

			if uploadConflictPolicy == "rename" && cloudConfig.ConflictPolicy != "" {
				uploadConflictPolicy = cloudConfig.ConflictPolicy
			}

			if uploadCacheDir == "" {
				uploadCacheDir = cloudConfig.CacheDir
			}

			// Parse include extensions
			var exts []string
			if uploadIncludeExts != "" {
				exts = strings.Split(uploadIncludeExts, ",")
				for i, ext := range exts {
					exts[i] = strings.TrimSpace(ext)
				}
			}

			// Validate conflict policy
			var conflictPolicy uploader.ConflictPolicy
			switch strings.ToLower(uploadConflictPolicy) {
			case "rename":
				conflictPolicy = uploader.ConflictPolicyRename
			case "version":
				conflictPolicy = uploader.ConflictPolicyVersion
			case "overwrite":
				conflictPolicy = uploader.ConflictPolicyOverwrite
			case "":
				conflictPolicy = uploader.ConflictPolicyRename // Default
			default:
				return fmt.Errorf("invalid conflict policy: %s (must be rename, version, or overwrite)", uploadConflictPolicy)
			}

			// For R2, use account ID from configuration file
			if strings.ToLower(uploadProvider) == "r2" && cloudConfig.AccountID == "" {
				fmt.Printf("Note: R2 account ID not found in configuration, please set account_id in config file if you want to use r2.dev public URLs\n")
			}

			// Create uploader
			up, err := uploader.New(uploader.UploaderConfig{
				SourceFile:     uploadSourceFile,
				SourceDir:      uploadSourceDir,
				Provider:       uploadProvider,
				Bucket:         uploadBucket,
				CustomDomain:   uploadCustomDomain,
				PathPrefix:     uploadPathPrefix,
				DryRun:         uploadDryRun,
				Concurrency:    uploadConcurrency,
				ForceUpload:    uploadForceUpload,
				SkipVerify:     uploadSkipVerify,
				CACertPath:     uploadCACertPath,
				ConflictPolicy: conflictPolicy,
				CacheDir:       uploadCacheDir,
				FileExtensions: exts,
			})
			if err != nil {
				return fmt.Errorf("failed to create uploader: %v", err)
			}

			// Process files
			stats, err := up.Process()
			if err != nil {
				return fmt.Errorf("failed to process files: %v", err)
			}

			// Print statistics
			fmt.Printf("\nUpload Statistics:\n")
			fmt.Printf("  Total Files Processed: %d\n", stats.ProcessedFiles)
			fmt.Printf("  Images Uploaded: %d\n", stats.UploadedImages)
			fmt.Printf("  Images Skipped: %d\n", stats.SkippedImages)
			fmt.Printf("  Failed Uploads: %d\n", stats.FailedImages)
			fmt.Printf("  Files Changed: %d\n", stats.ChangedFiles)

			return nil
		},
	}
)

func init() {
	// Add flags
	uploadCmd.Flags().StringVarP(&uploadSourceFile, "file", "f", "", "Source markdown file to process")
	uploadCmd.Flags().StringVarP(&uploadSourceDir, "dir", "d", "", "Source directory containing markdown files to process")
	uploadCmd.Flags().StringVarP(&uploadProvider, "provider", "p", "", "Cloud storage provider (s3, r2, minio)")
	uploadCmd.Flags().StringVarP(&uploadBucket, "bucket", "b", "", "Cloud storage bucket name")
	uploadCmd.Flags().StringVarP(&uploadCustomDomain, "custom-domain", "c", "", "Custom domain for generated URLs")
	uploadCmd.Flags().StringVar(&uploadPathPrefix, "prefix", "", "Path prefix for uploaded files")
	uploadCmd.Flags().BoolVar(&uploadDryRun, "dry-run", false, "Preview changes without uploading")
	uploadCmd.Flags().IntVar(&uploadConcurrency, "concurrency", 5, "Number of concurrent uploads")
	uploadCmd.Flags().BoolVarP(&uploadForceUpload, "force", "F", false, "Force upload even if file exists")
	uploadCmd.Flags().BoolVar(&uploadSkipVerify, "skip-verify", false, "Skip SSL verification")
	uploadCmd.Flags().StringVar(&uploadCACertPath, "ca-cert", "", "Path to CA certificate")
	uploadCmd.Flags().StringVar(&uploadConflictPolicy, "conflict", "rename", "Conflict policy (rename, version, overwrite)")
	uploadCmd.Flags().StringVar(&uploadCacheDir, "cache-dir", "", "Cache directory path")
	uploadCmd.Flags().StringVar(&uploadIncludeExts, "include", "", "Comma-separated list of file extensions to include")
	uploadCmd.Flags().StringVar(&uploadStorageName, "storage", "", "Storage name to use")
}
