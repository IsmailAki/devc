package cmd

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/IsmailAki/devc/internal/baseimage"
	"github.com/IsmailAki/devc/internal/config"
	devcdocker "github.com/IsmailAki/devc/internal/docker"
	"github.com/IsmailAki/devc/internal/feature"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var (
	buildNoCache bool
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build project Docker image with features",
	Long:  "Build the project-specific Docker image with all configured features",
	Run:   runBuild,
}

func init() {
	buildCmd.Flags().BoolVar(&buildNoCache, "no-cache", false, "Don't use cached layers")
	rootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) {
	if !config.IsInProject("") {
		fmt.Fprintln(os.Stderr, "Not in a devc project. Run 'devc init' first.")
		os.Exit(1)
	}

	cfg, err := config.LoadProjectConfig("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

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

	imageTag := feature.GenerateImageTag(cfg.Name, cfg.Features)

	fmt.Printf("Building image: %s\n", imageTag)
	fmt.Println("Features:")
	for _, f := range cfg.Features {
		version := f.Version
		if version == "" {
			version = "latest"
		}
		fmt.Printf("  - %s (%s)\n", f.Name, version)
	}

	if err := buildImage(ctx, dockerfile, imageTag, buildNoCache); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build image: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nSuccessfully built image: %s\n", imageTag)
	fmt.Println("Run 'devc up' to start the container")
}

func buildImage(ctx context.Context, dockerfile, tag string, noCache bool) error {
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

	fmt.Println("\n--- Dockerfile ---")
	fmt.Println(dockerfile)
	fmt.Println("---")
	fmt.Println()

	fmt.Println("Building Docker image...")
	resp, err := cli.ImageBuild(ctx, &buf, types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{tag},
		Remove:     true,
		NoCache:    noCache,
		Version:    types.BuilderBuildKit,
	})
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	if err := devcdocker.ConsumeBuildOutput(resp.Body, os.Stdout); err != nil {
		return err
	}

	return nil
}
