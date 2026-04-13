package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/IsmailAki/devc/internal/feature"
	"github.com/spf13/cobra"
)

var featuresCmd = &cobra.Command{
	Use:   "features",
	Short: "Manage development features",
	Long:  "List and inspect available development features",
}

var featuresListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all available features",
	Long:    "List all available development features that can be added to containers",
	Run:     runFeaturesList,
}

var featuresShowCmd = &cobra.Command{
	Use:   "show <feature>",
	Short: "Show feature details",
	Long:  "Show detailed information about a specific feature including parameters and dependencies",
	Args:  cobra.ExactArgs(1),
	Run:   runFeaturesShow,
}

func init() {
	featuresCmd.AddCommand(featuresListCmd)
	featuresCmd.AddCommand(featuresShowCmd)
	rootCmd.AddCommand(featuresCmd)
}

func runFeaturesList(cmd *cobra.Command, args []string) {
	registry := feature.NewRegistry()
	features := registry.List()

	if len(features) == 0 {
		fmt.Println("No features available")
		return
	}

	sort.Slice(features, func(i, j int) bool {
		return features[i].Name < features[j].Name
	})

	fmt.Println("Available features:")
	fmt.Println()
	for _, f := range features {
		desc := f.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Printf("  %-15s %s", f.Name, desc)
		if len(f.Depends) > 0 {
			fmt.Printf(" (depends: %s)", strings.Join(f.Depends, ", "))
		}
		fmt.Println()
	}
	fmt.Println()
	fmt.Println("Use 'devc features show <name>' for more details")
}

func runFeaturesShow(cmd *cobra.Command, args []string) {
	featureName := args[0]

	registry := feature.NewRegistry()
	f, ok := registry.Get(featureName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Feature '%s' not found\n", featureName)
		fmt.Fprintf(os.Stderr, "Use 'devc features list' to see available features\n")
		os.Exit(1)
	}

	fmt.Printf("Feature: %s\n", f.Name)
	fmt.Printf("Description: %s\n", f.Description)
	fmt.Printf("Version: %s\n", f.Version)

	if len(f.Depends) > 0 {
		fmt.Printf("Dependencies: %s\n", strings.Join(f.Depends, ", "))
	}

	if len(f.Params) > 0 {
		fmt.Println("\nParameters:")
		paramNames := make([]string, 0, len(f.Params))
		for name := range f.Params {
			paramNames = append(paramNames, name)
		}
		sort.Strings(paramNames)
		for _, name := range paramNames {
			param := f.Params[name]
			fmt.Printf("  %s", name)
			if param.Default != "" {
				fmt.Printf(" (default: %s)", param.Default)
			}
			fmt.Println()
			if param.Description != "" {
				fmt.Printf("    %s\n", param.Description)
			}
		}
	}

	if len(f.Install) > 0 {
		fmt.Printf("\nInstall steps: %d\n", len(f.Install))
		for i, step := range f.Install {
			fmt.Printf("  %d. %s\n", i+1, step.Name)
		}
	}
}
