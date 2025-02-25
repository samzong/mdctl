package cmd

import (
	"fmt"

	"github.com/samzong/mdctl/internal/processor"

	"github.com/spf13/cobra"
)

var (
	sourceFile     string
	sourceDir      string
	imageOutputDir string

	downloadCmd = &cobra.Command{
		Use:   "download",
		Short: "Download remote images in markdown files",
		Long: `Download remote images in markdown files to local storage and update references.
Examples:
  mdctl download -f post.md
  mdctl download -d content/posts
  mdctl download -f post.md -o assets/images`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sourceFile == "" && sourceDir == "" {
				return fmt.Errorf("either source file (-f) or source directory (-d) must be specified")
			}
			if sourceFile != "" && sourceDir != "" {
				return fmt.Errorf("cannot specify both source file (-f) and source directory (-d)")
			}

			p := processor.New(sourceFile, sourceDir, imageOutputDir)
			return p.Process()
		},
	}
)

func init() {
	downloadCmd.Flags().StringVarP(&sourceFile, "file", "f", "", "Source markdown file to process")
	downloadCmd.Flags().StringVarP(&sourceDir, "dir", "d", "", "Source directory containing markdown files to process")
	downloadCmd.Flags().StringVarP(&imageOutputDir, "output", "o", "", "Output directory for downloaded images (optional)")
}
