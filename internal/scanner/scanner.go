package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SearchItem struct {
	Repo     string
	Path     string
	Fragment string
}

type Result struct {
	Key        string
	Repo       string
	File       string
	CommitDate string
	Value      string
}

type Client struct {
	httpClient *http.Client
	token      string
}

func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		token:      token,
	}
}

func (c *Client) Search(ctx context.Context, query string, limit int) ([]SearchItem, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}

	apiURL := fmt.Sprintf(
		"https://api.github.com/search/code?q=%s&per_page=%d",
		url.QueryEscape(query), limit,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3.text-match+json")
	req.Header.Set("User-Agent", "key-scanner")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 || resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limited (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("invalid GitHub token")
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("search failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []struct {
			Path       string `json:"path"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
			TextMatches []struct {
				Fragment string `json:"fragment"`
			} `json:"text_matches"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var items []SearchItem
	for _, item := range result.Items {
		fragment := ""
		if len(item.TextMatches) > 0 {
			fragment = item.TextMatches[0].Fragment
		}
		items = append(items, SearchItem{
			Repo:     item.Repository.FullName,
			Path:     item.Path,
			Fragment: fragment,
		})
	}

	return items, nil
}

func (c *Client) GetCommitDate(ctx context.Context, repo, path string) (string, error) {
	apiURL := fmt.Sprintf(
		"https://api.github.com/repos/%s/commits?path=%s&per_page=1",
		repo, url.QueryEscape(path),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", "key-scanner")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("commits request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "unknown", nil
	}

	var commits []struct {
		Commit struct {
			Committer struct {
				Date string `json:"date"`
			} `json:"committer"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return "unknown", nil
	}

	if len(commits) == 0 {
		return "unknown", nil
	}

	date := commits[0].Commit.Committer.Date
	if len(date) >= 10 {
		return date[:10], nil
	}
	if date == "" {
		return "unknown", nil
	}
	return date, nil
}

func ExtractValue(fragment, envVar string) string {
	fragment = strings.NewReplacer("<em>", "", "</em>", "").Replace(fragment)

	patterns := []string{
		envVar + "=",
		envVar + " =",
		envVar + " :",
		envVar + ":",
	}

	for _, prefix := range patterns {
		idx := strings.Index(fragment, prefix)
		if idx == -1 {
			continue
		}
		value := fragment[idx+len(prefix):]
		value = strings.TrimLeft(value, " ")
		value = strings.TrimLeft(value, ":")
		value = strings.TrimLeft(value, " ")

		if newlineIdx := strings.IndexAny(value, "\r\n"); newlineIdx != -1 {
			value = value[:newlineIdx]
		}
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if value != "" && !IsPlaceholder(value) {
			return value
		}
	}

	return ""
}

var placeholderPatterns = []string{
	"your_api_key",
	"your-api-key",
	"your_key",
	"your-key",
	"your_token",
	"your-token",
	"your_secret",
	"your-secret",
	"api_key_here",
	"api-key-here",
	"key_here",
	"key-here",
	"token_here",
	"token-here",
	"placeholder",
	"changeme",
	"change_me",
	"changethis",
	"change_this",
	"<your_",
	"<api_",
}

func IsPlaceholder(value string) bool {
	lower := strings.ToLower(value)
	for _, p := range placeholderPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	if value == "" {
		return true
	}
	return false
}
