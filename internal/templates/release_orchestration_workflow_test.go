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
		"resolved exact current ready frontier",
		"scope-preserving repairs as ordinary, non-history-rewriting updates",
		"do not rebase, force-push, retarget, or otherwise\n   replace reviewed history",
		"Every changed existing head received fresh checks, review, and revalidation",
		"including heads changed by a human or\n  external system",
		"retained standing merge authority\n  without renewed merge authorization",
		"Every agent-performed in-place repair had explicit blocker-repair permission",
		"otherwise source, commit, and push stopped for renewed repair authority",
		"heads that cannot be updated safely, or explicit repository-policy or user",
		"Merge success was never substituted",
	} {
		if !strings.Contains(workflow, check) {
			t.Errorf("release-orchestration workflow missing %q", check)
		}
	}
	for _, forbidden := range []string{
		"changed or refreshed heads retain authority only when",
		"otherwise it received renewed authorization",
	} {
		if strings.Contains(workflow, forbidden) {
			t.Errorf("release-orchestration workflow conflates repair and merge authority with %q", forbidden)
		}
	}
}

func TestToolingTemplateRoutesPROrchestration(t *testing.T) {
	for _, check := range []string{
		"## PR Release Orchestration",
		"`kit pr orchestrate`",
		"`release-orchestration`",
		"does not enumerate PRs, merge, deploy, mutate infrastructure, or launch an agent",
		"Standing merge authority exists only when a human explicitly authorizes a bounded task, goal, or program",
		"may bind later-created in-scope PRs and refreshed heads",
	} {
		if !strings.Contains(agentsTooling, check) {
			t.Errorf("tooling template missing %q", check)
		}
	}
}
