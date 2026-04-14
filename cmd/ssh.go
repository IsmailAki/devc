package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/IsmailAki/devc/internal/sshconfig"
	"github.com/IsmailAki/devc/internal/state"
	"github.com/spf13/cobra"
)

var sshAsRoot bool

var sshCmd = &cobra.Command{
	Use:   "ssh [name]",
	Short: "SSH into the development container",
	Long:  "Connect to the development container via SSH",
	Args:  cobra.MaximumNArgs(1),
	Run:   runSSH,
}

func init() {
	sshCmd.Flags().BoolVar(&sshAsRoot, "root", false, "Connect as root instead of dev")
	rootCmd.AddCommand(sshCmd)
}

func runSSH(cmd *cobra.Command, args []string) {
	containerName, err := resolveSSHContainerName(args)
	if err != nil {
		if isPromptCancelled(err) {
			return
		}
		fmt.Fprintf(os.Stderr, "%v\n", err)
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

	targetHost := containerName
	if sshAsRoot {
		targetHost = sshconfig.RootHostName(containerName)
	}

	sshExecCmd := exec.Command("ssh", targetHost)
	sshExecCmd.Stdin = os.Stdin
	sshExecCmd.Stdout = os.Stdout
	sshExecCmd.Stderr = os.Stderr

	if err := sshExecCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "SSH connection failed: %v\n", err)
		os.Exit(1)
	}
}

func resolveSSHContainerName(args []string) (string, error) {
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

	return pickContainer(containers, "Select a container to connect:")
}
