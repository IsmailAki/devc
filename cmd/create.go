package cmd

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/IsmailAki/devc/internal/baseimage"
	"github.com/IsmailAki/devc/internal/config"
	"github.com/IsmailAki/devc/internal/container"
	devcdocker "github.com/IsmailAki/devc/internal/docker"
	"github.com/IsmailAki/devc/internal/feature"
	"github.com/IsmailAki/devc/internal/github"
	"github.com/IsmailAki/devc/internal/language"
	"github.com/IsmailAki/devc/internal/naming"
	"github.com/IsmailAki/devc/internal/port"
	"github.com/IsmailAki/devc/internal/sshconfig"
	"github.com/IsmailAki/devc/internal/state"
	"github.com/IsmailAki/devc/pkg/types"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	createBranch    string
	createLanguages string
)

var createCmd = &cobra.Command{
	Use:   "create <repo-url>",
	Short: "Create a new dev container from a GitHub repository",
	Long: `Create a new dev container from a GitHub repository URL.

This command:
1. Detects programming languages from the repository
2. Creates a container with the base image
3. Clones the repository inside the container
4. Generates devc.yml configuration
5. Builds the feature image
6. Starts the container with SSH access

Container naming: devc-<owner>-<repo>-<randomid>

Example:
  devc create https://github.com/user/repo
  devc create github.com/user/repo
  devc create https://github.com/user/repo --branch develop`,
	Args: cobra.ExactArgs(1),
	Run:  runCreate,
}

func init() {
	createCmd.Flags().StringVarP(&createBranch, "branch", "b", "", "Branch to clone")
	createCmd.Flags().StringVarP(&createLanguages, "languages", "l", "", "Override detected languages (comma-separated)")
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) {
	repoURL := args[0]

	owner, repo, err := parseRepoURL(repoURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	containerName := naming.GenerateContainerName(owner, repo)

	fmt.Printf("Creating container: %s\n", containerName)

	var detectedLanguages []string
	if createLanguages != "" {
		detectedLanguages = strings.Split(createLanguages, ",")
		for i, l := range detectedLanguages {
			detectedLanguages[i] = strings.TrimSpace(l)
		}
	} else {
		fmt.Printf("Fetching repository info for %s/%s...\n", owner, repo)

		langs, err := github.FetchLanguages(owner, repo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching languages: %v\n", err)
			os.Exit(1)
		}

		allLangs := language.DetectAll(langs)
		detectedLanguages = language.FilterSignificant(allLangs)
		if len(detectedLanguages) == 0 {
			detectedLanguages = []string{language.DetectDominant(langs)}
		}
	}

	fmt.Printf("Detected languages: %s\n", strings.Join(detectedLanguages, ", "))

	projectConfig := config.GenerateProjectConfig(repo, repoURL, createBranch, detectedLanguages)

	ctx := context.Background()

	fmt.Println("Ensuring base image exists...")
	if err := baseimage.Ensure(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to ensure base image: %v\n", err)
		os.Exit(1)
	}

	registry := feature.NewRegistry()

	dockerfile, err := feature.GenerateDockerfile(projectConfig.Features, registry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate Dockerfile: %v\n", err)
		os.Exit(1)
	}

	image := feature.GenerateImageTag(repo, projectConfig.Features)

	fmt.Printf("Building feature image: %s\n", image)
	fmt.Println("Features:")
	for _, f := range projectConfig.Features {
		version := f.Version
		if version == "" {
			version = "latest"
		}
		fmt.Printf("  - %s (%s)\n", f.Name, version)
	}

	if err := buildFeatureImage(ctx, dockerfile, image); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build feature image: %v\n", err)
		os.Exit(1)
	}

	allocatedPort, err := port.Allocate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to allocate port: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Allocated port: %d\n", allocatedPort)

	fmt.Println("Creating container...")
	createOpts := container.CreateOptions{
		Name:  containerName,
		Image: image,
		Port:  allocatedPort,
		Env:   mergeEnv(nil, projectConfig.Env),
	}
	configureContainerRuntime(containerName, projectConfig.Features, &createOpts)

	containerID, err := container.Create(ctx, createOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create container: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Starting container...")
	if err := container.Start(ctx, containerName, &container.StartOptions{
		WaitForSSH: true,
		Timeout:    120 * time.Second,
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

	if err := container.PrepareWorkspace(ctx, containerName, "/workspace"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to prepare workspace: %v\n", err)
		os.Exit(1)
	}

	repoPath := "/workspace/" + repo
	cloneCmd := []string{"git", "clone"}
	if createBranch != "" {
		cloneCmd = append(cloneCmd, "-b", createBranch)
	}
	cloneCmd = append(cloneCmd, repoURL, repoPath)

	fmt.Printf("Cloning repository to %s...\n", repoPath)
	if err := container.Exec(ctx, containerName, container.ExecOptions{
		Cmd:        cloneCmd,
		User:       "dev",
		WorkingDir: "/workspace",
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to clone repository: %v\n", err)
		if cleanupErr := container.Destroy(ctx, containerName, &container.DestroyOptions{
			VolumeName:   naming.GenerateVolumeName(containerName),
			ExtraVolumes: []string{naming.GenerateDockerVolumeName(containerName)},
		}); cleanupErr != nil {
			fmt.Fprintf(os.Stderr, "Cleanup warning: failed to remove incomplete container: %v\n", cleanupErr)
		}
		os.Exit(1)
	}

	configData, err := serializeProjectConfig(projectConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to serialize config: %v\n", err)
		os.Exit(1)
	}
	if err := state.SaveConfig(containerName, configData); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
		os.Exit(1)
	}

	featureNames := make([]string, len(projectConfig.Features))
	for i, f := range projectConfig.Features {
		featureNames[i] = f.Name
	}

	containerState := &types.ContainerState{
		ContainerID:     containerID,
		Image:           image,
		SSHPort:         allocatedPort,
		WorkspaceVolume: naming.GenerateVolumeName(containerName),
		DockerVolume:    createOpts.DockerDataVolume,
		Status:          "running",
		CreatedAt:       time.Now(),
	}

	if err := state.SaveState(containerName, containerState); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save container state: %v\n", err)
		os.Exit(1)
	}

	metadata := &types.ContainerMetadata{
		Name:        containerName,
		Repository:  repo,
		Owner:       owner,
		Branch:      createBranch,
		Languages:   detectedLanguages,
		Features:    featureNames,
		InitMode:    "github",
		ConfigPath:  state.GetConfigPath(containerName),
		ProjectPath: repoPath,
		DockerImage: image,
		CreatedAt:   time.Now(),
	}

	if err := state.SaveMetadata(containerName, metadata); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save metadata: %v\n", err)
		os.Exit(1)
	}

	if err := sshconfig.AddEntry(*containerState, containerName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to update SSH config: %v\n", err)
	}

	fmt.Printf("\nContainer '%s' created successfully!\n", containerName)
	fmt.Printf("  Languages: %s\n", strings.Join(detectedLanguages, ", "))
	fmt.Printf("  SSH: ssh %s\n", containerName)
	fmt.Printf("  Root SSH: ssh %s\n", sshconfig.RootHostName(containerName))
	fmt.Printf("  Port: %d\n", allocatedPort)
	fmt.Printf("\nTo connect:\n")
	fmt.Printf("  SSH: ssh %s\n", containerName)
	fmt.Printf("  Root SSH: ssh %s\n", sshconfig.RootHostName(containerName))
	fmt.Printf("  VS Code: code --remote ssh-remote+%s %s\n", containerName, repoPath)
}

func parseRepoURL(url string) (owner, repo string, err error) {
	patterns := []string{
		`^https?://github\.com/([^/]+)/([^/]+)/?$`,
		`^git@github\.com:([^/]+)/([^/]+)\.git$`,
		`^github\.com/([^/]+)/([^/]+)/?$`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if matches != nil {
			owner = matches[1]
			repo = strings.TrimSuffix(matches[2], ".git")
			return owner, repo, nil
		}
	}

	return "", "", fmt.Errorf("invalid repository URL format: use https://github.com/<owner>/<repo>, github.com/<owner>/<repo>, or git@github.com:<owner>/<repo>.git")
}

func serializeProjectConfig(cfg *types.ProjectConfig) ([]byte, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	return data, nil
}

func buildFeatureImage(ctx context.Context, dockerfile, tag string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	header := &tar.Header{
		Name: "Dockerfile",
		Mode: 0644,
		Size: int64(len(dockerfile)),
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}
	if _, err := tw.Write([]byte(dockerfile)); err != nil {
		return fmt.Errorf("failed to write dockerfile: %w", err)
	}
	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	options := dockertypes.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{tag},
		Remove:     true,
		Version:    dockertypes.BuilderBuildKit,
	}

	resp, err := cli.ImageBuild(ctx, &buf, options)
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	if err := devcdocker.ConsumeBuildOutput(resp.Body, os.Stdout); err != nil {
		return err
	}

	return nil
}
