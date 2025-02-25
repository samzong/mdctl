package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	BuildTime = "unknown"

	rootCmd = &cobra.Command{
		Use:   "mdctl",
		Short: "A CLI tool for markdown file operations",
		Long: `mdctl is a CLI tool that helps you manage and process markdown files.
Currently supports downloading remote images and more features to come.`,
		Version: fmt.Sprintf("%s (built at %s)", Version, BuildTime),
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add commands first
	rootCmd.AddCommand(translateCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(configCmd)

	// Then add groups and set group IDs
	rootCmd.AddGroup(&cobra.Group{
		ID:    "core",
		Title: "Core Commands:",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "config",
		Title: "Configuration Commands:",
	})

	// Set group for each command
	translateCmd.GroupID = "core"
	downloadCmd.GroupID = "core"
	configCmd.GroupID = "config"
}
