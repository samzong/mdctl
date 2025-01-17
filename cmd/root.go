package cmd

import (
	"fmt"
	"os"

	"mdctl/internal/processor"

	"github.com/spf13/cobra"
)

var (
	sourceFile     string
	sourceDir      string
	imageOutputDir string
	rootCmd        = &cobra.Command{
		Use:   "mdctl",
		Short: "A tool to process markdown files and download remote images",
		Long: `mdctl is a CLI tool that processes markdown files and downloads remote images to local storage.
It can process a single file or recursively process all files in a directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sourceFile == "" && sourceDir == "" {
				return fmt.Errorf("either source file or source directory must be specified")
			}
			if sourceFile != "" && sourceDir != "" {
				return fmt.Errorf("cannot specify both source file and source directory")
			}

			p := processor.New(sourceFile, sourceDir, imageOutputDir)
			return p.Process()
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&sourceFile, "file", "f", "", "Source markdown file to process")
	rootCmd.PersistentFlags().StringVarP(&sourceDir, "dir", "d", "", "Source directory containing markdown files to process")
	rootCmd.PersistentFlags().StringVarP(&imageOutputDir, "output", "o", "", "Output directory for downloaded images (optional)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
