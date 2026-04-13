package ide

import (
	"fmt"

	"github.com/IsmailAki/devc/internal/state"
)

func ConnectJetBrains(containerName string) error {
	_, err := state.LoadState(containerName)
	if err != nil {
		return fmt.Errorf("failed to get container state: %w", err)
	}

	return fmt.Errorf("JetBrains Gateway connection requires manual setup")
}

func JetBrainsConnectionInfo(containerName string) (string, error) {
	containerState, err := state.LoadState(containerName)
	if err != nil {
		return "", fmt.Errorf("failed to get container state: %w", err)
	}

	metadata, err := state.LoadMetadata(containerName)
	if err != nil {
		return "", fmt.Errorf("failed to get container metadata: %w", err)
	}

	info := fmt.Sprintf(`JetBrains Gateway Connection Instructions:

1. Open JetBrains Gateway
2. Select "Connect via SSH"
3. Configure SSH connection:
   - Host: localhost
   - Port: %d
   - User: dev
   - Authentication: SSH key

4. Or use the SSH config entry already configured:
   - Host: %s

5. Select the IDE to use (IntelliJ IDEA, PyCharm, etc.)
6. Set the project directory to: %s

SSH Config Entry:
Host %s
    HostName localhost
    Port %d
    User dev
    ForwardAgent yes
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null

SSH Connection:
ssh %s

Or manually:
ssh -p %d dev@localhost
`, containerState.SSHPort, containerName, metadata.ProjectPath, containerName, containerState.SSHPort, containerName, containerState.SSHPort)

	return info, nil
}
