package baseimage

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/IsmailAki/devc/internal/docker"
	"github.com/docker/docker/api/types"
)

func Ensure(ctx context.Context) error {
	exists, err := Exists(ctx)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return Build(ctx)
}

func Exists(ctx context.Context) (bool, error) {
	cli, err := docker.GetClient()
	if err != nil {
		return false, err
	}

	_, _, err = cli.ImageInspectWithRaw(ctx, FullImageName)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func Build(ctx context.Context) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	header := &tar.Header{
		Name: "Dockerfile",
		Mode: 0644,
		Size: int64(len(BaseDockerfile)),
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}
	if _, err := tw.Write([]byte(BaseDockerfile)); err != nil {
		return fmt.Errorf("failed to write tar content: %w", err)
	}
	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	cli, err := docker.GetClient()
	if err != nil {
		return err
	}

	fmt.Println("Building devc-base image...")
	resp, err := cli.ImageBuild(ctx, &buf, types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{FullImageName},
		Remove:     true,
	})
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	if err := docker.ConsumeBuildOutput(resp.Body, os.Stdout); err != nil {
		return err
	}

	fmt.Println("Successfully built devc-base image")
	return nil
}
