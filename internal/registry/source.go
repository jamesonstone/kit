package registry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Source interface {
	LoadCatalog(context.Context, SourceConfig) (Catalog, string, error)
	LoadArtifact(context.Context, SourceConfig, CatalogArtifact, string) (string, error)
}

type GitHubSource struct {
	Client  *http.Client
	APIBase string
	RawBase string
	Token   string
}

func NewGitHubSource() *GitHubSource {
	return &GitHubSource{
		Client:  http.DefaultClient,
		APIBase: "https://api.github.com",
		RawBase: "https://raw.githubusercontent.com",
		Token:   os.Getenv("GITHUB_TOKEN"),
	}
}

func (source *GitHubSource) LoadCatalog(ctx context.Context, cfg SourceConfig) (Catalog, string, error) {
	if strings.TrimSpace(cfg.Repo) == "" {
		return Catalog{}, "", fmt.Errorf("registry source repo is required")
	}
	ref := cfg.Branch
	if ref == "" {
		ref = "main"
	}
	revision, err := source.resolveRevision(ctx, cfg.Repo, ref)
	if err != nil {
		return Catalog{}, "", err
	}
	path := cfg.CatalogPath
	if path == "" {
		path = "registry/catalog.yaml"
	}
	content, err := source.get(ctx, rawURL(source.RawBase, cfg.Repo, revision, path))
	if err != nil {
		return Catalog{}, "", fmt.Errorf("fetch registry catalog: %w", err)
	}
	catalog, err := ParseCatalog(content)
	return catalog, revision, err
}

func (source *GitHubSource) LoadArtifact(ctx context.Context, cfg SourceConfig, artifact CatalogArtifact, revision string) (string, error) {
	content, err := source.get(ctx, rawURL(source.RawBase, cfg.Repo, revision, artifact.SourcePath))
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", ArtifactKey(artifact.Kind, artifact.Slug), err)
	}
	if HashContent(string(content)) != artifact.Digest {
		return "", fmt.Errorf("%s digest does not match the catalog", ArtifactKey(artifact.Kind, artifact.Slug))
	}
	return string(content), nil
}

func (source *GitHubSource) resolveRevision(ctx context.Context, repo, ref string) (string, error) {
	endpoint := strings.TrimRight(source.APIBase, "/") + "/repos/" + repo + "/commits/" + url.PathEscape(ref)
	content, err := source.get(ctx, endpoint)
	if err != nil {
		return "", fmt.Errorf("resolve registry revision: %w", err)
	}
	var response struct {
		SHA string `json:"sha"`
	}
	if err := json.Unmarshal(content, &response); err != nil || response.SHA == "" {
		return "", fmt.Errorf("resolve registry revision: invalid GitHub response")
	}
	return response.SHA, nil
}

func (source *GitHubSource) get(ctx context.Context, endpoint string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "kit-contract")
	if source.Token != "" {
		request.Header.Set("Authorization", "Bearer "+source.Token)
	}
	response, err := source.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("%s", response.Status)
	}
	return io.ReadAll(response.Body)
}

type LocalSource struct{ Root string }

func (source LocalSource) LoadCatalog(_ context.Context, cfg SourceConfig) (Catalog, string, error) {
	root := source.Root
	if root == "" && cfg.Path != "" {
		root = cfg.Path
	}
	path := cfg.CatalogPath
	if path == "" {
		path = "registry/catalog.yaml"
	}
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		return Catalog{}, "", err
	}
	catalog, err := ParseCatalog(content)
	if err != nil {
		return Catalog{}, "", err
	}
	sum := sha256.Sum256(content)
	return catalog, "local-" + hex.EncodeToString(sum[:]), nil
}

func (source LocalSource) LoadArtifact(_ context.Context, cfg SourceConfig, artifact CatalogArtifact, _ string) (string, error) {
	root := source.Root
	if root == "" && cfg.Path != "" {
		root = cfg.Path
	}
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(artifact.SourcePath)))
	if err != nil {
		return "", err
	}
	if HashContent(string(content)) != artifact.Digest {
		return "", fmt.Errorf("%s digest does not match the catalog", ArtifactKey(artifact.Kind, artifact.Slug))
	}
	return string(content), nil
}

func rawURL(base, repo, ref, path string) string {
	return strings.TrimRight(base, "/") + "/" + repo + "/" + ref + "/" + strings.TrimLeft(filepath.ToSlash(path), "/")
}
