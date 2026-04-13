package naming

import (
	"strings"
	"testing"
)

func TestGenerateLocalContainerNameIsStable(t *testing.T) {
	first := GenerateLocalContainerName("My Project", "/tmp/example/project")
	second := GenerateLocalContainerName("My Project", "/tmp/example/project")

	if first != second {
		t.Fatalf("GenerateLocalContainerName() = %q, want stable result %q", second, first)
	}
}

func TestGenerateLocalContainerNameUsesPathHash(t *testing.T) {
	first := GenerateLocalContainerName("project", "/tmp/example/project-a")
	second := GenerateLocalContainerName("project", "/tmp/example/project-b")

	if first == second {
		t.Fatal("expected different paths to generate different names")
	}
}

func TestGenerateContainerNameSanitizesSegments(t *testing.T) {
	name := GenerateContainerName("My Org", "Repo_Name")
	if !strings.HasPrefix(name, Prefix) {
		t.Fatalf("GenerateContainerName() = %q, want %q prefix", name, Prefix)
	}
}
