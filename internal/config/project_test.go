package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectRootFromNestedDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".devc"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".devc", "devc.yml"), []byte("name: example\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	nested := filepath.Join(root, "src", "pkg")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := FindProjectRoot(nested)
	if err != nil {
		t.Fatalf("FindProjectRoot() error = %v", err)
	}
	if got != root {
		t.Fatalf("FindProjectRoot() = %q, want %q", got, root)
	}
}

func TestGenerateProjectConfigDeduplicatesFeatures(t *testing.T) {
	cfg := GenerateProjectConfig("example", "", "", []string{"javascript", "typescript", "csharp"})
	if len(cfg.Features) != 2 {
		t.Fatalf("GenerateProjectConfig() features = %d, want 2", len(cfg.Features))
	}
	if cfg.Features[0].Name != "node" {
		t.Fatalf("GenerateProjectConfig() first feature = %q, want node", cfg.Features[0].Name)
	}
	if cfg.Features[1].Name != "dotnet" {
		t.Fatalf("GenerateProjectConfig() second feature = %q, want dotnet", cfg.Features[1].Name)
	}
}
