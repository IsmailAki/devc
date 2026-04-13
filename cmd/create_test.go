package cmd

import (
	"testing"

	"github.com/IsmailAki/devc/internal/naming"
)

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "https github url",
			url:       "https://github.com/torvalds/linux",
			wantOwner: "torvalds",
			wantRepo:  "linux",
			wantErr:   false,
		},
		{
			name:      "https github url with trailing slash",
			url:       "https://github.com/facebook/react/",
			wantOwner: "facebook",
			wantRepo:  "react",
			wantErr:   false,
		},
		{
			name:      "ssh github url",
			url:       "git@github.com:torvalds/linux.git",
			wantOwner: "torvalds",
			wantRepo:  "linux",
			wantErr:   false,
		},
		{
			name:      "github.com shorthand",
			url:       "github.com/owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "owner/repo shorthand",
			url:       "torvalds/linux",
			wantOwner: "torvalds",
			wantRepo:  "linux",
			wantErr:   false,
		},
		{
			name:      "https with .git extension",
			url:       "https://github.com/owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:    "invalid url",
			url:     "not-a-valid-url",
			wantErr: true,
		},
		{
			name:    "empty string",
			url:     "",
			wantErr: true,
		},
		{
			name:    "partial github url is invalid",
			url:     "github.com/owner",
			wantErr: true,
		},
		{
			name:    "single word without slash",
			url:     "just-a-word",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseRepoURL(tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseRepoURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("parseRepoURL() owner = %q, want %q", owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("parseRepoURL() repo = %q, want %q", repo, tt.wantRepo)
				}
			}
		})
	}
}

func TestGenerateContainerNameUniqueness(t *testing.T) {
	names := make(map[string]bool)
	for i := 0; i < 100; i++ {
		name := naming.GenerateContainerName("test", "repo")
		if names[name] {
			t.Errorf("duplicate container name generated: %s", name)
		}
		names[name] = true
	}
}
