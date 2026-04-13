package container

import (
	"context"
	"fmt"
	"time"

	"github.com/IsmailAki/devc/internal/naming"
	"github.com/IsmailAki/devc/internal/state"
	"github.com/IsmailAki/devc/pkg/types"
)

type CreateOptions struct {
	Name            string
	Image           string
	Port            int
	VolumeName      string
	BindMountSource string
	BindMountTarget string
	WorkingDir      string
	Env             map[string]string
}

func Create(ctx context.Context, opts CreateOptions) (string, error) {
	cli, err := getDockerClient()
	if err != nil {
		return "", fmt.Errorf("failed to get docker client: %w", err)
	}

	if opts.VolumeName == "" && opts.BindMountSource == "" {
		opts.VolumeName = naming.GenerateVolumeName(opts.Name)
	}

	if opts.VolumeName != "" {
		if err := createVolume(ctx, opts.VolumeName); err != nil {
			return "", fmt.Errorf("failed to create volume: %w", err)
		}
	}

	containerID, err := createContainer(ctx, cli, opts.Name, opts)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	return containerID, nil
}

type StartOptions struct {
	WaitForSSH bool
	Timeout    time.Duration
	Port       int
}

func Start(ctx context.Context, name string, opts *StartOptions) error {
	if opts == nil {
		opts = &StartOptions{WaitForSSH: true, Timeout: 30 * time.Second}
	}

	cli, err := getDockerClient()
	if err != nil {
		return fmt.Errorf("failed to get docker client: %w", err)
	}

	if err := startContainer(ctx, cli, name); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	if opts.WaitForSSH && opts.Port > 0 {
		if err := waitForSSH(opts.Port, opts.Timeout); err != nil {
			return fmt.Errorf("failed waiting for SSH: %w", err)
		}
	}

	return nil
}

func Stop(ctx context.Context, name string) error {
	cli, err := getDockerClient()
	if err != nil {
		return fmt.Errorf("failed to get docker client: %w", err)
	}

	if err := stopContainer(ctx, cli, name); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	return nil
}

type DestroyOptions struct {
	KeepVolume bool
	VolumeName string
}

func Destroy(ctx context.Context, name string, opts *DestroyOptions) error {
	if opts == nil {
		opts = &DestroyOptions{}
	}

	cli, err := getDockerClient()
	if err != nil {
		return fmt.Errorf("failed to get docker client: %w", err)
	}

	_ = stopContainer(ctx, cli, name)

	if err := removeContainer(ctx, cli, name); err != nil {
		fmt.Printf("Warning: failed to remove container: %v\n", err)
	}

	if !opts.KeepVolume {
		volumeName := opts.VolumeName
		if volumeName != "" {
			if err := removeVolume(ctx, cli, volumeName); err != nil {
				fmt.Printf("Warning: failed to remove volume: %v\n", err)
			}
		}
	}

	return nil
}

func GetState(name string) (*types.ContainerState, error) {
	return state.LoadState(name)
}

func SaveState(name string, s *types.ContainerState) error {
	return state.SaveState(name, s)
}

func RemoveContainerDir(name string) error {
	return state.RemoveContainer(name)
}
