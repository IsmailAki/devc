package language

import "sort"

const SignificanceThreshold = 0.05

type LanguageInfo struct {
	Name       string
	Bytes      int64
	Percentage float64
}

func DetectAll(languages map[string]int64) []LanguageInfo {
	if len(languages) == 0 {
		return nil
	}

	var total int64
	for _, bytes := range languages {
		total += bytes
	}

	var result []LanguageInfo
	for name, bytes := range languages {
		normalized := normalizeLanguage(name)
		percentage := float64(bytes) / float64(total)
		result = append(result, LanguageInfo{
			Name:       normalized,
			Bytes:      bytes,
			Percentage: percentage,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Bytes > result[j].Bytes
	})

	return result
}

func FilterSignificant(languages []LanguageInfo) []string {
	var result []string
	for _, lang := range languages {
		if lang.Percentage >= SignificanceThreshold {
			result = append(result, lang.Name)
		}
	}
	return result
}

func DetectDominant(languages map[string]int64) string {
	if len(languages) == 0 {
		return "unknown"
	}

	type langBytes struct {
		name  string
		bytes int64
	}

	var sorted []langBytes
	for name, bytes := range languages {
		sorted = append(sorted, langBytes{name: name, bytes: bytes})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].bytes > sorted[j].bytes
	})

	return normalizeLanguage(sorted[0].name)
}

func normalizeLanguage(lang string) string {
	normalizations := map[string]string{
		"C++":              "cpp",
		"C#":               "csharp",
		"Jupyter Notebook": "python",
		"Shell":            "bash",
		"Vim script":       "vim",
		"Emacs Lisp":       "elisp",
		"TypeScript":       "typescript",
		"JavaScript":       "javascript",
		"Python":           "python",
		"Java":             "java",
		"Go":               "go",
		"Rust":             "rust",
		"Ruby":             "ruby",
		"PHP":              "php",
		"C":                "c",
		"Kotlin":           "kotlin",
		"Scala":            "scala",
		"Swift":            "swift",
		"Objective-C":      "objective-c",
		"Dart":             "dart",
		"Elixir":           "elixir",
		"Erlang":           "erlang",
		"Haskell":          "haskell",
		"Lua":              "lua",
		"Perl":             "perl",
		"R":                "r",
		"Zig":              "zig",
	}

	if normalized, ok := normalizations[lang]; ok {
		return normalized
	}

	return lang
}
