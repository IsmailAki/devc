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
	"github.com/IsmailAki/devc/internal/sshconfig"
	"github.com/IsmailAki/devc/internal/state"
	"github.com/IsmailAki/devc/pkg/types"
	"github.com/spf13/cobra"
)

var (
	rebuildForce      bool
	rebuildConfigPath string
)

var rebuildCmd = &cobra.Command{
	Use:   "rebuild [container-name]",
	Short: "Rebuild container with updated features",
	Long: `Rebuild the development container with updated features from devc.yml.

This command:
1. Reads the current devc.yml configuration
2. Builds a new Docker image with updated features
3. Stops and removes the old container (preserving volume)
4. Creates a new container with the preserved volume
5. Restarts the container

Your workspace data in the volume is preserved across the rebuild.

Config override priority:
1. --config flag (highest)
2. Local .devc/devc.yml (if exists)
3. Stored config in ~/.devc/containers/<name>/devc.yml

Example:
  devc rebuild                    # Rebuild with current features
  devc rebuild --force            # Force rebuild even if unchanged
  devc rebuild --config ./devc.yml # Use specific config file`,
	Args: cobra.MaximumNArgs(1),
	Run:  runRebuild,
}

func init() {
	rebuildCmd.Flags().BoolVarP(&rebuildForce, "force", "f", false, "Force rebuild even if features unchanged")
	rebuildCmd.Flags().StringVarP(&rebuildConfigPath, "config", "c", "", "Path to config file to use")
	rootCmd.AddCommand(rebuildCmd)
}

func runRebuild(cmd *cobra.Command, args []string) {
	var containerName string

	if len(args) > 0 {
		containerName = args[0]
	} else {
		containerName = resolveDefaultContainerName()
	}

	if containerName == "" {
		fmt.Fprintln(os.Stderr, "No container specified and not in a devc project directory")
		os.Exit(1)
	}

	currentState, err := state.LoadState(containerName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "No existing container found: %s\n", containerName)
		fmt.Fprintln(os.Stderr, "Run 'devc up' to create a new container")
		os.Exit(1)
	}

	metadata, err := state.LoadMetadata(containerName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load metadata: %v\n", err)
		os.Exit(1)
	}

	cfg, configPath, err := config.ResolveConfig(containerName, rebuildConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if rebuildConfigPath != "" || configPath != state.GetConfigPath(containerName) {
		fmt.Printf("Copying override config to stored location...\n")
		if err := config.CopyOverrideConfig(containerName, configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to store config: %v\n", err)
			os.Exit(1)
		}
	}

	repoName := metadata.Repository
	newImage := feature.GenerateImageTag(repoName, cfg.Features)

	if newImage == currentState.Image && !rebuildForce {
		fmt.Printf("No feature changes detected (image: %s)\n", newImage)
		fmt.Println("Use --force to rebuild anyway")
		os.Exit(0)
	}

	ctx := context.Background()

	fmt.Printf("Rebuilding container '%s'...\n", containerName)
	fmt.Printf("Old image: %s\n", currentState.Image)
	fmt.Printf("New image: %s\n", newImage)

	fmt.Println("\nEnsuring base image exists...")
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

	fmt.Println("\nFeatures:")
	for _, f := range cfg.Features {
		version := f.Version
		if version == "" {
			version = "latest"
		}
		fmt.Printf("  - %s (%s)\n", f.Name, version)
	}

	if len(dockerfile) > 0 {
		fmt.Println("\nBuilding new image...")
		if err := buildFeatureImage(ctx, dockerfile, newImage); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to build image: %v\n", err)
			os.Exit(1)
		}
	}

	volumeName := currentState.WorkspaceVolume
	allocatedPort := currentState.SSHPort

	fmt.Println("\nStopping and removing old container...")
	if err := container.Destroy(ctx, containerName, &container.DestroyOptions{
		KeepVolume: true,
		VolumeName: currentState.WorkspaceVolume,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to remove old container: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Creating new container...")
	createOpts := container.CreateOptions{
		Name:  containerName,
		Image: newImage,
		Port:  allocatedPort,
	}
	if metadata.InitMode == "local" {
		projectRoot := metadata.SourcePath
		if projectRoot == "" {
			projectRoot = currentWorkingDir()
		}
		if metadata.ProjectPath == "" {
			metadata.ProjectPath = remoteProjectPath(projectRoot)
		}
		createOpts.BindMountSource = projectRoot
		createOpts.BindMountTarget = metadata.ProjectPath
		createOpts.WorkingDir = metadata.ProjectPath
	} else {
		createOpts.VolumeName = volumeName
	}

	containerID, err := container.Create(ctx, createOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create container: %v\n", err)
		if volumeName != "" {
			fmt.Fprintf(os.Stderr, "\nYour data is safe in volume: %s\n", volumeName)
		} else if metadata.SourcePath != "" {
			fmt.Fprintf(os.Stderr, "\nYour source data remains in: %s\n", metadata.SourcePath)
		}
		fmt.Fprintln(os.Stderr, "You can retry or manually recover")
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

	newState := &types.ContainerState{
		ContainerID:     containerID,
		Image:           newImage,
		SSHPort:         allocatedPort,
		WorkspaceVolume: volumeName,
		Status:          "running",
		CreatedAt:       time.Now(),
	}

	if err := state.SaveState(containerName, newState); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save container state: %v\n", err)
		os.Exit(1)
	}

	metadata.Features = featureNames
	metadata.DockerImage = newImage
	if err := state.SaveMetadata(containerName, metadata); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to update metadata: %v\n", err)
	}

	if err := sshconfig.AddEntry(*newState, containerName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to update SSH config: %v\n", err)
	}

	fmt.Printf("\nContainer '%s' rebuilt successfully!\n", containerName)
	if volumeName != "" {
		fmt.Printf("Volume preserved: %s\n", volumeName)
	} else if metadata.SourcePath != "" {
		fmt.Printf("Bind mount preserved: %s\n", metadata.SourcePath)
	}
	fmt.Printf("SSH: ssh %s\n", containerName)
	fmt.Printf("Port: %d\n", allocatedPort)
}
