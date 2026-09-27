package templates

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestCheckedInContextWorkflowsMatchEmbeddedArtifacts(t *testing.T) {
	artifacts, err := ContextWorkflowArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	if len(artifacts) != 7 {
		t.Fatalf("workflow count = %d, want 7", len(artifacts))
	}
	for _, artifact := range artifacts {
		path := filepath.Join("..", "..", filepath.FromSlash(artifact.Path))
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read checked-in %s: %v", artifact.Path, err)
		}
		if string(content) != artifact.Content {
			t.Fatalf("checked-in workflow %s differs from embedded artifact", artifact.Path)
		}
	}
}

func TestPullRequestMergeWorkflowPreservesInPlaceRemediationBoundary(t *testing.T) {
	artifacts, err := ContextWorkflowArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	for _, artifact := range artifacts {
		if artifact.Slug != "pull-request-merge" {
			continue
		}
		for _, want := range []string{
			"standing-authority-permitted\n  actions to explicitly include blocker repair",
			"SHA/head OID as the evidence key only",
			"Do not request exact-head reauthorization after checks pass",
			"standing-\n   authority-permitted actions explicitly include blocker repair",
			"stop before source, commit, or push and obtain renewed repair authority",
			"Any changed or refreshed head,\n   including a human or external update, returns to `UNKNOWN`",
			"returns to `MERGE_READY`, merge under standing merge\n   authority without renewed merge authorization",
			"Reserve replacement pull\n   requests for material scope changes, heads that cannot be updated safely",
			"explicit repository-policy or user requirements",
			"Do not rebase, force-push,\n   retarget, or otherwise replace the branch's reviewed history",
			"Continue already-authorized standard deployment and browser verification",
			"do not stop for SHA-specific permission",
			"no changed head reuses readiness, review, or checks",
		} {
			if !strings.Contains(artifact.Content, want) {
				t.Fatalf("pull-request-merge workflow missing %q", want)
			}
		}
		for _, forbidden := range []string{
			"changed or refreshed heads retain authority only when",
			"otherwise require renewed authorization before merging it",
		} {
			if strings.Contains(artifact.Content, forbidden) {
				t.Fatalf("pull-request-merge workflow conflates repair and merge authority with %q", forbidden)
			}
		}
		return
	}
	t.Fatal("embedded pull-request-merge workflow not found")
}

func TestPRFeedbackWorkflowKeepsDelegationSafe(t *testing.T) {
	artifacts, err := ContextWorkflowArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	for _, artifact := range artifacts {
		if artifact.Slug != "pr-feedback-repair" {
			continue
		}
		for _, want := range []string{
			"Parallelize independent investigation",
			"serialize writes to shared files",
			"keep Git and GitHub mutations in the primary agent",
			"fresh independent read-only verifier",
			"supervisor self-review",
		} {
			if !strings.Contains(artifact.Content, want) {
				t.Fatalf("PR-feedback workflow missing %q", want)
			}
		}
		for _, retired := range []string{"at most three", "never more than four"} {
			if strings.Contains(artifact.Content, retired) {
				t.Fatalf("PR-feedback workflow retained fixed-cap policy %q", retired)
			}
		}
		return
	}
	t.Fatal("embedded pr-feedback-repair workflow not found")
}

// Workflow contracts require only the rules that define their own domain;
// everything else is contextual so unrelated tasks do not load it.
func TestContextWorkflowsRequireOnlyDomainRules(t *testing.T) {
	domain := map[string][]string{
		"implementation-delivery":               nil,
		"pr-feedback-repair":                    {"github-pr-delivery", "work-lane-gating"},
		"pull-request-merge":                    {"github-pr-merge"},
		"release-orchestration":                 {"github-pr-merge"},
		"repository-bootstrap":                  {"constitution-curation"},
		"repository-maintenance":                {"constitution-curation"},
		"cross-repository-program-coordination": {"cross-repository-program-coordination"},
	}
	artifacts, err := ContextWorkflowArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	for _, artifact := range artifacts {
		allowed, ok := domain[artifact.Slug]
		if !ok {
			t.Errorf("unexpected workflow %s", artifact.Slug)
			continue
		}
		required := regexp.MustCompile(`(?m)^  - slug: (\S+)\n    required: true`).FindAllStringSubmatch(artifact.Content, -1)
		if len(required) != len(allowed) {
			t.Errorf("workflow %s requires %d rules, want %d domain rules", artifact.Slug, len(required), len(allowed))
		}
		for _, match := range required {
			if !slices.Contains(allowed, match[1]) {
				t.Errorf("workflow %s requires non-domain rule %s", artifact.Slug, match[1])
			}
		}
		for _, retired := range []string{"docs/agents/GUARDRAILS.md", "docs/agents/RLM.md", "docs/agents/TOOLING.md", "docs/agents/WORKFLOWS.md", "coding-agent-context-usage"} {
			if strings.Contains(artifact.Content, retired) {
				t.Errorf("workflow %s still references %s", artifact.Slug, retired)
			}
		}
	}
}
