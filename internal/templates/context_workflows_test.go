package templates

import (
	"os"
	"path/filepath"
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
			"standing-\n   authority-permitted actions explicitly include blocker repair",
			"stop before source, commit, or push and obtain renewed repair authority",
			"Any changed or refreshed head,\n   including a human or external update, returns to `UNKNOWN`",
			"returns to `MERGE_READY`, merge under standing merge\n   authority without renewed merge authorization",
			"Reserve replacement pull\n   requests for material scope changes, heads that cannot be updated safely",
			"explicit repository-policy or user requirements",
			"Do not rebase, force-push,\n   retarget, or otherwise replace the branch's reviewed history",
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

func TestPRFeedbackWorkflowUsesCapabilityAwareOrchestration(t *testing.T) {
	artifacts, err := ContextWorkflowArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	for _, artifact := range artifacts {
		if artifact.Slug != "pr-feedback-repair" {
			continue
		}
		for _, want := range []string{
			"Negotiate host-confirmed agent controls",
			"actual agents from logical and omitted lanes",
			"requested from effective profiles",
			"confirmed from unconfirmed parallelism",
			"replacement rebriefs",
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
