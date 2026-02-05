// Package monitor provides functionality for monitoring AWS SDK releases.
package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Release represents a GitHub release.
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
}

// SDKReleaseMonitor monitors AWS SDK Go V2 releases on GitHub.
type SDKReleaseMonitor struct {
	githubToken string
	httpClient  *http.Client
	repoOwner   string
	repoName    string
}

// NewSDKReleaseMonitor creates a new SDK release monitor.
func NewSDKReleaseMonitor(githubToken string) *SDKReleaseMonitor {
	return &SDKReleaseMonitor{
		githubToken: githubToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		repoOwner: "aws",
		repoName:  "aws-sdk-go-v2",
	}
}

// GetLatestRelease fetches the latest release from GitHub.
func (m *SDKReleaseMonitor) GetLatestRelease(ctx context.Context) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", m.repoOwner, m.repoName)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if m.githubToken != "" {
		req.Header.Set("Authorization", "token "+m.githubToken)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release: %w", err)
	}

	return &release, nil
}

// GetReleases fetches recent releases from GitHub.
func (m *SDKReleaseMonitor) GetReleases(ctx context.Context, count int) ([]Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=%d", m.repoOwner, m.repoName, count)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if m.githubToken != "" {
		req.Header.Set("Authorization", "token "+m.githubToken)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode releases: %w", err)
	}

	return releases, nil
}

// GetCurrentVersion reads the current SDK version from a file.
func (m *SDKReleaseMonitor) GetCurrentVersion(versionFile string) (string, error) {
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// SaveVersion saves the SDK version to a file.
func (m *SDKReleaseMonitor) SaveVersion(versionFile, version string) error {
	return os.WriteFile(versionFile, []byte(version+"\n"), 0o644)
}

// HasNewVersion checks if there's a new SDK version available.
func (m *SDKReleaseMonitor) HasNewVersion(ctx context.Context, versionFile string) (bool, *Release, error) {
	release, err := m.GetLatestRelease(ctx)
	if err != nil {
		return false, nil, err
	}

	currentVersion, err := m.GetCurrentVersion(versionFile)
	if err != nil {
		if os.IsNotExist(err) {
			// No version file exists, treat as new version
			return true, release, nil
		}
		return false, nil, err
	}

	return release.TagName != currentVersion, release, nil
}

// GetReleaseByTag fetches a specific release by tag name.
func (m *SDKReleaseMonitor) GetReleaseByTag(ctx context.Context, tag string) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", m.repoOwner, m.repoName, tag)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if m.githubToken != "" {
		req.Header.Set("Authorization", "token "+m.githubToken)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release %s not found", tag)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release: %w", err)
	}

	return &release, nil
}

// CompareReleases returns the releases between two versions (exclusive of old, inclusive of new).
func (m *SDKReleaseMonitor) CompareReleases(ctx context.Context, oldVersion, newVersion string) ([]Release, error) {
	releases, err := m.GetReleases(ctx, 100)
	if err != nil {
		return nil, err
	}

	var between []Release
	inRange := false

	for _, r := range releases {
		if r.TagName == newVersion {
			inRange = true
		}
		if inRange {
			between = append(between, r)
		}
		if r.TagName == oldVersion {
			// Don't include the old version
			if len(between) > 0 {
				between = between[:len(between)-1]
			}
			break
		}
	}

	return between, nil
}
