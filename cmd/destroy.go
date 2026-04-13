package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

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
		fmt.Fprintf(os.Stderr, "Container '%s' not found\n", containerName)
		os.Exit(1)
	}

	if !destroyForce {
		fmt.Printf("Are you sure you want to destroy container '%s'? [y/N]: ", containerName)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled")
			return
		}
	}

	ctx := context.Background()

	fmt.Printf("Destroying container '%s'...\n", containerName)

	if err := container.Destroy(ctx, containerName, &container.DestroyOptions{
		KeepVolume: destroyKeepVolume,
		VolumeName: containerState.WorkspaceVolume,
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
