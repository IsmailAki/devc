package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/IsmailAki/devc/internal/config"
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
