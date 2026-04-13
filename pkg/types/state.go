package types

import "time"

type ContainerState struct {
	ContainerID     string    `json:"container_id"`
	Image           string    `json:"image"`
	SSHPort         int       `json:"ssh_port"`
	WorkspaceVolume string    `json:"workspace_volume"`
	DockerVolume    string    `json:"docker_volume,omitempty"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type State struct {
	Version    string                    `json:"version"`
	Containers map[string]ContainerState `json:"containers"`
}

func NewState() *State {
	return &State{
		Version:    "2",
		Containers: make(map[string]ContainerState),
	}
}
