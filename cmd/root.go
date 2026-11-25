package cmd

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

var (
	Version     = "dev"
	BuildTime   = "unknown"
	verbose     bool
	veryVerbose bool

	rootCmd = &cobra.Command{
		Use:   "mdctl",
		Short: "A CLI tool for markdown file operations",
		Long: `mdctl is a CLI tool that helps you manage and process markdown files.
Currently supports downloading remote images and more features to come.`,
		Version: fmt.Sprintf("%s (built at %s)", resolvedVersion(Version, debug.ReadBuildInfo), BuildTime),
	}
)

func resolvedVersion(version string, readBuildInfo func() (*debug.BuildInfo, bool)) string {
	if version != "" && version != "dev" {
		return strings.TrimPrefix(version, "v")
	}
	info, ok := readBuildInfo()
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return strings.TrimPrefix(info.Main.Version, "v")
	}
	return "dev"
}

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
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(llmstxtCmd)
	rootCmd.AddCommand(lintCmd)

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&veryVerbose, "vv", false, "Enable very verbose output with detailed information")

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
	uploadCmd.GroupID = "core"
	exportCmd.GroupID = "core"
	llmstxtCmd.GroupID = "core"
	lintCmd.GroupID = "core"
	configCmd.GroupID = "config"
}
