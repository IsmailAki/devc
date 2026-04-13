package version

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func DetectVersions(repoPath string, languages []string) map[string]string {
	versions := make(map[string]string)

	for _, lang := range languages {
		version := detectVersion(repoPath, lang)
		if version != "" {
			versions[lang] = version
		}
	}

	return versions
}

func detectVersion(repoPath, lang string) string {
	switch lang {
	case "javascript", "typescript":
		return detectNodeVersion(repoPath)
	case "python":
		return detectPythonVersion(repoPath)
	case "go":
		return detectGoVersion(repoPath)
	case "rust":
		return detectRustVersion(repoPath)
	case "ruby":
		return detectRubyVersion(repoPath)
	case "java", "kotlin", "scala":
		return detectJavaVersion(repoPath)
	case "csharp":
		return detectDotnetVersion(repoPath)
	default:
		return ""
	}
}

func detectNodeVersion(repoPath string) string {
	if content, err := os.ReadFile(filepath.Join(repoPath, ".nvmrc")); err == nil {
		return strings.TrimSpace(string(content))
	}

	if content, err := os.ReadFile(filepath.Join(repoPath, "package.json")); err == nil {
		var pkg struct {
			Engines struct {
				Node string `json:"node"`
			} `json:"engines"`
		}
		if err := json.Unmarshal(content, &pkg); err == nil && pkg.Engines.Node != "" {
			return pkg.Engines.Node
		}
	}

	return ""
}

func detectPythonVersion(repoPath string) string {
	if content, err := os.ReadFile(filepath.Join(repoPath, ".python-version")); err == nil {
		return strings.TrimSpace(string(content))
	}

	if content, err := os.ReadFile(filepath.Join(repoPath, "pyproject.toml")); err == nil {
		re := regexp.MustCompile(`requires-python\s*=\s*["']([^"']+)["']`)
		matches := re.FindSubmatch(content)
		if len(matches) > 1 {
			return string(matches[1])
		}
	}

	return ""
}

func detectGoVersion(repoPath string) string {
	if content, err := os.ReadFile(filepath.Join(repoPath, "go.mod")); err == nil {
		re := regexp.MustCompile(`(?m)^go\s+(\d+(?:\.\d+)?)`)
		matches := re.FindSubmatch(content)
		if len(matches) > 1 {
			return string(matches[1])
		}
	}

	return ""
}

func detectRustVersion(repoPath string) string {
	if content, err := os.ReadFile(filepath.Join(repoPath, "rust-toolchain.toml")); err == nil {
		re := regexp.MustCompile(`channel\s*=\s*["']([^"']+)["']`)
		matches := re.FindSubmatch(content)
		if len(matches) > 1 {
			return string(matches[1])
		}
	}

	return ""
}

func detectRubyVersion(repoPath string) string {
	if content, err := os.ReadFile(filepath.Join(repoPath, ".ruby-version")); err == nil {
		return strings.TrimSpace(string(content))
	}

	if content, err := os.ReadFile(filepath.Join(repoPath, "Gemfile")); err == nil {
		re := regexp.MustCompile(`ruby\s+["']([^"']+)["']`)
		matches := re.FindSubmatch(content)
		if len(matches) > 1 {
			return string(matches[1])
		}
	}

	return ""
}

func detectJavaVersion(repoPath string) string {
	if content, err := os.ReadFile(filepath.Join(repoPath, "pom.xml")); err == nil {
		re := regexp.MustCompile(`<maven\.compiler\.(?:source|release)>(\d+)</maven\.compiler\.(?:source|release)>`)
		matches := re.FindSubmatch(content)
		if len(matches) > 1 {
			return string(matches[1])
		}
	}

	if content, err := os.ReadFile(filepath.Join(repoPath, "build.gradle")); err == nil {
		re := regexp.MustCompile(`sourceCompatibility\s*=\s*["']?(\d+)["']?`)
		matches := re.FindSubmatch(content)
		if len(matches) > 1 {
			return string(matches[1])
		}
	}

	return ""
}

func detectDotnetVersion(repoPath string) string {
	if content, err := os.ReadFile(filepath.Join(repoPath, "global.json")); err == nil {
		var global struct {
			Sdk struct {
				Version string `json:"version"`
			} `json:"sdk"`
		}
		if err := json.Unmarshal(content, &global); err == nil && global.Sdk.Version != "" {
			parts := strings.Split(global.Sdk.Version, ".")
			if len(parts) >= 2 {
				return fmt.Sprintf("%s.%s", parts[0], parts[1])
			}
			return global.Sdk.Version
		}
	}

	matches, _ := filepath.Glob(filepath.Join(repoPath, "*.csproj"))
	if len(matches) > 0 {
		if content, err := os.ReadFile(matches[0]); err == nil {
			re := regexp.MustCompile(`<TargetFramework>(?:net)?(\d+(?:\.\d+)?)`)
			m := re.FindSubmatch(content)
			if len(m) > 1 {
				return string(m[1])
			}
		}
	}

	return ""
}
