package types

type Feature struct {
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	Version     string              `yaml:"version"`
	Depends     []string            `yaml:"depends,omitempty"`
	Params      map[string]ParamDef `yaml:"params,omitempty"`
	Install     []InstallStep       `yaml:"install,omitempty"`
}

type ParamDef struct {
	Description string `yaml:"description"`
	Default     string `yaml:"default"`
}

type InstallStep struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}
