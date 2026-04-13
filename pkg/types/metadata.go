package types

import "time"

type ContainerMetadata struct {
	Name        string    `json:"name"`
	Repository  string    `json:"repository"`
	Owner       string    `json:"owner,omitempty"`
	Branch      string    `json:"branch,omitempty"`
	Languages   []string  `json:"languages"`
	Features    []string  `json:"features"`
	InitMode    string    `json:"init_mode"`
	ConfigPath  string    `json:"config_path"`
	ProjectPath string    `json:"project_path,omitempty"`
	SourcePath  string    `json:"source_path,omitempty"`
	DockerImage string    `json:"docker_image,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
