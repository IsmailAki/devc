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

	initial := "Host github.com\n    User git\n\n" + MarkerStart + "\nHost existing\n    HostName localhost\n    Port 2222\n    User dev\n\nHost existing-root\n    HostName localhost\n    Port 2222\n    User root\n" + MarkerEnd + "\n"
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
	if !strings.Contains(text, "Host existing-root") {
		t.Fatal("expected existing root managed entry to be preserved")
	}
	if !strings.Contains(text, "Host new-container") {
		t.Fatal("expected new managed entry to be added")
	}
	if !strings.Contains(text, "Host new-container-root") {
		t.Fatal("expected new root managed entry to be added")
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

	initial := MarkerStart + "\nHost first\n    HostName localhost\n    Port 2222\n    User dev\n\nHost first-root\n    HostName localhost\n    Port 2222\n    User root\n\nHost second\n    HostName localhost\n    Port 2223\n    User dev\n\nHost second-root\n    HostName localhost\n    Port 2223\n    User root\n" + MarkerEnd + "\n"
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
	if strings.Contains(text, "Host first-root") {
		t.Fatal("expected removed root host to be absent")
	}
	if !strings.Contains(text, "Host second") {
		t.Fatal("expected other managed entry to remain")
	}
	if !strings.Contains(text, "Host second-root") {
		t.Fatal("expected other root managed entry to remain")
	}
	if !strings.Contains(text, MarkerStart) || !strings.Contains(text, MarkerEnd) {
		t.Fatal("expected managed block markers to remain")
	}
}

func TestGenerateEntryUsesRequestedUser(t *testing.T) {
	entry := GenerateEntry(types.ContainerState{SSHPort: 2222}, "example-root", "root")

	if !strings.Contains(entry, "Host example-root") {
		t.Fatal("expected host name to be rendered")
	}
	if !strings.Contains(entry, "User root") {
		t.Fatal("expected SSH user to be rendered")
	}
}
