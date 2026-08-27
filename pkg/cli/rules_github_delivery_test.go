package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGitHubPRDeliveryRulesetUsesAutonomousRecovery(t *testing.T) {
	ruleset := loadGitHubPRDeliveryRuleset(t)
	for _, check := range []string{
		"retry autonomously",
		"another supported authenticated path such as `gh`",
		"without requesting routine retry permission",
		"Verify that no duplicate issue or PR was created",
		"Project-Oriented Worktree Delivery",
		"`~/worktrees/<owner>/<repository>/<lane>`",
		"native `git worktree` commands as the portable authority",
		"rules and reconciled guidance must not depend on them",
		"git worktree add -b",
		"Writable review repair must use the pull request's same-repository head branch",
		"capture the exact version-control-eligible paths it",
		"owns and each path's pre-command state",
		"abort rather than overwrite or combine ambiguous staged",
		"content state, removed",
		"paths must remain absent",
		"the staged index must contain exactly the",
		"trigger `work-lane-gating` recovery",
		"do not automatically transfer, stage, commit, push, restore, or discard it",
		"Never transfer or stage `.env`, secrets, ignored files, or machine-local configuration",
		"symlink the clone's primary checkout repository-root `.env` and `.envrc` by default",
		"preserve a repository- or user-supplied `.envrc`",
		"restore them if removal fails",
		"application startup, databases, port allocation, Temporal state",
		"work-lane routing as an earlier hard gate",
		"default to a new",
		"Continue an existing lane only",
		"primary/root checkout read-only",
		"Post-Merge Primary Leftover Cleanup",
		"`git clean -fd`",
		"leftover disposal after an authorized merge",
		"Remain in the coding-agent session",
		"Do not treat pull-request creation as session completion",
		"Merge the worktree pull request only after merge is authorized",
		"names the exact authorized pull request set",
		"following `github-pr-merge`",
		"enumerate or dry-run all untracked files",
		"verify every candidate is command-owned",
		"`git clean -fd` with only those verified paths",
		"restore those exact paths in both the index and the worktree",
		"only after revalidating",
		"still match the captured command-owned snapshot",
		"if any path mismatches or is ambiguous, stop",
	} {
		if !strings.Contains(ruleset.Body, check) {
			t.Fatalf("expected github-pr-delivery ruleset to contain %q", check)
		}
	}
	assertScopedPostMergeCleanupOrder(t, ruleset.Body)
	if strings.Contains(ruleset.Body, "stop and do not mutate to retry") {
		t.Fatal("expected github-pr-delivery ruleset to omit blanket mutation retry prohibition")
	}
	for _, forbidden := range []string{"`--no-link-env`", "`git wt repair", "`git wt path", "GitWT"} {
		if strings.Contains(ruleset.Body, forbidden) {
			t.Fatalf("github-pr-delivery must not depend on wrapper-specific policy %q", forbidden)
		}
	}
}

func TestGitHubPRDeliveryRulesetPreservesAdditionalScopeLane(t *testing.T) {
	ruleset := loadGitHubPRDeliveryRuleset(t)
	for _, check := range []string{
		"Create or reuse a separate GitHub issue for the additional scope",
		"Keep the existing pull request head branch. Do not create a second branch or pull request",
		"Scope every new commit for the additional work to its own issue number",
		"append a separate `Closes #123` line",
	} {
		if !strings.Contains(ruleset.Body, check) {
			t.Fatalf("expected github-pr-delivery ruleset to contain %q", check)
		}
	}
}

func TestGitHubPRDeliveryRulesetSkipsCIForDocumentationOnlySquashMerges(t *testing.T) {
	ruleset := loadGitHubPRDeliveryRuleset(t)
	for _, check := range []string{
		"Source-changing pull requests are not eligible for Kit-only CI skips",
		"complete branch and pull request diff",
		"`kit reconcile` is a candidate signal, not proof of eligibility",
		"documentation-only follow-up on a mixed pull request is not eligible",
		"inspect branch protection, repository rulesets, and other required-check policy",
		"Append the literal suffix `[skip ci]` to every qualifying commit title",
		"Append the same literal `[skip ci]` suffix to the pull request title",
		"apply only to workflows triggered by `push` and `pull_request`",
		"They do not suppress `pull_request_target` or other events",
		"Before selecting `Confirm squash and merge`",
		"generated squash commit title and body",
		"confirm the pull request title and HEAD commit message contain `[skip ci]`",
		"gh run list --commit \"$SQUASH_SHA\"",
	} {
		if !strings.Contains(ruleset.Body, check) {
			t.Fatalf("expected github-pr-delivery ruleset to contain %q", check)
		}
	}
}

func loadGitHubPRDeliveryRuleset(t *testing.T) rulesetDocument {
	t.Helper()
	path := filepath.Join("..", "..", "docs", "references", "rules", "github-pr-delivery.md")
	ruleset, err := parseRulesetFile(path)
	if err != nil {
		t.Fatalf("parseRulesetFile() error = %v", err)
	}
	if issues := validateRulesetDocument(ruleset, "github-pr-delivery"); len(issues) > 0 {
		t.Fatalf("github-pr-delivery ruleset issues = %#v", issues)
	}
	return ruleset
}
