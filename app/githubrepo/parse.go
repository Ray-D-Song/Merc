package githubrepo

import (
	"fmt"
	"net/url"
	"strings"
)

type Repository struct {
	Owner string
	Repo  string
	URL   string
}

func Parse(raw string) (Repository, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Repository{}, fmt.Errorf("repository URL is required")
	}

	if strings.HasPrefix(raw, "git@github.com:") {
		path := strings.TrimPrefix(raw, "git@github.com:")
		return fromPath(path)
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return Repository{}, fmt.Errorf("parse repository URL: %w", err)
	}
	if parsed.Host != "github.com" && parsed.Host != "www.github.com" {
		return Repository{}, fmt.Errorf("only github.com repositories are supported")
	}
	return fromPath(parsed.Path)
}

func fromPath(path string) (Repository, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return Repository{}, fmt.Errorf("invalid GitHub repository path")
	}
	owner := parts[0]
	repo := strings.TrimSuffix(parts[1], ".git")
	if repo == "" {
		return Repository{}, fmt.Errorf("invalid GitHub repository name")
	}
	return Repository{
		Owner: owner,
		Repo:  repo,
		URL:   fmt.Sprintf("https://github.com/%s/%s", owner, repo),
	}, nil
}
