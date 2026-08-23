package templates

import (
	"strings"
	"testing"
)

func TestReleaseOrchestrationWorkflowPreservesAuthorityBoundaries(t *testing.T) {
	artifacts, err := ContextWorkflowArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	var workflow string
	for _, artifact := range artifacts {
		if artifact.Slug == "release-orchestration" {
			workflow = artifact.Content
			break
		}
	}
	if workflow == "" {
		t.Fatal("release-orchestration workflow is missing")
	}
	for _, check := range []string{
		"dependencies:\n  - implementation-delivery",
		"slug: github-pr-merge\n    required: true",
		"slug: infrastructure-change-approval\n    required: false",
		"slug: agent-team-orchestration\n    required: false",
		"slug: cross-repository-program-coordination\n    required: false",
		"Resolve `pull-request-merge`",
		"current authorized `MERGE_READY` frontier",
		"scope-preserving repairs as ordinary, non-history-rewriting updates",
		"do not rebase, force-push, retarget, or otherwise\n   replace reviewed history",
		"Every changed existing head received fresh checks, review, revalidation",
		"exact-head authorization before merge",
		"heads that cannot be updated safely, or explicit repository-policy or user",
		"Merge success was never substituted",
	} {
		if !strings.Contains(workflow, check) {
			t.Errorf("release-orchestration workflow missing %q", check)
		}
	}
}

func TestToolingTemplateRoutesPROrchestration(t *testing.T) {
	for _, check := range []string{
		"## PR Release Orchestration",
		"`kit pr orchestrate`",
		"`release-orchestration`",
		"does not enumerate PRs, merge, deploy, mutate infrastructure, or launch an agent",
		"PR delivery never implies merge consent. A direct user request or accepted bounded merge plan must name the exact authorized PR set.",
	} {
		if !strings.Contains(agentsTooling, check) {
			t.Errorf("tooling template missing %q", check)
		}
	}
}
