package ide

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/IsmailAki/devc/internal/state"
)

func ConnectVSCode(containerName string) error {
	_, err := state.LoadState(containerName)
	if err != nil {
		return fmt.Errorf("failed to get container state: %w", err)
	}

	metadata, err := state.LoadMetadata(containerName)
	if err != nil {
		return fmt.Errorf("failed to get container metadata: %w", err)
	}

	cmd := exec.Command("code", "--remote", fmt.Sprintf("ssh-remote+%s", containerName), metadata.ProjectPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch VS Code: %w", err)
	}

	return nil
}

func VSCodeConnectionInfo(containerName string) (string, error) {
	containerState, err := state.LoadState(containerName)
	if err != nil {
		return "", fmt.Errorf("failed to get container state: %w", err)
	}

	metadata, err := state.LoadMetadata(containerName)
	if err != nil {
		return "", fmt.Errorf("failed to get container metadata: %w", err)
	}

	info := fmt.Sprintf(`VS Code Connection Instructions:

1. Install the "Remote - SSH" extension in VS Code
2. Run the following command:
   code --remote ssh-remote+%s %s

3. Or connect manually:
   - Open VS Code
   - Press Cmd+Shift+P (Mac) or Ctrl+Shift+P (Windows/Linux)
   - Type "Remote-SSH: Connect to Host"
   - Select "%s"
   - Open folder: %s

SSH Config Entry:
Host %s
    HostName localhost
    Port %d
    User dev
    ForwardAgent yes
`, containerName, metadata.ProjectPath, containerName, metadata.ProjectPath, containerName, containerState.SSHPort)

	return info, nil
}
