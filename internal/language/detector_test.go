package language

import (
	"testing"
)

func TestDetectDominant(t *testing.T) {
	tests := []struct {
		name      string
		languages map[string]int64
		want      string
	}{
		{
			name: "python dominant",
			languages: map[string]int64{
				"Python":     150000,
				"JavaScript": 50000,
				"Shell":      1000,
			},
			want: "python",
		},
		{
			name: "javascript dominant",
			languages: map[string]int64{
				"JavaScript": 200000,
				"TypeScript": 100000,
				"CSS":        10000,
			},
			want: "javascript",
		},
		{
			name: "c dominant",
			languages: map[string]int64{
				"C":        10000000,
				"Python":   500000,
				"Assembly": 100000,
				"Shell":    10000,
			},
			want: "c",
		},
		{
			name: "go dominant",
			languages: map[string]int64{
				"Go":       500000,
				"Shell":    5000,
				"Makefile": 2000,
			},
			want: "go",
		},
		{
			name:      "empty languages",
			languages: map[string]int64{},
			want:      "unknown",
		},
		{
			name: "single language",
			languages: map[string]int64{
				"Rust": 100000,
			},
			want: "rust",
		},
		{
			name: "jupyter notebook maps to python",
			languages: map[string]int64{
				"Jupyter Notebook": 200000,
				"Python":           50000,
			},
			want: "python",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectDominant(tt.languages)
			if got != tt.want {
				t.Errorf("DetectDominant() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Python", "python"},
		{"JavaScript", "javascript"},
		{"TypeScript", "typescript"},
		{"C++", "cpp"},
		{"C#", "csharp"},
		{"Jupyter Notebook", "python"},
		{"Shell", "bash"},
		{"Go", "go"},
		{"Rust", "rust"},
		{"Java", "java"},
		{"Ruby", "ruby"},
		{"PHP", "php"},
		{"Kotlin", "kotlin"},
		{"Swift", "swift"},
		{"Unknown", "Unknown"},
		{"SomeNewLang", "SomeNewLang"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeLanguage(tt.input)
			if got != tt.want {
				t.Errorf("normalizeLanguage(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
