package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/IsmailAki/devc/internal/config"
	"github.com/IsmailAki/devc/internal/container"
	"github.com/IsmailAki/devc/internal/naming"
	"github.com/IsmailAki/devc/internal/state"
	"github.com/IsmailAki/devc/pkg/types"
)

type editableConfigTarget struct {
	ContainerName string
	Metadata      *types.ContainerMetadata
	LocalPath     string
	GlobalPath    string
	ActivePath    string
	UseLocal      bool
}

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

	containers, err := loadContainerInfos(true)
	if err != nil {
		return ""
	}

	for _, container := range containers {
		metadata := container.Metadata
		if metadata == nil || metadata.InitMode != "local" {
			continue
		}

		if normalizePath(metadata.SourcePath) == projectRoot {
			return container.Name
		}

		if normalizePath(metadata.ProjectPath) == projectRoot {
			return container.Name
		}

		if normalizePath(filepath.Dir(metadata.ConfigPath)) == filepath.Join(projectRoot, config.ProjectConfigDir) {
			return container.Name
		}
	}

	return ""
}

func firstContainerName() string {
	containers, err := loadContainerInfos(true)
	if err != nil || len(containers) == 0 {
		return ""
	}

	if len(containers) == 1 {
		return containers[0].Name
	}

	for _, container := range containers {
		if container.State != nil && container.State.Status == "running" {
			return container.Name
		}
	}

	return containers[0].Name
}

func loadContainerInfos(includeStopped bool) ([]containerInfo, error) {
	containers, err := state.ListContainers()
	if err != nil {
		return nil, err
	}

	filtered := make([]containerInfo, 0, len(containers))
	for _, name := range containers {
		containerState, err := state.LoadState(name)
		if err != nil {
			continue
		}

		metadata, err := state.LoadMetadata(name)
		if err != nil {
			continue
		}

		if includeStopped || containerState.Status == "running" {
			filtered = append(filtered, containerInfo{
				Name:     name,
				State:    containerState,
				Metadata: metadata,
			})
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].State.Status == filtered[j].State.Status {
			return filtered[i].Name < filtered[j].Name
		}
		return filtered[i].State.Status == "running"
	})

	return filtered, nil
}

func resolveEditableConfigTarget(containerName string) (*editableConfigTarget, *types.ProjectConfig, error) {
	metadata, err := state.LoadMetadata(containerName)
	if err != nil {
		return nil, nil, err
	}

	target := &editableConfigTarget{
		ContainerName: containerName,
		Metadata:      metadata,
		GlobalPath:    state.GetConfigPath(containerName),
	}

	if metadata.InitMode == "local" && metadata.SourcePath != "" {
		target.LocalPath = config.GetProjectConfigPath(metadata.SourcePath)
		if _, err := os.Stat(target.LocalPath); err == nil {
			cfg, err := config.LoadConfigFromPath(target.LocalPath)
			if err != nil {
				return nil, nil, err
			}
			target.ActivePath = target.LocalPath
			target.UseLocal = true
			return target, cfg, nil
		}
	}

	cfg, err := config.LoadGlobalConfig(containerName)
	if err != nil {
		return nil, nil, err
	}

	target.ActivePath = target.GlobalPath
	return target, cfg, nil
}

func (t *editableConfigTarget) Save(cfg *types.ProjectConfig) error {
	if t.UseLocal {
		if err := config.SaveProjectConfig(cfg, t.Metadata.SourcePath); err != nil {
			return err
		}
		if err := config.SaveGlobalConfig(t.ContainerName, cfg); err != nil {
			return err
		}
		t.ActivePath = t.LocalPath
	} else {
		if err := config.SaveGlobalConfig(t.ContainerName, cfg); err != nil {
			return err
		}
		t.ActivePath = t.GlobalPath
	}

	if t.Metadata != nil {
		t.Metadata.ConfigPath = t.ActivePath
		if err := state.SaveMetadata(t.ContainerName, t.Metadata); err != nil {
			return err
		}
	}

	return nil
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
