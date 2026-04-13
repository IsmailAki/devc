package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LanguagesResponse map[string]int64

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func FetchLanguages(owner, repo string) (LanguagesResponse, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/languages", owner, repo)
	return fetchLanguages(httpClient, url, owner, repo)
}

func fetchLanguages(client *http.Client, url, owner, repo string) (LanguagesResponse, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch languages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("repository not found: %s/%s", owner, repo)
	}

	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("GitHub API rate limit exceeded")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var langs LanguagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&langs); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return langs, nil
}
