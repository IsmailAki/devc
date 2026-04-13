package sshconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/IsmailAki/devc/pkg/types"
)

func TestAddEntryPreservesManagedEntries(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	configPath := filepath.Join(tmpDir, ".ssh", "config")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatal(err)
	}

	initial := "Host github.com\n    User git\n\n" + MarkerStart + "\nHost existing\n    HostName localhost\n    Port 2222\n    User dev\n" + MarkerEnd + "\n"
	if err := os.WriteFile(configPath, []byte(initial), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := AddEntry(types.ContainerState{SSHPort: 2223}, "new-container"); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)

	if !strings.Contains(text, "Host github.com") {
		t.Fatal("expected unmanaged config to be preserved")
	}
	if !strings.Contains(text, "Host existing") {
		t.Fatal("expected existing managed entry to be preserved")
	}
	if !strings.Contains(text, "Host new-container") {
		t.Fatal("expected new managed entry to be added")
	}
	if strings.Count(text, MarkerStart) != 1 || strings.Count(text, MarkerEnd) != 1 {
		t.Fatal("expected exactly one managed block")
	}
}

func TestRemoveEntryKeepsOtherManagedEntries(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	configPath := filepath.Join(tmpDir, ".ssh", "config")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatal(err)
	}

	initial := MarkerStart + "\nHost first\n    HostName localhost\n    Port 2222\n    User dev\n\nHost second\n    HostName localhost\n    Port 2223\n    User dev\n" + MarkerEnd + "\n"
	if err := os.WriteFile(configPath, []byte(initial), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := RemoveEntry("first"); err != nil {
		t.Fatalf("RemoveEntry() error = %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)

	if strings.Contains(text, "Host first") {
		t.Fatal("expected removed host to be absent")
	}
	if !strings.Contains(text, "Host second") {
		t.Fatal("expected other managed entry to remain")
	}
	if !strings.Contains(text, MarkerStart) || !strings.Contains(text, MarkerEnd) {
		t.Fatal("expected managed block markers to remain")
	}
}
