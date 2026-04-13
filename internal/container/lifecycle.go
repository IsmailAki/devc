package container

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func createVolume(ctx context.Context, name string) error {
	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	_, err = cli.VolumeCreate(ctx, volume.CreateOptions{
		Name: name,
	})
	if err != nil {
		return fmt.Errorf("failed to create volume: %w", err)
	}

	return nil
}

func createContainer(ctx context.Context, cli dockerClient, name string, opts CreateOptions) (string, error) {
	env := make([]string, 0, len(opts.Env))
	for k, v := range opts.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	workingDir := opts.WorkingDir
	if workingDir == "" {
		workingDir = "/workspace"
	}

	containerConfig := &container.Config{
		Image: opts.Image,
		Env:   env,
		ExposedPorts: nat.PortSet{
			"22/tcp": struct{}{},
		},
		WorkingDir: workingDir,
		Tty:        true,
		OpenStdin:  true,
	}

	mounts := make([]mount.Mount, 0, 1)
	if opts.BindMountSource != "" {
		target := opts.BindMountTarget
		if target == "" {
			target = "/workspace"
		}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: opts.BindMountSource,
			Target: target,
		})
	} else {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeVolume,
			Source: opts.VolumeName,
			Target: "/workspace",
		})
	}

	if opts.DockerDataVolume != "" {
		if err := createVolume(ctx, opts.DockerDataVolume); err != nil {
			return "", fmt.Errorf("failed to create docker data volume: %w", err)
		}

		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeVolume,
			Source: opts.DockerDataVolume,
			Target: "/var/lib/docker",
		})
	}

	hostConfig := &container.HostConfig{
		Mounts:     mounts,
		Privileged: opts.Privileged,
		PortBindings: nat.PortMap{
			"22/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: fmt.Sprintf("%d", opts.Port),
				},
			},
		},
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, &network.NetworkingConfig{}, &v1.Platform{}, name)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	return resp.ID, nil
}

func startContainer(ctx context.Context, cli dockerClient, name string) error {
	return cli.ContainerStart(ctx, name, container.StartOptions{})
}

func stopContainer(ctx context.Context, cli dockerClient, name string) error {
	timeout := 30
	return cli.ContainerStop(ctx, name, container.StopOptions{Timeout: &timeout})
}

func removeContainer(ctx context.Context, cli dockerClient, name string) error {
	return cli.ContainerRemove(ctx, name, container.RemoveOptions{
		Force:         true,
		RemoveVolumes: false,
	})
}

func removeVolume(ctx context.Context, cli dockerClient, name string) error {
	return cli.VolumeRemove(ctx, name, true)
}

func waitForSSH(port int, timeout time.Duration) error {
	addr := fmt.Sprintf("localhost:%d", port)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for SSH on port %d", port)
}
