package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/IsmailAki/devc/internal/config"
	"github.com/IsmailAki/devc/internal/container"
	"github.com/IsmailAki/devc/internal/naming"
	"github.com/IsmailAki/devc/internal/state"
	"github.com/IsmailAki/devc/pkg/types"
)

func resolveCurrentProject() (string, *types.ProjectConfig, error) {
	projectRoot, err := config.FindProjectRoot("")
	if err != nil {
		return "", nil, err
	}

	cfg, err := config.LoadProjectConfig(projectRoot)
	if err != nil {
		return "", nil, err
	}

	return projectRoot, cfg, nil
}

func resolveLocalContainerName(projectRoot string, cfg *types.ProjectConfig) string {
	if name := findLocalContainerByProject(projectRoot); name != "" {
		return name
	}

	return naming.GenerateLocalContainerName(cfg.Name, projectRoot)
}

func findLocalContainerByProject(projectRoot string) string {
	projectRoot = normalizePath(projectRoot)

	containers, err := state.ListContainers()
	if err != nil {
		return ""
	}

	for _, name := range containers {
		metadata, err := state.LoadMetadata(name)
		if err != nil || metadata.InitMode != "local" {
			continue
		}

		if normalizePath(metadata.SourcePath) == projectRoot {
			return name
		}

		if normalizePath(metadata.ProjectPath) == projectRoot {
			return name
		}

		if normalizePath(filepath.Dir(metadata.ConfigPath)) == filepath.Join(projectRoot, config.ProjectConfigDir) {
			return name
		}
	}

	return ""
}

func firstContainerName() string {
	containers, err := state.ListContainers()
	if err != nil || len(containers) == 0 {
		return ""
	}

	if len(containers) == 1 {
		return containers[0]
	}

	for _, name := range containers {
		st, err := state.LoadState(name)
		if err != nil {
			continue
		}
		if st.Status == "running" {
			return name
		}
	}

	return containers[0]
}

func resolveDefaultContainerName() string {
	projectRoot, cfg, err := resolveCurrentProject()
	if err == nil {
		return resolveLocalContainerName(projectRoot, cfg)
	}

	return firstContainerName()
}

func localMetadata(containerName, projectRoot, remoteProjectPath string, cfg *types.ProjectConfig) *types.ContainerMetadata {
	featureNames := make([]string, len(cfg.Features))
	for i, f := range cfg.Features {
		featureNames[i] = f.Name
	}

	return &types.ContainerMetadata{
		Name:        containerName,
		Repository:  cfg.Name,
		Features:    featureNames,
		InitMode:    "local",
		ConfigPath:  config.GetProjectConfigPath(projectRoot),
		ProjectPath: remoteProjectPath,
		SourcePath:  projectRoot,
	}
}

func remoteProjectPath(projectRoot string) string {
	return path.Join("/workspace", filepath.Base(projectRoot))
}

func normalizePath(value string) string {
	if value == "" {
		return ""
	}

	abs, err := filepath.Abs(value)
	if err != nil {
		return filepath.Clean(value)
	}

	return filepath.Clean(abs)
}

func ensureProjectName(cfg *types.ProjectConfig, projectRoot string) error {
	if cfg.Name != "" {
		return nil
	}

	cfg.Name = filepath.Base(projectRoot)
	if cfg.Name == "." || cfg.Name == string(filepath.Separator) {
		return fmt.Errorf("project name is empty")
	}

	return nil
}

func currentWorkingDir() string {
	cwd, _ := os.Getwd()
	return cwd
}

func hasFeature(features []types.FeatureSpec, name string) bool {
	for _, feature := range features {
		if feature.Name == name {
			return true
		}
	}

	return false
}

func mergeEnv(base map[string]string, additions map[string]string) map[string]string {
	if len(base) == 0 && len(additions) == 0 {
		return nil
	}

	merged := make(map[string]string, len(base)+len(additions))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range additions {
		merged[k] = v
	}

	return merged
}

func localUserEnv() map[string]string {
	uid := os.Getuid()
	gid := os.Getgid()
	if uid <= 0 || gid <= 0 {
		return nil
	}

	return map[string]string{
		"DEVC_UID": strconv.Itoa(uid),
		"DEVC_GID": strconv.Itoa(gid),
	}
}

func configureContainerRuntime(containerName string, features []types.FeatureSpec, opts *container.CreateOptions) {
	if opts.BindMountSource != "" {
		opts.Env = mergeEnv(opts.Env, localUserEnv())
	}

	if !hasFeature(features, "docker") {
		return
	}

	opts.Privileged = true
	opts.DockerDataVolume = naming.GenerateDockerVolumeName(containerName)
	opts.Env = mergeEnv(opts.Env, map[string]string{
		"DEVC_ENABLE_DIND": "1",
		"DOCKER_BUILDKIT":  "1",
		"DOCKER_HOST":      "unix:///var/run/docker.sock",
	})
}
