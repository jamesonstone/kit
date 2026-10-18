package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

type mintActionRelease struct{ Version, SHA string }

var lookupMintActionRelease = latestMintActionRelease
var mintActionHTTPClient = &http.Client{
	Timeout: 20 * time.Second,
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return fmt.Errorf("mint release redirects are unsupported")
	},
}

const mintActionAPI = "https://api.github.com/repos/jamesonstone/mint"

var mintActionVersion = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
var mintActionSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)

// Resolve one published release to immutable source, then verify the Action
// inputs used by Kit. No credentials, local tools or repository state are used.
func latestMintActionRelease() (mintActionRelease, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var release struct {
		Tag        string `json:"tag_name"`
		Published  string `json:"published_at"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
	}
	if err := mintActionJSON(ctx, "/releases/latest", &release); err != nil {
		return mintActionRelease{}, err
	}
	_, publishedErr := time.Parse(time.RFC3339, release.Published)
	if release.Draft || release.Prerelease || publishedErr != nil || !mintActionVersion.MatchString(release.Tag) || !validVersion(release.Tag) || compareVersions(release.Tag, "v0.9.0") < 0 {
		return mintActionRelease{}, fmt.Errorf("latest official Mint release must be published, stable and at least v0.9.0")
	}
	type object struct {
		Type string `json:"type"`
		SHA  string `json:"sha"`
	}
	var ref struct {
		Ref    string `json:"ref"`
		Object object `json:"object"`
	}
	if err := mintActionJSON(ctx, "/git/ref/tags/"+release.Tag, &ref); err != nil {
		return mintActionRelease{}, err
	}
	if ref.Ref != "refs/tags/"+release.Tag {
		return mintActionRelease{}, fmt.Errorf("official Mint release tag reference mismatch")
	}
	resolved := ref.Object
	for depth := 0; resolved.Type == "tag" && depth < 4; depth++ {
		if !mintActionSHA.MatchString(resolved.SHA) {
			return mintActionRelease{}, fmt.Errorf("official Mint tag has invalid object identity")
		}
		var tag struct {
			SHA    string `json:"sha"`
			Object object `json:"object"`
		}
		if err := mintActionJSON(ctx, "/git/tags/"+resolved.SHA, &tag); err != nil {
			return mintActionRelease{}, err
		}
		if tag.SHA != resolved.SHA {
			return mintActionRelease{}, fmt.Errorf("official Mint annotated tag identity mismatch")
		}
		resolved = tag.Object
	}
	if resolved.Type != "commit" || !mintActionSHA.MatchString(resolved.SHA) {
		return mintActionRelease{}, fmt.Errorf("official Mint release does not resolve to a commit")
	}
	data, err := mintActionRead(ctx, "/contents/action.yml?ref="+resolved.SHA, "application/vnd.github.raw+json")
	if err != nil {
		return mintActionRelease{}, err
	}
	var action struct {
		Inputs map[string]any `yaml:"inputs"`
	}
	if err := yaml.Unmarshal(data, &action); err != nil {
		return mintActionRelease{}, fmt.Errorf("published Mint Action metadata is invalid")
	}
	for _, input := range []string{"command", "prev-tag", "current-tag", "current-ref", "owner", "repo", "output", "commitish", "github-token"} {
		if _, ok := action.Inputs[input]; !ok {
			return mintActionRelease{}, fmt.Errorf("published Mint Action is missing required input %s", input)
		}
	}
	return mintActionRelease{Version: release.Tag, SHA: resolved.SHA}, nil
}

func mintActionJSON(ctx context.Context, path string, target any) error {
	data, err := mintActionRead(ctx, path, "application/vnd.github+json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("official Mint release metadata is invalid")
	}
	return nil
}

func mintActionRead(ctx context.Context, path, accept string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mintActionAPI+path, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare official Mint release lookup")
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("User-Agent", "kit-mint-scaffold")
	response, err := mintActionHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("official Mint release lookup failed; check network access and retry")
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("official Mint release lookup returned HTTP %d", response.StatusCode)
	}
	const limit = 1024 * 1024
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil || len(data) > limit {
		return nil, fmt.Errorf("official Mint release response is unreadable or too large")
	}
	return data, nil
}
