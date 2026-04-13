package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	reader := bufio.NewReader(os.Stdin)

	cwd, _ := os.Getwd()
	defaultName := filepath.Base(cwd)

	fmt.Printf("Project name [%s]: ", defaultName)
	nameInput, _ := reader.ReadString('\n')
	name := strings.TrimSpace(nameInput)
	if name == "" {
		name = defaultName
	}

	registry := feature.NewRegistry()

	fmt.Println("\nAvailable features:")
	for _, f := range registry.List() {
		fmt.Printf("  - %s: %s\n", f.Name, f.Description)
	}

	fmt.Print("\nEnter features (comma-separated, e.g., node,go,python): ")
	featuresInput, _ := reader.ReadString('\n')
	featuresInput = strings.TrimSpace(featuresInput)

	features := []types.FeatureSpec{}
	if featuresInput != "" {
		for _, f := range strings.Split(featuresInput, ",") {
			f = strings.TrimSpace(f)
			if f != "" {
				features = append(features, types.FeatureSpec{Name: f})
			}
		}
	}

	cfg := &types.ProjectConfig{
		Name:     name,
		Features: features,
	}

	if err := config.SaveProjectConfig(cfg, ""); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	containerName := naming.GenerateLocalContainerName(name, cwd)

	fmt.Printf("\nCreated .devc/devc.yml\n")
	fmt.Printf("Container name will be: %s\n", containerName)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit .devc/devc.yml to customize features")
	fmt.Println("  2. Run 'devc build' to build the development image")
	fmt.Println("  3. Run 'devc up' to start the container")
}
