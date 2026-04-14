package cmd

import (
	"runtime/debug"
	"strings"
)

const (
	defaultAppVersion = "dev"
	defaultCommit     = "none"
	defaultDate       = "unknown"
)

type buildMetadata struct {
	version string
	commit  string
	date    string
}

func resolveBuildMetadata(version, commit, date string, info *debug.BuildInfo) buildMetadata {
	resolved := buildMetadata{
		version: version,
		commit:  commit,
		date:    date,
	}

	if info != nil {
		if needsVersionFallback(resolved.version) {
			if buildVersion := normalizeVersion(info.Main.Version); buildVersion != "" {
				resolved.version = buildVersion
			}
		}

		if needsCommitFallback(resolved.commit) {
			if revision := buildSetting(info, "vcs.revision"); revision != "" {
				resolved.commit = revision
			} else if pseudoCommit := commitFromVersion(info.Main.Version); pseudoCommit != "" {
				resolved.commit = pseudoCommit
			}
		}

		if needsDateFallback(resolved.date) {
			if buildDate := buildSetting(info, "vcs.time"); buildDate != "" {
				resolved.date = buildDate
			}
		}
	}

	if resolved.version == "" {
		resolved.version = defaultAppVersion
	}
	if resolved.commit == "" {
		resolved.commit = defaultCommit
	}
	if resolved.date == "" {
		resolved.date = defaultDate
	}

	return resolved
}

func loadBuildMetadata() buildMetadata {
	info, _ := debug.ReadBuildInfo()
	return resolveBuildMetadata(appVersion, commit, date, info)
}

func needsVersionFallback(version string) bool {
	return version == "" || version == defaultAppVersion || version == "(devel)"
}

func needsCommitFallback(commit string) bool {
	return commit == "" || commit == defaultCommit || commit == defaultDate
}

func needsDateFallback(date string) bool {
	return date == "" || date == defaultDate
}

func normalizeVersion(version string) string {
	if version == "" || version == "(devel)" {
		return ""
	}

	return strings.TrimPrefix(version, "v")
}

func buildSetting(info *debug.BuildInfo, key string) string {
	for _, setting := range info.Settings {
		if setting.Key == key {
			return setting.Value
		}
	}

	return ""
}

func commitFromVersion(version string) string {
	version = normalizeVersion(version)
	if version == "" {
		return ""
	}

	parts := strings.Split(version, "-")
	if len(parts) < 3 {
		return ""
	}

	candidate := parts[len(parts)-1]
	if len(candidate) != 12 {
		return ""
	}

	for _, r := range candidate {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return ""
		}
	}

	return candidate
}
