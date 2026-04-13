package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/IsmailAki/devc/internal/container"
	"github.com/IsmailAki/devc/internal/state"
	"github.com/IsmailAki/devc/pkg/types"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop a development container",
	Long:  "Stop the development container",
	Args:  cobra.MaximumNArgs(1),
	Run:   runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) {
	var containerName string

	if len(args) > 0 {
		containerName = args[0]
	} else {
		containerName = firstContainerName()
	}

	if containerName == "" {
		fmt.Fprintln(os.Stderr, "No container specified and no containers found")
		os.Exit(1)
	}

	containerState, err := state.LoadState(containerName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Container '%s' not found: %v\n", containerName, err)
		os.Exit(1)
	}
	if containerState.Status != "running" {
		fmt.Printf("Container '%s' is already stopped\n", containerName)
		return
	}

	ctx := context.Background()

	if err := container.Stop(ctx, containerName); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop container: %v\n", err)
		os.Exit(1)
	}

	newState := &types.ContainerState{
		ContainerID:     containerState.ContainerID,
		Image:           containerState.Image,
		SSHPort:         containerState.SSHPort,
		WorkspaceVolume: containerState.WorkspaceVolume,
		DockerVolume:    containerState.DockerVolume,
		Status:          "stopped",
		CreatedAt:       containerState.CreatedAt,
	}

	if err := state.SaveState(containerName, newState); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to update container state: %v\n", err)
	}

	fmt.Printf("Container '%s' stopped\n", containerName)
}
