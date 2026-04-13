package cmd

import (
	"fmt"
	"os"

	"github.com/IsmailAki/devc/internal/ide"
	"github.com/IsmailAki/devc/internal/state"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect <ide> [name]",
	Short: "Connect IDE to the development container",
	Long:  "Connect VS Code or JetBrains IDE to the development container via SSH",
	Args:  cobra.MinimumNArgs(1),
	Run:   runConnect,
}

func init() {
	rootCmd.AddCommand(connectCmd)
}

func runConnect(cmd *cobra.Command, args []string) {
	ideName := args[0]

	var containerName string
	if len(args) > 1 {
		containerName = args[1]
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

	var info string
	switch ideName {
	case "vscode", "code":
		if err := ide.ConnectVSCode(containerName); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect VS Code: %v\n", err)
			fmt.Fprintln(os.Stderr, "\nManual connection instructions:")
			info, _ = ide.VSCodeConnectionInfo(containerName)
			fmt.Println(info)
			os.Exit(1)
		}

	case "jetbrains", "idea", "intellij":
		info, err = ide.JetBrainsConnectionInfo(containerName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get JetBrains connection info: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(info)
	default:
		fmt.Fprintf(os.Stderr, "Unknown IDE: %s\n", ideName)
		fmt.Fprintln(os.Stderr, "Supported IDEs: vscode, jetbrains")
		os.Exit(1)
	}
}
