package container

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	container_types "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type dockerClient interface {
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, name string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error)
	VolumeCreate(ctx context.Context, options volume.CreateOptions) (volume.Volume, error)
	VolumeRemove(ctx context.Context, volumeID string, force bool) error
	ContainerExecCreate(ctx context.Context, container string, config container_types.ExecOptions) (container.ExecCreateResponse, error)
	ContainerExecStart(ctx context.Context, execID string, config container.ExecStartOptions) error
	ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error)
}

func getDockerClient() (dockerClient, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}
