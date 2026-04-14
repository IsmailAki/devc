package cmd

import (
	"testing"
	"time"

	"github.com/IsmailAki/devc/internal/state"
	"github.com/IsmailAki/devc/pkg/types"
)

func TestResolveDestroyContainerNameUsesArgument(t *testing.T) {
	got, err := resolveDestroyContainerName([]string{"devc-example"})
	if err != nil {
		t.Fatalf("resolveDestroyContainerName() error = %v", err)
	}
	if got != "devc-example" {
		t.Fatalf("resolveDestroyContainerName() = %q, want %q", got, "devc-example")
	}
}

func TestResolveDestroyContainerNameUsesSingleContainer(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	if err := saveTestContainer("devc-example", "running"); err != nil {
		t.Fatal(err)
	}

	got, err := resolveDestroyContainerName(nil)
	if err != nil {
		t.Fatalf("resolveDestroyContainerName() error = %v", err)
	}
	if got != "devc-example" {
		t.Fatalf("resolveDestroyContainerName() = %q, want %q", got, "devc-example")
	}
}

func TestResolveRebuildContainerNameUsesArgument(t *testing.T) {
	got, err := resolveRebuildContainerName([]string{"devc-example"})
	if err != nil {
		t.Fatalf("resolveRebuildContainerName() error = %v", err)
	}
	if got != "devc-example" {
		t.Fatalf("resolveRebuildContainerName() = %q, want %q", got, "devc-example")
	}
}

func TestResolveRebuildContainerNameUsesSingleContainer(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	if err := saveTestContainer("devc-example", "stopped"); err != nil {
		t.Fatal(err)
	}

	got, err := resolveRebuildContainerName(nil)
	if err != nil {
		t.Fatalf("resolveRebuildContainerName() error = %v", err)
	}
	if got != "devc-example" {
		t.Fatalf("resolveRebuildContainerName() = %q, want %q", got, "devc-example")
	}
}

func saveTestContainer(name, status string) error {
	if err := state.SaveState(name, &types.ContainerState{
		ContainerID: name + "-id",
		Status:      status,
		CreatedAt:   time.Now(),
	}); err != nil {
		return err
	}

	return state.SaveMetadata(name, &types.ContainerMetadata{
		Name:       name,
		Repository: name,
		InitMode:   "github",
		CreatedAt:  time.Now(),
	})
}
