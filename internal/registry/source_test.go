package registry

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGitHubSourcePinsCatalogAndArtifactsToResolvedRevision(t *testing.T) {
	document := testRuleDocument("Purpose.", "Rule.")
	digest := HashContent(document)
	catalog := fmt.Sprintf(`schema_version: 1
artifacts:
  - kind: ruleset
    slug: example
    description: Example rules
    visibility: downstream
    source_path: rules/example.md
    target_path: docs/references/rules/example.md
    version: 1
    digest: %s
    read_policy: must
`, digest)
	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests = append(requests, request.URL.Path)
		switch request.URL.Path {
		case "/repos/owner/repo/commits/main":
			_, _ = fmt.Fprint(writer, `{"sha":"abc123"}`)
		case "/owner/repo/abc123/registry/catalog.yaml":
			_, _ = fmt.Fprint(writer, catalog)
		case "/owner/repo/abc123/rules/example.md":
			_, _ = fmt.Fprint(writer, document)
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()
	source := &GitHubSource{Client: server.Client(), APIBase: server.URL, RawBase: server.URL}
	cfg := SourceConfig{Repo: "owner/repo", Branch: "main", CatalogPath: "registry/catalog.yaml"}
	parsed, revision, err := source.LoadCatalog(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if revision != "abc123" || len(parsed.Artifacts) != 1 {
		t.Fatalf("revision = %q, artifacts = %d", revision, len(parsed.Artifacts))
	}
	content, err := source.LoadArtifact(context.Background(), cfg, parsed.Artifacts[0], revision)
	if err != nil {
		t.Fatal(err)
	}
	if content != document || !strings.Contains(strings.Join(requests, "\n"), "/abc123/rules/example.md") {
		t.Fatalf("artifact content or pinned request was incorrect: %v", requests)
	}
}

func TestLocalSourceRejectsDigestMismatch(t *testing.T) {
	root, cfg := writeTestRegistry(t, "Purpose.", "Rule.")
	catalog, revision, err := (LocalSource{Root: root}).LoadCatalog(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root+"/rules/example.md", testRuleDocument("Changed.", "Rule."))
	_, err = (LocalSource{Root: root}).LoadArtifact(context.Background(), cfg, catalog.Artifacts[0], revision)
	if err == nil || !strings.Contains(err.Error(), "digest") {
		t.Fatalf("digest error = %v", err)
	}
}
