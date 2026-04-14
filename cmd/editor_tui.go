package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/IsmailAki/devc/internal/feature"
	"github.com/IsmailAki/devc/pkg/types"
)

type featureOption struct {
	Name  string
	Label string
}

func promptProjectName(defaultName string) (string, error) {
	name := strings.TrimSpace(defaultName)
	if err := survey.AskOne(&survey.Input{
		Message: "Project name:",
		Default: defaultName,
	}, &name, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	return strings.TrimSpace(name), nil
}

func pickContainer(containers []containerInfo, message string) (string, error) {
	if len(containers) == 0 {
		return "", fmt.Errorf("no containers available")
	}

	options := make([]string, 0, len(containers))
	labelToName := make(map[string]string, len(containers))
	for _, container := range containers {
		project := container.Metadata.Repository
		if project == "" {
			project = container.Metadata.Name
		}

		detail := container.Metadata.SourcePath
		if detail == "" && container.Metadata.Branch != "" {
			detail = container.Metadata.Branch
		}
		if detail == "" {
			detail = container.Metadata.ProjectPath
		}

		label := fmt.Sprintf("%s [%s | %s] %s", container.Name, container.State.Status, container.Metadata.InitMode, project)
		if detail != "" {
			label += " - " + detail
		}

		options = append(options, label)
		labelToName[label] = container.Name
	}

	var selected string
	if err := survey.AskOne(&survey.Select{
		Message:  message,
		Options:  options,
		PageSize: minInt(len(options), 12),
	}, &selected); err != nil {
		return "", err
	}

	return labelToName[selected], nil
}

func editProjectPlugins(cfg *types.ProjectConfig, registry *feature.Registry) error {
	options, defaults := buildFeatureOptions(cfg.Features, registry)
	if len(options) == 0 {
		cfg.Features = nil
		return nil
	}

	labels := make([]string, 0, len(options))
	labelToName := make(map[string]string, len(options))
	for _, option := range options {
		labels = append(labels, option.Label)
		labelToName[option.Label] = option.Name
	}

	var selectedLabels []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message:  "Select plugins:",
		Options:  labels,
		Default:  defaults,
		PageSize: minInt(len(labels), 12),
	}, &selectedLabels); err != nil {
		return err
	}

	selectedNames := make([]string, 0, len(selectedLabels))
	for _, label := range selectedLabels {
		selectedNames = append(selectedNames, labelToName[label])
	}

	cfg.Features = mergeSelectedFeatures(cfg.Features, selectedNames)
	for i := range cfg.Features {
		definition, ok := registry.Get(cfg.Features[i].Name)
		if !ok {
			continue
		}
		if err := promptFeatureSettings(&cfg.Features[i], *definition); err != nil {
			return err
		}
	}

	return nil
}

func promptRebuildNow(containerName string) (bool, error) {
	var rebuild bool
	if err := survey.AskOne(&survey.Confirm{
		Message: fmt.Sprintf("Rebuild '%s' now?", containerName),
		Default: true,
	}, &rebuild); err != nil {
		return false, err
	}

	return rebuild, nil
}

func promptConfirm(message string, defaultValue bool) (bool, error) {
	var confirmed bool
	if err := survey.AskOne(&survey.Confirm{
		Message: message,
		Default: defaultValue,
	}, &confirmed); err != nil {
		return false, err
	}

	return confirmed, nil
}

func isPromptCancelled(err error) bool {
	return err == terminal.InterruptErr
}

func buildFeatureOptions(current []types.FeatureSpec, registry *feature.Registry) ([]featureOption, []string) {
	registryFeatures := registry.List()
	sort.Slice(registryFeatures, func(i, j int) bool {
		return registryFeatures[i].Name < registryFeatures[j].Name
	})

	options := make([]featureOption, 0, len(registryFeatures)+len(current))
	defaults := make([]string, 0, len(current))
	seen := make(map[string]struct{}, len(registryFeatures)+len(current))
	selected := make(map[string]struct{}, len(current))

	for _, spec := range current {
		selected[spec.Name] = struct{}{}
	}

	for _, item := range registryFeatures {
		label := item.Name
		if item.Description != "" {
			label += " - " + item.Description
		}
		option := featureOption{Name: item.Name, Label: label}
		options = append(options, option)
		seen[item.Name] = struct{}{}
		if _, ok := selected[item.Name]; ok {
			defaults = append(defaults, label)
		}
	}

	for _, spec := range current {
		if _, ok := seen[spec.Name]; ok {
			continue
		}
		label := spec.Name + " - unavailable (kept from current config)"
		options = append(options, featureOption{Name: spec.Name, Label: label})
		defaults = append(defaults, label)
	}

	return options, defaults
}

func mergeSelectedFeatures(current []types.FeatureSpec, selected []string) []types.FeatureSpec {
	if len(selected) == 0 {
		return nil
	}

	selectedSet := make(map[string]struct{}, len(selected))
	currentByName := make(map[string]types.FeatureSpec, len(current))
	for _, name := range selected {
		selectedSet[name] = struct{}{}
	}
	for _, spec := range current {
		currentByName[spec.Name] = cloneFeatureSpec(spec)
	}

	merged := make([]types.FeatureSpec, 0, len(selected))
	for _, spec := range current {
		if _, ok := selectedSet[spec.Name]; !ok {
			continue
		}
		merged = append(merged, currentByName[spec.Name])
		delete(selectedSet, spec.Name)
	}

	for _, name := range selected {
		if _, ok := selectedSet[name]; !ok {
			continue
		}
		merged = append(merged, types.FeatureSpec{Name: name})
		delete(selectedSet, name)
	}

	return merged
}

func promptFeatureSettings(spec *types.FeatureSpec, definition types.Feature) error {
	if version, ok := definition.Params["version"]; ok {
		updated, err := promptOptionalValue(
			fmt.Sprintf("%s version:", spec.Name),
			spec.Version,
			version.Default,
		)
		if err != nil {
			return err
		}
		spec.Version = updated
	}

	paramNames := make([]string, 0, len(definition.Params))
	for name := range definition.Params {
		if name == "version" {
			continue
		}
		paramNames = append(paramNames, name)
	}
	sort.Strings(paramNames)

	for _, name := range paramNames {
		param := definition.Params[name]
		current := ""
		if spec.Params != nil {
			if value, ok := spec.Params[name]; ok {
				current = fmt.Sprint(value)
			}
		}

		message := fmt.Sprintf("%s %s:", spec.Name, name)
		if param.Description != "" {
			message = fmt.Sprintf("%s (%s)", message, param.Description)
		}

		updated, err := promptOptionalValue(message, current, param.Default)
		if err != nil {
			return err
		}

		if updated == "" {
			if spec.Params != nil {
				delete(spec.Params, name)
			}
			continue
		}

		if spec.Params == nil {
			spec.Params = make(map[string]interface{})
		}
		spec.Params[name] = updated
	}

	if len(spec.Params) == 0 {
		spec.Params = nil
	}

	return nil
}

func promptOptionalValue(message, currentValue, defaultValue string) (string, error) {
	answer := currentValue
	if answer == "" {
		answer = defaultValue
	}

	if err := survey.AskOne(&survey.Input{
		Message: message,
		Default: answer,
	}, &answer); err != nil {
		return "", err
	}

	answer = strings.TrimSpace(answer)
	if currentValue != "" {
		return answer, nil
	}
	if answer == defaultValue {
		return "", nil
	}
	return answer, nil
}

func cloneProjectConfig(cfg *types.ProjectConfig) *types.ProjectConfig {
	if cfg == nil {
		return nil
	}

	clone := *cfg
	clone.Features = cloneFeatureSpecs(cfg.Features)
	if cfg.Env != nil {
		clone.Env = make(map[string]string, len(cfg.Env))
		for key, value := range cfg.Env {
			clone.Env[key] = value
		}
	}
	if cfg.Ports != nil {
		clone.Ports = append([]int(nil), cfg.Ports...)
	}
	if cfg.Git != nil {
		git := *cfg.Git
		clone.Git = &git
	}

	return &clone
}

func cloneFeatureSpecs(features []types.FeatureSpec) []types.FeatureSpec {
	cloned := make([]types.FeatureSpec, 0, len(features))
	for _, feature := range features {
		cloned = append(cloned, cloneFeatureSpec(feature))
	}
	return cloned
}

func cloneFeatureSpec(spec types.FeatureSpec) types.FeatureSpec {
	cloned := spec
	if spec.Params != nil {
		cloned.Params = make(map[string]interface{}, len(spec.Params))
		for key, value := range spec.Params {
			cloned.Params[key] = value
		}
	}
	return cloned
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
