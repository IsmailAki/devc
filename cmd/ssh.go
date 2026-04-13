package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/IsmailAki/devc/internal/state"
	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh [name]",
	Short: "SSH into the development container",
	Long:  "Connect to the development container via SSH",
	Args:  cobra.MaximumNArgs(1),
	Run:   runSSH,
}

func init() {
	rootCmd.AddCommand(sshCmd)
}

func runSSH(cmd *cobra.Command, args []string) {
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
		fmt.Fprintf(os.Stderr, "Container '%s' is not running (status: %s)\n", containerName, containerState.Status)
		fmt.Fprintln(os.Stderr, "Run 'devc up' to start the container")
		os.Exit(1)
	}

	sshCmd := exec.Command("ssh", containerName)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	if err := sshCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "SSH connection failed: %v\n", err)
		os.Exit(1)
	}
}
