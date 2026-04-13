package feature

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/IsmailAki/devc/pkg/types"
	"gopkg.in/yaml.v3"
)

//go:embed embed
var embeddedFeatures embed.FS

type Registry struct {
	features map[string]types.Feature
	mu       sync.RWMutex
}

func NewRegistry() *Registry {
	r := &Registry{
		features: make(map[string]types.Feature),
	}

	r.loadBuiltinFeatures()
	_ = r.LoadUserFeatures()

	return r
}

func (r *Registry) Register(f types.Feature) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.features[f.Name] = f
}

func (r *Registry) Get(name string) (*types.Feature, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	f, ok := r.features[name]
	if !ok {
		return nil, false
	}
	return &f, true
}

func (r *Registry) List() []types.Feature {
	r.mu.RLock()
	defer r.mu.RUnlock()

	features := make([]types.Feature, 0, len(r.features))
	for _, f := range r.features {
		features = append(features, f)
	}
	return features
}

func (r *Registry) loadBuiltinFeatures() {
	builtinFeatures := []string{
		"node",
		"go",
		"python",
		"rust",
		"docker",
		"terraform",
		"java",
		"ruby",
		"dotnet",
		"opencode",
		"claude-code",
	}

	for _, name := range builtinFeatures {
		if f, err := r.loadEmbeddedFeature(name); err == nil {
			r.Register(*f)
		}
	}
}

func (r *Registry) loadEmbeddedFeature(name string) (*types.Feature, error) {
	featureYamlPath := filepath.Join("embed", name, "feature.yml")
	data, err := embeddedFeatures.ReadFile(featureYamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read feature %s: %w", name, err)
	}

	var f types.Feature
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("failed to parse feature %s: %w", name, err)
	}

	return &f, nil
}

func GetFeaturesDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".devc", "features")
}

func (r *Registry) LoadUserFeatures() error {
	featuresDir := GetFeaturesDir()

	entries, err := os.ReadDir(featuresDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read features directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		featurePath := filepath.Join(featuresDir, name, "feature.yml")

		data, err := os.ReadFile(featurePath)
		if err != nil {
			continue
		}

		var f types.Feature
		if err := yaml.Unmarshal(data, &f); err != nil {
			continue
		}

		r.Register(f)
	}

	return nil
}
