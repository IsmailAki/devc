package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	appVersion = defaultAppVersion
	commit     = defaultCommit
	date       = defaultDate
)

var rootCmd = &cobra.Command{
	Use:   "devc",
	Short: "Development Container Manager",
	Long: `Devc is a CLI tool that creates and manages development containers.

Each container provides an isolated development environment accessible viaSSH,
with support for multiple IDEs (VS Code, JetBrains) through standard SSH connections.

Quick start:
  devc init          Create a new devc project
  devc edit          Edit plugins for an existing container
  devc build         Build the development image
  devc up            Start the development container
  devc ssh           Connect to the container
  devc connect vscode  Connect VS Code to the container`,
	Version: defaultAppVersion,
	Args:    cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && isRepoURL(args[0]) {
			runCreate(cmd, args)
			return
		}
		cmd.Help()
	},
}

func isRepoURL(arg string) bool {
	if !strings.HasPrefix(arg, "http://") && !strings.HasPrefix(arg, "https://") && !strings.HasPrefix(arg, "git@") && !strings.HasPrefix(arg, "github.com/") {
		return false
	}

	_, _, err := parseRepoURL(arg)
	return err == nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	metadata := loadBuildMetadata()
	appVersion = metadata.version
	commit = metadata.commit
	date = metadata.date
	rootCmd.Version = appVersion
	rootCmd.SetVersionTemplate(fmt.Sprintf("devc version %s (commit: %s, built: %s)\n", appVersion, commit, date))
}
