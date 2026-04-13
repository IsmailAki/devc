package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/IsmailAki/devc/pkg/types"
)

const (
	ContainersDir    = "containers"
	StateFileName    = "state.json"
	MetadataFileName = "metadata.json"
	ConfigFileName   = "devc.yml"
)

func GetDevcDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".devc")
}

func GetContainersBaseDir() string {
	return filepath.Join(GetDevcDir(), ContainersDir)
}

func GetContainerDir(name string) string {
	return filepath.Join(GetContainersBaseDir(), name)
}

func GetStatePath(name string) string {
	return filepath.Join(GetContainerDir(name), StateFileName)
}

func GetMetadataPath(name string) string {
	return filepath.Join(GetContainerDir(name), MetadataFileName)
}

func GetConfigPath(name string) string {
	return filepath.Join(GetContainerDir(name), ConfigFileName)
}

func ensureContainerDir(name string) error {
	dir := GetContainerDir(name)
	return os.MkdirAll(dir, 0755)
}

func LoadState(name string) (*types.ContainerState, error) {
	path := GetStatePath(name)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("container %s not found", name)
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var s types.ContainerState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &s, nil
}

func SaveState(name string, state *types.ContainerState) error {
	if err := ensureContainerDir(name); err != nil {
		return fmt.Errorf("failed to create container directory: %w", err)
	}

	path := GetStatePath(name)

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

func LoadMetadata(name string) (*types.ContainerMetadata, error) {
	path := GetMetadataPath(name)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("metadata for container %s not found", name)
		}
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var m types.ContainerMetadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse metadata file: %w", err)
	}

	return &m, nil
}

func SaveMetadata(name string, metadata *types.ContainerMetadata) error {
	if err := ensureContainerDir(name); err != nil {
		return fmt.Errorf("failed to create container directory: %w", err)
	}

	path := GetMetadataPath(name)

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

func LoadConfig(name string) ([]byte, error) {
	path := GetConfigPath(name)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config for container %s not found", name)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return data, nil
}

func SaveConfig(name string, configData []byte) error {
	if err := ensureContainerDir(name); err != nil {
		return fmt.Errorf("failed to create container directory: %w", err)
	}

	path := GetConfigPath(name)

	if err := os.WriteFile(path, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func ContainerExists(name string) bool {
	path := GetStatePath(name)
	_, err := os.Stat(path)
	return err == nil
}

func ListContainers() ([]string, error) {
	baseDir := GetContainersBaseDir()

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read containers directory: %w", err)
	}

	var containers []string
	for _, entry := range entries {
		if entry.IsDir() {
			statePath := filepath.Join(baseDir, entry.Name(), StateFileName)
			if _, err := os.Stat(statePath); err == nil {
				containers = append(containers, entry.Name())
			}
		}
	}

	return containers, nil
}

func RemoveContainer(name string) error {
	dir := GetContainerDir(name)
	return os.RemoveAll(dir)
}
