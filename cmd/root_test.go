package cmd

import (
	"testing"
)

func TestIsRepoURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://github.com/owner/repo", true},
		{"http://github.com/owner/repo", true},
		{"https://github.com/owner/repo.git", true},
		{"git@github.com:owner/repo.git", true},
		{"github.com/owner/repo", true},
		{"owner/repo", false},
		{"create", false},
		{"list", false},
		{"delete", false},
		{"start", false},
		{"stop", false},
		{"", false},
		{"not-a-url", false},
		{"https://gitlab.com/owner/repo", false},
		{"https://bitbucket.org/owner/repo", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isRepoURL(tt.input)
			if got != tt.want {
				t.Errorf("isRepoURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
