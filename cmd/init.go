package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/IsmailAki/devc/internal/config"
	"github.com/IsmailAki/devc/internal/feature"
	"github.com/IsmailAki/devc/internal/naming"
	"github.com/IsmailAki/devc/pkg/types"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new devc project",
	Long:  "Initialize a new devc project with interactive configuration",
	Run:   runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	if config.IsInProject("") {
		fmt.Println("devc.yml already exists in this project")
		os.Exit(1)
	}

	cwd, _ := os.Getwd()
	defaultName := filepath.Base(cwd)

	name, err := promptProjectName(defaultName)
	if err != nil {
		if isPromptCancelled(err) {
			return
		}
		fmt.Fprintf(os.Stderr, "Error reading project name: %v\n", err)
		os.Exit(1)
	}

	registry := feature.NewRegistry()
	cfg := &types.ProjectConfig{Name: name}
	if err := editProjectPlugins(cfg, registry); err != nil {
		if isPromptCancelled(err) {
			return
		}
		fmt.Fprintf(os.Stderr, "Error selecting plugins: %v\n", err)
		os.Exit(1)
	}

	if err := config.SaveProjectConfig(cfg, ""); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	containerName := naming.GenerateLocalContainerName(name, cwd)

	fmt.Printf("\nCreated .devc/devc.yml\n")
	fmt.Printf("Container name will be: %s\n", containerName)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Run 'devc edit' to update plugins")
	fmt.Println("  2. Run 'devc build' to build the development image")
	fmt.Println("  3. Run 'devc up' to start the container")
}
