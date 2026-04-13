package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/IsmailAki/devc/internal/state"
	"github.com/IsmailAki/devc/pkg/types"
	"gopkg.in/yaml.v3"
)

const ProjectConfigDir = ".devc"
const ProjectConfigFile = "devc.yml"

func GetProjectConfigPath(workDir string) string {
	return filepath.Join(workDir, ProjectConfigDir, ProjectConfigFile)
}

func LoadProjectConfig(workDir string) (*types.ProjectConfig, error) {
	if workDir == "" {
		projectRoot, err := FindProjectRoot("")
		if err != nil {
			return nil, err
		}
		workDir = projectRoot
	}

	path := GetProjectConfigPath(workDir)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("devc.yml not found in %s", workDir)
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg types.ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func SaveProjectConfig(cfg *types.ProjectConfig, workDir string) error {
	if workDir == "" {
		cwd, _ := os.Getwd()
		workDir = cwd
	}

	configDir := filepath.Join(workDir, ProjectConfigDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	path := GetProjectConfigPath(workDir)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func SaveGlobalConfig(containerName string, cfg *types.ProjectConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return state.SaveConfig(containerName, data)
}

func LoadGlobalConfig(containerName string) (*types.ProjectConfig, error) {
	data, err := state.LoadConfig(containerName)
	if err != nil {
		return nil, err
	}

	var cfg types.ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func LoadConfigFromPath(path string) (*types.ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config from %s: %w", path, err)
	}

	var cfg types.ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func ResolveConfig(containerName string, overridePath string) (config *types.ProjectConfig, configPath string, err error) {
	if overridePath != "" {
		cfg, err := LoadConfigFromPath(overridePath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to load override config: %w", err)
		}
		return cfg, overridePath, nil
	}

	if projectRoot, err := FindProjectRoot(""); err == nil {
		localConfigPath := GetProjectConfigPath(projectRoot)
		cfg, err := LoadConfigFromPath(localConfigPath)
		if err != nil {
			return nil, "", err
		}
		return cfg, localConfigPath, nil
	}

	storedConfig, err := LoadGlobalConfig(containerName)
	if err == nil {
		return storedConfig, state.GetConfigPath(containerName), nil
	}

	return nil, "", fmt.Errorf("no config found for container %s", containerName)
}

func CopyOverrideConfig(containerName string, overridePath string) error {
	data, err := os.ReadFile(overridePath)
	if err != nil {
		return fmt.Errorf("failed to read override config: %w", err)
	}

	return state.SaveConfig(containerName, data)
}

func IsInProject(workDir string) bool {
	if workDir == "" {
		_, err := FindProjectRoot("")
		return err == nil
	}

	path := GetProjectConfigPath(workDir)
	_, err := os.Stat(path)
	return err == nil
}

func FindProjectRoot(startDir string) (string, error) {
	if startDir == "" {
		cwd, _ := os.Getwd()
		startDir = cwd
	}

	dir := startDir
	for {
		path := GetProjectConfigPath(dir)
		if _, err := os.Stat(path); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("devc.yml not found in any parent directory")
		}
		dir = parent
	}
}

func GenerateProjectConfig(name string, gitURL string, gitBranch string, languages []string) *types.ProjectConfig {
	features := make([]types.FeatureSpec, 0, len(languages))
	seen := make(map[string]struct{}, len(languages))

	for _, lang := range languages {
		feature := languageToFeature(lang)
		if feature != "" {
			if _, exists := seen[feature]; exists {
				continue
			}
			seen[feature] = struct{}{}
			features = append(features, types.FeatureSpec{
				Name: feature,
			})
		}
	}

	cfg := &types.ProjectConfig{
		Name:     name,
		Features: features,
	}

	if gitURL != "" {
		cfg.Git = &types.GitConfig{
			URL:    gitURL,
			Branch: gitBranch,
		}
	}

	return cfg
}

func languageToFeature(lang string) string {
	mapping := map[string]string{
		"javascript": "node",
		"typescript": "node",
		"go":         "go",
		"python":     "python",
		"rust":       "rust",
		"java":       "java",
		"kotlin":     "java",
		"scala":      "java",
		"ruby":       "ruby",
		"csharp":     "dotnet",
	}

	if feature, ok := mapping[lang]; ok {
		return feature
	}
	return ""
}
