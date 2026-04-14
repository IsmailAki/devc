package cmd

import (
	"runtime/debug"
	"testing"
)

func TestResolveBuildMetadataPrefersInjectedValues(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "v0.1.2"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abc123"},
			{Key: "vcs.time", Value: "2026-04-14T00:00:00Z"},
		},
	}

	metadata := resolveBuildMetadata("0.2.0", "deadbeef", "2026-04-13T00:00:00Z", info)

	if metadata.version != "0.2.0" {
		t.Fatalf("version = %q, want injected value", metadata.version)
	}
	if metadata.commit != "deadbeef" {
		t.Fatalf("commit = %q, want injected value", metadata.commit)
	}
	if metadata.date != "2026-04-13T00:00:00Z" {
		t.Fatalf("date = %q, want injected value", metadata.date)
	}
}

func TestResolveBuildMetadataFallsBackToBuildInfo(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "v0.1.3-0.20260414073448-3afa53696cd2"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "3afa53696cd26c4abd2dff82d9e1bb509f475667"},
			{Key: "vcs.time", Value: "2026-04-14T07:34:48Z"},
		},
	}

	metadata := resolveBuildMetadata(defaultAppVersion, defaultCommit, defaultDate, info)

	if metadata.version != "0.1.3-0.20260414073448-3afa53696cd2" {
		t.Fatalf("version = %q", metadata.version)
	}
	if metadata.commit != "3afa53696cd26c4abd2dff82d9e1bb509f475667" {
		t.Fatalf("commit = %q", metadata.commit)
	}
	if metadata.date != "2026-04-14T07:34:48Z" {
		t.Fatalf("date = %q", metadata.date)
	}
}

func TestResolveBuildMetadataDerivesCommitFromPseudoVersion(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "v0.1.3-0.20260414073448-3afa53696cd2"},
	}

	metadata := resolveBuildMetadata(defaultAppVersion, defaultCommit, defaultDate, info)

	if metadata.commit != "3afa53696cd2" {
		t.Fatalf("commit = %q, want pseudo-version suffix", metadata.commit)
	}
	if metadata.version != "0.1.3-0.20260414073448-3afa53696cd2" {
		t.Fatalf("version = %q", metadata.version)
	}
	if metadata.date != defaultDate {
		t.Fatalf("date = %q, want %q", metadata.date, defaultDate)
	}
}
