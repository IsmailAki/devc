package cmd

import (
	"fmt"
	"os"
	"reflect"

	"github.com/IsmailAki/devc/internal/feature"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [container-name]",
	Short: "Edit container plugins",
	Long:  "Edit plugins for an existing development container with an interactive terminal UI",
	Args:  cobra.MaximumNArgs(1),
	Run:   runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) {
	containerName, err := resolveEditContainerName(args)
	if err != nil {
		if isPromptCancelled(err) {
			return
		}
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	target, cfg, err := resolveEditableConfigTarget(containerName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load editable config: %v\n", err)
		os.Exit(1)
	}

	edited := cloneProjectConfig(cfg)
	registry := feature.NewRegistry()
	if err := editProjectPlugins(edited, registry); err != nil {
		if isPromptCancelled(err) {
			return
		}
		fmt.Fprintf(os.Stderr, "Failed to edit plugins: %v\n", err)
		os.Exit(1)
	}

	if reflect.DeepEqual(cfg.Features, edited.Features) {
		fmt.Println("No plugin changes made")
		return
	}

	if err := target.Save(edited); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Saved plugin changes to %s\n", target.ActivePath)

	rebuild, err := promptRebuildNow(containerName)
	if err != nil {
		if isPromptCancelled(err) {
			return
		}
		fmt.Fprintf(os.Stderr, "Failed to read rebuild choice: %v\n", err)
		os.Exit(1)
	}

	if !rebuild {
		fmt.Printf("Run 'devc rebuild %s' when you are ready to apply the changes\n", containerName)
		return
	}

	if err := rebuildContainer(containerName, target.ActivePath, false); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rebuild container: %v\n", err)
		os.Exit(1)
	}
}

func resolveEditContainerName(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	containers, err := loadContainerInfos(true)
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}
	if len(containers) == 0 {
		return "", fmt.Errorf("no dev containers found")
	}

	containerName, err := pickContainer(containers, "Select a container to edit:")
	if err != nil {
		return "", err
	}

	return containerName, nil
}
