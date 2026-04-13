package types

type FeatureSpec struct {
	Name    string                 `yaml:"name"`
	Version string                 `yaml:"version,omitempty"`
	Params  map[string]interface{} `yaml:"params,omitempty"`
}

type GitConfig struct {
	URL    string `yaml:"url"`
	Branch string `yaml:"branch,omitempty"`
}

type ProjectConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Features    []FeatureSpec     `yaml:"features,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	Ports       []int             `yaml:"ports,omitempty"`
	Git         *GitConfig        `yaml:"git,omitempty"`
}
