package feature

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/IsmailAki/devc/pkg/types"
)

const DockerfileTemplate = `FROM devc-base:latest

# Install features
%s

# Cleanup
RUN rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

WORKDIR /workspace
`

func GenerateDockerfile(features []types.FeatureSpec, registry *Registry) (string, error) {
	ordered, err := resolveDependencies(features, registry)
	if err != nil {
		return "", err
	}

	var commands []string

	for _, spec := range ordered {
		feature, ok := registry.Get(spec.Name)
		if !ok {
			return "", fmt.Errorf("feature %s not found", spec.Name)
		}

		featureCommands, err := generateInstallCommands(spec, *feature)
		if err != nil {
			return "", fmt.Errorf("failed to generate commands for %s: %w", spec.Name, err)
		}

		commands = append(commands, featureCommands...)
	}

	installSection := strings.Join(commands, "\n")
	return fmt.Sprintf(DockerfileTemplate, installSection), nil
}

func generateInstallCommands(spec types.FeatureSpec, feature types.Feature) ([]string, error) {
	if len(feature.Install) == 0 {
		return nil, fmt.Errorf("feature %s has no install steps defined", spec.Name)
	}

	params := buildTemplateParams(spec, feature)

	var commands []string
	for _, step := range feature.Install {
		rendered, err := renderTemplate(step.Run, params)
		if err != nil {
			return nil, fmt.Errorf("failed to render step %q: %w", step.Name, err)
		}
		escaped := escapeForDockerfile(rendered)
		commands = append(commands, fmt.Sprintf("RUN %s", escaped))
	}

	return commands, nil
}

func escapeForDockerfile(script string) string {
	lines := strings.Split(script, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		cleanedLines = append(cleanedLines, trimmed)
	}
	if len(cleanedLines) == 0 {
		return ""
	}
	if len(cleanedLines) == 1 {
		return cleanedLines[0]
	}
	scriptContent := strings.Join(cleanedLines, "\n")
	encoded := base64.StdEncoding.EncodeToString([]byte(scriptContent))
	return "echo '" + encoded + "' | base64 -d > /tmp/install.sh && sh /tmp/install.sh && rm /tmp/install.sh"
}

func buildTemplateParams(spec types.FeatureSpec, feature types.Feature) map[string]interface{} {
	params := make(map[string]interface{})

	version := spec.Version
	if version == "" {
		if v, ok := feature.Params["version"]; ok {
			version = v.Default
		}
	}
	params["Version"] = version

	for name, value := range spec.Params {
		params[strings.Title(name)] = value
	}

	for name, param := range feature.Params {
		key := strings.Title(name)
		if _, exists := params[key]; !exists {
			params[key] = param.Default
		}
	}

	return params
}

func renderTemplate(tmpl string, params map[string]interface{}) (string, error) {
	t, err := template.New("install").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	if err := t.Execute(&result, params); err != nil {
		return "", err
	}

	return result.String(), nil
}

func resolveDependencies(features []types.FeatureSpec, registry *Registry) ([]types.FeatureSpec, error) {
	graph := make(map[string][]string)
	specMap := make(map[string]types.FeatureSpec)

	for _, spec := range features {
		specMap[spec.Name] = spec

		feature, ok := registry.Get(spec.Name)
		if !ok {
			return nil, fmt.Errorf("feature %s not found", spec.Name)
		}

		graph[spec.Name] = feature.Depends
	}

	var result []types.FeatureSpec
	inDegree := make(map[string]int)

	for name := range graph {
		inDegree[name] = 0
	}

	for _, deps := range graph {
		for _, dep := range deps {
			if _, exists := inDegree[dep]; !exists {
				if _, ok := registry.Get(dep); ok {
					feature, _ := registry.Get(dep)
					specMap[dep] = types.FeatureSpec{Name: dep}
					graph[dep] = feature.Depends
					inDegree[dep] = 0
				}
			}
			inDegree[dep]++
		}
	}

	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	sort.Strings(queue)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if spec, exists := specMap[current]; exists {
			result = append(result, spec)
		}

		for _, dep := range graph[current] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
				sort.Strings(queue)
			}
		}
	}

	if len(result) != len(graph) {
		return nil, fmt.Errorf("cyclic dependency detected in features")
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

func GenerateImageTag(projectName string, features []types.FeatureSpec) string {
	h := sha256.New()

	for _, f := range features {
		h.Write([]byte(f.Name))
		h.Write([]byte(f.Version))
		for k, v := range f.Params {
			h.Write([]byte(k))
			h.Write([]byte(fmt.Sprintf("%v", v)))
		}
	}

	hash := hex.EncodeToString(h.Sum(nil))[:12]
	return fmt.Sprintf("devc/%s:%s", sanitizeImageName(projectName), hash)
}

var invalidImageNameChars = regexp.MustCompile(`[^a-z0-9._-]+`)

func sanitizeImageName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = invalidImageNameChars.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "project"
	}
	return value
}
