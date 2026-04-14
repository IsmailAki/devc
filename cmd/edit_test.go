package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/IsmailAki/devc/internal/config"
	"github.com/IsmailAki/devc/internal/state"
	"github.com/IsmailAki/devc/pkg/types"
)

func TestResolveEditableConfigTargetPrefersLocalConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	projectRoot := filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	containerName := "devc-example-12345"
	localCfg := &types.ProjectConfig{Name: "local", Features: []types.FeatureSpec{{Name: "node"}}}
	globalCfg := &types.ProjectConfig{Name: "global", Features: []types.FeatureSpec{{Name: "python"}}}

	if err := config.SaveProjectConfig(localCfg, projectRoot); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveGlobalConfig(containerName, globalCfg); err != nil {
		t.Fatal(err)
	}
	if err := state.SaveMetadata(containerName, &types.ContainerMetadata{
		Name:       containerName,
		Repository: "example",
		InitMode:   "local",
		SourcePath: projectRoot,
	}); err != nil {
		t.Fatal(err)
	}

	target, cfg, err := resolveEditableConfigTarget(containerName)
	if err != nil {
		t.Fatalf("resolveEditableConfigTarget() error = %v", err)
	}

	if !target.UseLocal {
		t.Fatal("expected local config to be selected")
	}

	localPath := config.GetProjectConfigPath(projectRoot)
	if target.ActivePath != localPath {
		t.Fatalf("resolveEditableConfigTarget() path = %q, want %q", target.ActivePath, localPath)
	}

	if !reflect.DeepEqual(cfg, localCfg) {
		t.Fatalf("resolveEditableConfigTarget() config = %#v, want %#v", cfg, localCfg)
	}
}

func TestResolveEditableConfigTargetFallsBackToGlobalConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	projectRoot := filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	containerName := "devc-example-12345"
	globalCfg := &types.ProjectConfig{Name: "global", Features: []types.FeatureSpec{{Name: "python"}}}

	if err := config.SaveGlobalConfig(containerName, globalCfg); err != nil {
		t.Fatal(err)
	}
	if err := state.SaveMetadata(containerName, &types.ContainerMetadata{
		Name:       containerName,
		Repository: "example",
		InitMode:   "local",
		SourcePath: projectRoot,
	}); err != nil {
		t.Fatal(err)
	}

	target, cfg, err := resolveEditableConfigTarget(containerName)
	if err != nil {
		t.Fatalf("resolveEditableConfigTarget() error = %v", err)
	}

	if target.UseLocal {
		t.Fatal("did not expect local config to be selected")
	}

	globalPath := state.GetConfigPath(containerName)
	if target.ActivePath != globalPath {
		t.Fatalf("resolveEditableConfigTarget() path = %q, want %q", target.ActivePath, globalPath)
	}

	if !reflect.DeepEqual(cfg, globalCfg) {
		t.Fatalf("resolveEditableConfigTarget() config = %#v, want %#v", cfg, globalCfg)
	}
}

func TestEditableConfigTargetSaveSyncsLocalAndGlobal(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	projectRoot := filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	containerName := "devc-example-12345"
	initialCfg := &types.ProjectConfig{Name: "example", Features: []types.FeatureSpec{{Name: "node"}}}
	updatedCfg := &types.ProjectConfig{Name: "example", Features: []types.FeatureSpec{{Name: "go"}}}

	if err := config.SaveProjectConfig(initialCfg, projectRoot); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveGlobalConfig(containerName, initialCfg); err != nil {
		t.Fatal(err)
	}
	if err := state.SaveMetadata(containerName, &types.ContainerMetadata{
		Name:       containerName,
		Repository: "example",
		InitMode:   "local",
		SourcePath: projectRoot,
	}); err != nil {
		t.Fatal(err)
	}

	target, _, err := resolveEditableConfigTarget(containerName)
	if err != nil {
		t.Fatalf("resolveEditableConfigTarget() error = %v", err)
	}

	if err := target.Save(updatedCfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	localSaved, err := config.LoadProjectConfig(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	globalSaved, err := config.LoadGlobalConfig(containerName)
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := state.LoadMetadata(containerName)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(localSaved, updatedCfg) {
		t.Fatalf("local config = %#v, want %#v", localSaved, updatedCfg)
	}
	if !reflect.DeepEqual(globalSaved, updatedCfg) {
		t.Fatalf("global config = %#v, want %#v", globalSaved, updatedCfg)
	}

	localPath := config.GetProjectConfigPath(projectRoot)
	if metadata.ConfigPath != localPath {
		t.Fatalf("metadata config path = %q, want %q", metadata.ConfigPath, localPath)
	}
}

func TestMergeSelectedFeaturesPreservesExistingOrder(t *testing.T) {
	current := []types.FeatureSpec{{Name: "node"}, {Name: "python"}}
	selected := []string{"go", "node"}

	merged := mergeSelectedFeatures(current, selected)
	if len(merged) != 2 {
		t.Fatalf("mergeSelectedFeatures() len = %d, want 2", len(merged))
	}
	if merged[0].Name != "node" {
		t.Fatalf("mergeSelectedFeatures() first = %q, want node", merged[0].Name)
	}
	if merged[1].Name != "go" {
		t.Fatalf("mergeSelectedFeatures() second = %q, want go", merged[1].Name)
	}
}
