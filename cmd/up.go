package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/IsmailAki/devc/internal/baseimage"
	"github.com/IsmailAki/devc/internal/config"
	"github.com/IsmailAki/devc/internal/container"
	"github.com/IsmailAki/devc/internal/feature"
	"github.com/IsmailAki/devc/internal/port"
	"github.com/IsmailAki/devc/internal/sshconfig"
	"github.com/IsmailAki/devc/internal/state"
	"github.com/IsmailAki/devc/pkg/types"
	"github.com/spf13/cobra"
)

var (
	upBuild bool
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start the development container",
	Long:  "Start the development container with SSH access",
	Run:   runUp,
}

func init() {
	upCmd.Flags().BoolVar(&upBuild, "build", false, "Build image if needed")
	rootCmd.AddCommand(upCmd)
}

func runUp(cmd *cobra.Command, args []string) {
	projectRoot, cfg, err := resolveCurrentProject()
	if err != nil {
		if !config.IsInProject("") {
			fmt.Fprintln(os.Stderr, "Not in a devc project. Run 'devc init' first.")
		} else {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		}
		os.Exit(1)
	}

	if err := ensureProjectName(cfg, projectRoot); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve project name: %v\n", err)
		os.Exit(1)
	}

	projectName := cfg.Name
	containerName := resolveLocalContainerName(projectRoot, cfg)
	remotePath := remoteProjectPath(projectRoot)
	ctx := context.Background()

	existingState, err := state.LoadState(containerName)
	if err == nil {
		if existingState.Status != "running" {
			fmt.Printf("Starting existing container '%s'...\n", containerName)
			if err := container.Start(ctx, containerName, &container.StartOptions{
				WaitForSSH: true,
				Timeout:    60 * time.Second,
				Port:       existingState.SSHPort,
			}); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to start container: %v\n", err)
				os.Exit(1)
			}

			updatedState := *existingState
			updatedState.Status = "running"
			if err := state.SaveState(containerName, &updatedState); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to update container state: %v\n", err)
			}
			existingState = &updatedState
		} else {
			fmt.Printf("Container '%s' is already running\n", containerName)
		}

		if err := container.SetupSSHKey(ctx, containerName); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to setup SSH key: %v\n", err)
			os.Exit(1)
		}

		if err := config.SaveGlobalConfig(containerName, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save config to global directory: %v\n", err)
		}

		metadata := localMetadata(containerName, projectRoot, remotePath, cfg)
		metadata.CreatedAt = time.Now()
		if err := state.SaveMetadata(containerName, metadata); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save metadata: %v\n", err)
		}

		if err := sshconfig.AddEntry(*existingState, containerName); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to update SSH config: %v\n", err)
		}

		fmt.Printf("SSH: ssh %s\n", containerName)
		fmt.Printf("Port: %d\n", existingState.SSHPort)
		fmt.Printf("\nTo connect with VS Code:\n")
		fmt.Printf("  code --remote ssh-remote+%s %s\n", containerName, remotePath)
		return
	}

	fmt.Printf("Creating container for project '%s'...\n", projectName)
	fmt.Printf("Container name: %s\n", containerName)

	allocatedPort, err := port.Allocate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to allocate port: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Allocated port: %d\n", allocatedPort)

	image := feature.GenerateImageTag(projectName, cfg.Features)

	if upBuild {
		fmt.Println("Ensuring base image exists...")
		if err := baseimage.Ensure(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to ensure base image: %v\n", err)
			os.Exit(1)
		}

		registry := feature.NewRegistry()
		dockerfile, err := feature.GenerateDockerfile(cfg.Features, registry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate Dockerfile: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Building image: %s\n", image)
		if err := buildFeatureImage(ctx, dockerfile, image); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to build image: %v\n", err)
			os.Exit(1)
		}
	}

	createOpts := container.CreateOptions{
		Name:            containerName,
		Image:           image,
		Port:            allocatedPort,
		BindMountSource: projectRoot,
		BindMountTarget: remotePath,
		WorkingDir:      remotePath,
		Env:             mergeEnv(nil, cfg.Env),
	}
	configureContainerRuntime(containerName, cfg.Features, &createOpts)

	containerID, err := container.Create(ctx, createOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create container: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Starting container...")
	if err := container.Start(ctx, containerName, &container.StartOptions{
		WaitForSSH: true,
		Timeout:    60 * time.Second,
		Port:       allocatedPort,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start container: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Setting up SSH key...")
	if err := container.SetupSSHKey(ctx, containerName); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup SSH key: %v\n", err)
		os.Exit(1)
	}

	featureNames := make([]string, len(cfg.Features))
	for i, f := range cfg.Features {
		featureNames[i] = f.Name
	}

	containerState := &types.ContainerState{
		ContainerID:     containerID,
		Image:           image,
		SSHPort:         allocatedPort,
		WorkspaceVolume: "",
		DockerVolume:    createOpts.DockerDataVolume,
		Status:          "running",
		CreatedAt:       time.Now(),
	}

	if err := state.SaveState(containerName, containerState); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save container state: %v\n", err)
		os.Exit(1)
	}

	if err := config.SaveGlobalConfig(containerName, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save config to global directory: %v\n", err)
	}

	metadata := localMetadata(containerName, projectRoot, remotePath, cfg)
	metadata.Features = featureNames
	metadata.CreatedAt = time.Now()

	if err := state.SaveMetadata(containerName, metadata); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save metadata: %v\n", err)
	}

	if err := sshconfig.AddEntry(*containerState, containerName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to update SSH config: %v\n", err)
	}

	fmt.Printf("\nContainer '%s' started successfully!\n", containerName)
	fmt.Printf("SSH: ssh %s\n", containerName)
	fmt.Printf("Port: %d\n", allocatedPort)
	fmt.Printf("\nTo connect with VS Code:\n")
	fmt.Printf("  code --remote ssh-remote+%s %s\n", containerName, remotePath)
}
