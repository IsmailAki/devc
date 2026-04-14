package container

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
)

type ExecOptions struct {
	Cmd        []string
	User       string
	WorkingDir string
	Env        []string
}

func GetSSHPublicKey() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	keyPaths := []string{
		filepath.Join(home, ".ssh", "id_ed25519.pub"),
		filepath.Join(home, ".ssh", "id_rsa.pub"),
		filepath.Join(home, ".ssh", "id_ecdsa.pub"),
	}

	for _, path := range keyPaths {
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data), nil
		}
	}

	return "", fmt.Errorf("no SSH public key found in ~/.ssh/")
}

func Exec(ctx context.Context, containerName string, opts ExecOptions) error {
	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	execConfig := container.ExecOptions{
		Cmd:          opts.Cmd,
		User:         opts.User,
		WorkingDir:   opts.WorkingDir,
		Env:          opts.Env,
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := cli.ContainerExecCreate(ctx, containerName, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	err = cli.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start exec: %w", err)
	}

	for {
		inspect, err := cli.ContainerExecInspect(ctx, execResp.ID)
		if err != nil {
			return fmt.Errorf("failed to inspect exec: %w", err)
		}
		if !inspect.Running {
			if inspect.ExitCode != 0 {
				return fmt.Errorf("command exited with code %d", inspect.ExitCode)
			}
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func PrepareWorkspace(ctx context.Context, containerName, workspacePath string) error {
	quotedPath := shellQuote(workspacePath)
	cmd := []string{"sh", "-lc", fmt.Sprintf("mkdir -p %s && chown -R dev:dev %s", quotedPath, quotedPath)}
	return Exec(ctx, containerName, ExecOptions{Cmd: cmd})
}

func SetupSSHKey(ctx context.Context, containerName string) error {
	publicKey, err := GetSSHPublicKey()
	if err != nil {
		return fmt.Errorf("failed to get SSH public key: %w", err)
	}

	cmds := [][]string{
		{"mkdir", "-p", "/home/dev/.ssh"},
		{"mkdir", "-p", "/root/.ssh"},
		{"chmod", "700", "/home/dev/.ssh"},
		{"chmod", "700", "/root/.ssh"},
		{"sh", "-c", fmt.Sprintf("echo '%s' > /home/dev/.ssh/authorized_keys", publicKey)},
		{"sh", "-c", fmt.Sprintf("echo '%s' > /root/.ssh/authorized_keys", publicKey)},
		{"chmod", "600", "/home/dev/.ssh/authorized_keys"},
		{"chmod", "600", "/root/.ssh/authorized_keys"},
		{"chown", "-R", "dev:dev", "/home/dev/.ssh"},
		{"chown", "-R", "root:root", "/root/.ssh"},
	}

	for _, cmd := range cmds {
		if err := Exec(ctx, containerName, ExecOptions{Cmd: cmd}); err != nil {
			return fmt.Errorf("failed to run %v: %w", cmd, err)
		}
	}

	return nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
