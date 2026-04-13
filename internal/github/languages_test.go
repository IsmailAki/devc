package github

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFetchLanguagesMock(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		want       LanguagesResponse
		wantErr    bool
	}{
		{
			name:       "valid response",
			response:   `{"Python": 150000, "JavaScript": 50000, "Shell": 1000}`,
			statusCode: 200,
			want: LanguagesResponse{
				"Python":     150000,
				"JavaScript": 50000,
				"Shell":      1000,
			},
		},
		{
			name:       "empty response",
			response:   `{}`,
			statusCode: 200,
			want:       LanguagesResponse{},
		},
		{
			name:       "repository not found",
			response:   `{"message": "Not Found"}`,
			statusCode: 404,
			wantErr:    true,
		},
		{
			name:       "rate limited",
			response:   `{"message": "API rate limit exceeded"}`,
			statusCode: 403,
			wantErr:    true,
		},
		{
			name:       "server error",
			response:   `{"message": "Internal Server Error"}`,
			statusCode: 500,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					if req.URL.String() != "https://example.test/repos/owner/repo/languages" {
						t.Fatalf("unexpected URL: %s", req.URL.String())
					}

					return &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(bytes.NewBufferString(tt.response)),
						Header:     make(http.Header),
					}, nil
				}),
			}

			got, err := fetchLanguages(client, "https://example.test/repos/owner/repo/languages", "owner", "repo")
			if (err != nil) != tt.wantErr {
				t.Fatalf("fetchLanguages() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if len(got) != len(tt.want) {
				t.Fatalf("fetchLanguages() length = %d, want %d", len(got), len(tt.want))
			}

			for key, wantValue := range tt.want {
				if got[key] != wantValue {
					t.Fatalf("fetchLanguages()[%s] = %d, want %d", key, got[key], wantValue)
				}
			}
		})
	}
}

func TestFetchLanguagesNetworkError(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("boom")
		}),
	}

	_, err := fetchLanguages(client, "https://example.test/repos/owner/repo/languages", "owner", "repo")
	if err == nil {
		t.Fatal("fetchLanguages() expected error, got nil")
	}
}
