package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/IsmailAki/devc/internal/container"
	"github.com/IsmailAki/devc/internal/sshconfig"
	"github.com/IsmailAki/devc/internal/state"
	"github.com/spf13/cobra"
)

var (
	destroyForce      bool
	destroyKeepVolume bool
)

var destroyCmd = &cobra.Command{
	Use:     "destroy [name]",
	Aliases: []string{"rm", "remove"},
	Short:   "Destroy a development container",
	Long:    "Destroy the development container and optionally remove volumes and images",
	Args:    cobra.MaximumNArgs(1),
	Run:     runDestroy,
}

func init() {
	destroyCmd.Flags().BoolVarP(&destroyForce, "force", "f", false, "Skip confirmation")
	destroyCmd.Flags().BoolVar(&destroyKeepVolume, "keep-volume", false, "Don't remove the volume")
	rootCmd.AddCommand(destroyCmd)
}

func runDestroy(cmd *cobra.Command, args []string) {
	containerName, err := resolveDestroyContainerName(args)
	if err != nil {
		if isPromptCancelled(err) {
			return
		}
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	containerState, err := state.LoadState(containerName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Container '%s' not found\n", containerName)
		os.Exit(1)
	}

	if !destroyForce {
		confirmed, err := promptConfirm(fmt.Sprintf("Destroy container '%s'?", containerName), false)
		if err != nil {
			if isPromptCancelled(err) {
				return
			}
			fmt.Fprintf(os.Stderr, "Failed to read confirmation: %v\n", err)
			os.Exit(1)
		}
		if !confirmed {
			fmt.Println("Cancelled")
			return
		}
	}

	ctx := context.Background()

	fmt.Printf("Destroying container '%s'...\n", containerName)

	if err := container.Destroy(ctx, containerName, &container.DestroyOptions{
		KeepVolume:   destroyKeepVolume,
		VolumeName:   containerState.WorkspaceVolume,
		ExtraVolumes: []string{containerState.DockerVolume},
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to destroy container: %v\n", err)
		os.Exit(1)
	}

	if err := sshconfig.RemoveEntry(containerName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to remove SSH config entry: %v\n", err)
	}

	if err := container.RemoveContainerDir(containerName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to remove container directory: %v\n", err)
	}

	fmt.Printf("Container '%s' destroyed\n", containerName)
}

func resolveDestroyContainerName(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	containers, err := loadContainerInfos(true)
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}
	if len(containers) == 0 {
		return "", fmt.Errorf("no container specified and no containers found")
	}

	if len(containers) == 1 {
		return containers[0].Name, nil
	}

	return pickContainer(containers, "Select a container to destroy:")
}
