package minttool

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

func (systemBackend) release(ctx context.Context, version string) (string, error) {
	if version != "" && !validVersion(version) {
		return "", errors.New("invalid Mint release tag")
	}
	endpoint := "https://api.github.com/repos/jamesonstone/mint/releases/latest"
	if version != "" {
		endpoint = "https://api.github.com/repos/jamesonstone/mint/releases/tags/" + version
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", errors.New("cannot prepare official Mint release lookup")
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("release redirects are unsupported") }}
	res, err := client.Do(req)
	if err != nil {
		return "", errors.New("official Mint release lookup failed")
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return "", errors.New("official Mint release is unavailable (GitHub release lookup did not return 200)")
	}
	return decodeRelease(io.LimitReader(res.Body, 1024*1024), version)
}

func decodeRelease(body io.Reader, requested string) (string, error) {
	var release struct {
		Tag        string `json:"tag_name"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
		Published  string `json:"published_at"`
	}
	if json.NewDecoder(body).Decode(&release) != nil || release.Draft || release.Prerelease || release.Published == "" || !validVersion(release.Tag) || (requested != "" && requested != release.Tag) {
		return "", errors.New("mint target is not a verified published stable release")
	}
	return release.Tag, nil
}
