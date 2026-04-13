package sshconfig

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/IsmailAki/devc/pkg/types"
)

const (
	MarkerStart = "# === devc managed ==="
	MarkerEnd   = "# === end devc ==="
)

func GetSSHConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ssh", "config")
}

func GenerateEntry(state types.ContainerState, containerName string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Host %s\n", containerName))
	buf.WriteString("    HostName localhost\n")
	buf.WriteString(fmt.Sprintf("    Port %d\n", state.SSHPort))
	buf.WriteString("    User dev\n")
	buf.WriteString("    ForwardAgent yes\n")
	buf.WriteString("    StrictHostKeyChecking no\n")
	buf.WriteString("    UserKnownHostsFile /dev/null\n")
	return buf.String()
}

func AddEntry(state types.ContainerState, containerName string) error {
	configPath := GetSSHConfigPath()
	sshDir := filepath.Dir(configPath)

	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	content, err := readConfig(configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read ssh config: %w", err)
	}

	unmanaged, entries := splitManagedSection(string(content))
	entries[containerName] = strings.TrimSpace(GenerateEntry(state, containerName))

	if err := os.WriteFile(configPath, []byte(renderConfig(unmanaged, entries)), 0600); err != nil {
		return fmt.Errorf("failed to write ssh config: %w", err)
	}

	return nil
}

func RemoveEntry(containerName string) error {
	configPath := GetSSHConfigPath()

	content, err := readConfig(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read ssh config: %w", err)
	}

	unmanaged, entries := splitManagedSection(string(content))
	delete(entries, containerName)

	if err := os.WriteFile(configPath, []byte(renderConfig(unmanaged, entries)), 0600); err != nil {
		return fmt.Errorf("failed to write ssh config: %w", err)
	}

	return nil
}

func readConfig(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []byte{}, nil
	}
	return content, err
}

func splitManagedSection(content string) (string, map[string]string) {
	entries := make(map[string]string)
	start := strings.Index(content, MarkerStart)
	end := strings.Index(content, MarkerEnd)
	if start == -1 || end == -1 || end < start {
		return strings.TrimRight(content, "\n"), entries
	}

	unmanaged := strings.TrimRight(content[:start], "\n")
	managed := content[start+len(MarkerStart) : end]
	parseManagedEntries(managed, entries)
	return unmanaged, entries
}

func parseManagedEntries(content string, entries map[string]string) {
	lines := strings.Split(content, "\n")
	var current []string
	var currentHost string

	flush := func() {
		if currentHost == "" || len(current) == 0 {
			return
		}
		entries[currentHost] = strings.TrimSpace(strings.Join(current, "\n"))
		current = nil
		currentHost = ""
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "Host ") {
			flush()
			currentHost = strings.TrimSpace(strings.TrimPrefix(trimmed, "Host "))
		}
		if currentHost != "" {
			current = append(current, line)
		}
	}

	flush()
}

func renderConfig(unmanaged string, entries map[string]string) string {
	var buf bytes.Buffer
	if strings.TrimSpace(unmanaged) != "" {
		buf.WriteString(strings.TrimRight(unmanaged, "\n"))
		buf.WriteString("\n")
	}

	if len(entries) == 0 {
		return strings.TrimRight(buf.String(), "\n") + "\n"
	}

	if buf.Len() > 0 {
		buf.WriteString("\n")
	}
	buf.WriteString(MarkerStart + "\n")

	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)

	for i, name := range names {
		buf.WriteString(strings.TrimSpace(entries[name]))
		buf.WriteString("\n")
		if i < len(names)-1 {
			buf.WriteString("\n")
		}
	}

	buf.WriteString(MarkerEnd + "\n")
	return buf.String()
}
