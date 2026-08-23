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
			"use ordinary,\n   non-history-rewriting commits to update the existing pull-request head",
			"return it to `UNKNOWN` pending fresh checks, review, revalidation, and\n   exact-head authorization",
			"Reserve replacement pull\n   requests for material scope changes, heads that cannot be updated safely",
			"explicit repository-policy or user requirements",
			"Do not rebase,\n   force-push, retarget, or otherwise replace the branch's reviewed history",
			"no changed head reuses readiness, review, checks, or merge authority",
		} {
			if !strings.Contains(artifact.Content, want) {
				t.Fatalf("pull-request-merge workflow missing %q", want)
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
