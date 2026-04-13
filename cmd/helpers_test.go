package cmd

import (
	"testing"

	"github.com/IsmailAki/devc/internal/container"
	"github.com/IsmailAki/devc/internal/naming"
	"github.com/IsmailAki/devc/pkg/types"
)

func TestConfigureContainerRuntimeEnablesDindForDockerFeature(t *testing.T) {
	opts := &container.CreateOptions{}
	configureContainerRuntime("devc-example-12345", []types.FeatureSpec{{Name: "docker"}}, opts)

	if !opts.Privileged {
		t.Fatal("expected docker feature to enable privileged mode")
	}

	if opts.DockerDataVolume != naming.GenerateDockerVolumeName("devc-example-12345") {
		t.Fatalf("unexpected docker data volume %q", opts.DockerDataVolume)
	}

	if opts.Env["DEVC_ENABLE_DIND"] != "1" {
		t.Fatalf("expected DEVC_ENABLE_DIND to be set, got %q", opts.Env["DEVC_ENABLE_DIND"])
	}

	if opts.Env["DOCKER_HOST"] != "unix:///var/run/docker.sock" {
		t.Fatalf("unexpected DOCKER_HOST %q", opts.Env["DOCKER_HOST"])
	}
}

func TestConfigureContainerRuntimeLeavesNonDockerContainersUnprivileged(t *testing.T) {
	opts := &container.CreateOptions{}
	configureContainerRuntime("devc-example-12345", []types.FeatureSpec{{Name: "go"}}, opts)

	if opts.Privileged {
		t.Fatal("did not expect privileged mode without docker feature")
	}

	if opts.DockerDataVolume != "" {
		t.Fatalf("expected no docker data volume, got %q", opts.DockerDataVolume)
	}

	if opts.Env != nil {
		t.Fatalf("expected no runtime env, got %#v", opts.Env)
	}
}
