package update

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

// httpClient is used for API and download requests.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// noRedirectClient resolves the "latest release" web redirect without
// following it, which is the fallback when the API is rate-limited.
var noRedirectClient = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// Latest returns the tag name of the newest published release, e.g. "v0.2.0".
func Latest(ctx context.Context) (string, error) {
	tag, err := latestFromAPI(ctx)
	if err == nil {
		return tag, nil
	}
	tag, fallbackErr := latestFromRedirect(ctx)
	if fallbackErr != nil {
		return "", err
	}
	return tag, nil
}

func latestFromAPI(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+"/repos/"+Repo+"/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "gocraft")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return "", err
	}
	if payload.TagName == "" {
		return "", fmt.Errorf("GitHub API returned an empty tag")
	}
	return payload.TagName, nil
}

// latestFromRedirect follows releases/latest and reads the tag from the
// Location header, avoiding the unauthenticated API rate limit.
func latestFromRedirect(ctx context.Context) (string, error) {
	endpoint := downloadBase + "/" + Repo + "/releases/latest"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))

	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no release redirect from %s", endpoint)
	}
	u, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 || parts[len(parts)-2] != "tag" || parts[len(parts)-1] == "" {
		return "", fmt.Errorf("unexpected release redirect %q", location)
	}
	return parts[len(parts)-1], nil
}
