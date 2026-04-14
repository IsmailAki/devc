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

	if err := rebuildContainer(containerName, rebuildConfigPath, rebuildForce); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func rebuildContainer(containerName string, overridePath string, force bool) error {
	currentState, err := state.LoadState(containerName)
	if err != nil {
		return fmt.Errorf("no existing container found: %s\nRun 'devc up' to create a new container", containerName)
	}

	metadata, err := state.LoadMetadata(containerName)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	cfg, configPath, err := config.ResolveConfig(containerName, overridePath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if overridePath != "" || configPath != state.GetConfigPath(containerName) {
		fmt.Printf("Copying override config to stored location...\n")
		if err := config.CopyOverrideConfig(containerName, configPath); err != nil {
			return fmt.Errorf("failed to store config: %w", err)
		}
	}

	repoName := metadata.Repository
	newImage := feature.GenerateImageTag(repoName, cfg.Features)

	if newImage == currentState.Image && !force {
		fmt.Printf("No feature changes detected (image: %s)\n", newImage)
		fmt.Println("Use --force to rebuild anyway")
		return nil
	}

	ctx := context.Background()

	fmt.Printf("Rebuilding container '%s'...\n", containerName)
	fmt.Printf("Old image: %s\n", currentState.Image)
	fmt.Printf("New image: %s\n", newImage)

	fmt.Println("\nEnsuring base image exists...")
	if err := baseimage.Ensure(ctx); err != nil {
		return fmt.Errorf("failed to ensure base image: %w", err)
	}

	registry := feature.NewRegistry()

	dockerfile, err := feature.GenerateDockerfile(cfg.Features, registry)
	if err != nil {
		return fmt.Errorf("failed to generate Dockerfile: %w", err)
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
			return fmt.Errorf("failed to build image: %w", err)
		}
	}

	volumeName := currentState.WorkspaceVolume
	allocatedPort := currentState.SSHPort

	fmt.Println("\nStopping and removing old container...")
	keepDockerVolume := hasFeature(cfg.Features, "docker") && currentState.DockerVolume != ""
	if err := container.Destroy(ctx, containerName, &container.DestroyOptions{
		KeepVolume:       true,
		KeepExtraVolumes: keepDockerVolume,
		VolumeName:       currentState.WorkspaceVolume,
		ExtraVolumes:     []string{currentState.DockerVolume},
	}); err != nil {
		return fmt.Errorf("failed to remove old container: %w", err)
	}

	fmt.Println("Creating new container...")
	createOpts := container.CreateOptions{
		Name:  containerName,
		Image: newImage,
		Port:  allocatedPort,
		Env:   mergeEnv(nil, cfg.Env),
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
	configureContainerRuntime(containerName, cfg.Features, &createOpts)

	containerID, err := container.Create(ctx, createOpts)
	if err != nil {
		if volumeName != "" {
			return fmt.Errorf("failed to create container: %w\n\nYour data is safe in volume: %s\nYou can retry or manually recover", err, volumeName)
		} else if metadata.SourcePath != "" {
			return fmt.Errorf("failed to create container: %w\n\nYour source data remains in: %s\nYou can retry or manually recover", err, metadata.SourcePath)
		}
		return fmt.Errorf("failed to create container: %w", err)
	}

	fmt.Println("Starting container...")
	if err := container.Start(ctx, containerName, &container.StartOptions{
		WaitForSSH: true,
		Timeout:    60 * time.Second,
		Port:       allocatedPort,
	}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Println("Setting up SSH key...")
	if err := container.SetupSSHKey(ctx, containerName); err != nil {
		return fmt.Errorf("failed to setup SSH key: %w", err)
	}

	if metadata.InitMode != "local" {
		if err := container.PrepareWorkspace(ctx, containerName, "/workspace"); err != nil {
			return fmt.Errorf("failed to prepare workspace: %w", err)
		}
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
		DockerVolume:    createOpts.DockerDataVolume,
		Status:          "running",
		CreatedAt:       time.Now(),
	}

	if err := state.SaveState(containerName, newState); err != nil {
		return fmt.Errorf("failed to save container state: %w", err)
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

	return nil
}
