package cli

import (
	"strings"
	"testing"
)

func assertCapabilitiesInitTarget(t *testing.T) {
	t.Helper()

	payload := mustCapabilityDetail(t, "init", "init")
	if payload.Command.Command != "init" {
		t.Fatalf("command = %q, want init", payload.Command.Command)
	}
	if !strings.Contains(payload.Command.NetworkUse.FlagDependent, "--refresh") {
		t.Fatalf("expected init network use to document refresh registry fetch, got %#v", payload.Command.NetworkUse)
	}
	if !strings.Contains(payload.Command.NetworkUse.FlagDependent, "gh repo visibility") {
		t.Fatalf("expected init network use to document README badge visibility lookup, got %#v", payload.Command.NetworkUse)
	}
	if !strings.Contains(payload.Command.FileWrites.FlagDependent, "--dry-run") {
		t.Fatalf("expected init file writes to document dry-run, got %#v", payload.Command.FileWrites)
	}
	if !strings.Contains(payload.Command.FileWrites.Summary, "auto-assign.yml") {
		t.Fatalf("expected init file writes to document auto-assign workflow, got %#v", payload.Command.FileWrites)
	}
	if !strings.Contains(payload.Command.FileWrites.Summary, "README.md managed status badges and final Maintainers section") {
		t.Fatalf("expected init file writes to document README badge management, got %#v", payload.Command.FileWrites)
	}
	if !strings.Contains(payload.Command.FileWrites.FlagDependent, "github.default_assignees") {
		t.Fatalf("expected init file writes to document auto-assignee config fallback, got %#v", payload.Command.FileWrites)
	}
	if !strings.Contains(payload.Command.FileWrites.FlagDependent, "github.repository") {
		t.Fatalf("expected init file writes to document README badge repository source, got %#v", payload.Command.FileWrites)
	}
	if !strings.Contains(payload.Command.FileWrites.FlagDependent, "private repositories skip public Shields") {
		t.Fatalf("expected init file writes to document private README badge behavior, got %#v", payload.Command.FileWrites)
	}
	if !strings.Contains(payload.Command.FileWrites.FlagDependent, "jamesonstone attribution") {
		t.Fatalf("expected init file writes to document README Maintainers attribution, got %#v", payload.Command.FileWrites)
	}
	if !strings.Contains(payload.Command.FileWrites.FlagDependent, "documentation review prompt") {
		t.Fatalf("expected init file writes to document force refresh prompt, got %#v", payload.Command.FileWrites)
	}

	refreshFlag := findDetailedFlag(payload.Command.DetailedFlagBehavior, "--refresh")
	if refreshFlag == nil || !strings.Contains(refreshFlag.Summary, "loop.agent.command") || !strings.Contains(refreshFlag.Summary, "auto-assignment workflow") || !strings.Contains(refreshFlag.Summary, "README.md managed badges and Maintainers section") {
		t.Fatalf("expected --refresh flag to document loop agent, README badge, Maintainers, and auto-assignment workflow backfill, got %#v", refreshFlag)
	}
	if findDetailedFlag(payload.Command.DetailedFlagBehavior, "--diff") == nil {
		t.Fatalf("expected init detailed flags to include --diff")
	}
	forceFlag := findDetailedFlag(payload.Command.DetailedFlagBehavior, "--force")
	if forceFlag == nil || !strings.Contains(forceFlag.Summary, "documentation review prompt") {
		t.Fatalf("expected --force flag to document documentation review prompt, got %#v", forceFlag)
	}
}

func assertCapabilitiesSpecTarget(t *testing.T) {
	t.Helper()

	payload := mustCapabilityDetail(t, "spec", "spec")
	if !strings.Contains(payload.Command.FileWrites.Summary, "setup readiness") {
		t.Fatalf("expected spec file writes to document setup gate, got %#v", payload.Command.FileWrites)
	}
	if !strings.Contains(strings.Join(payload.Command.Caveats, " "), "bypass creates only the minimal baseline") {
		t.Fatalf("expected spec caveats to document bypass behavior, got %#v", payload.Command.Caveats)
	}
}

func assertCapabilitiesLegacyVerifyTarget(t *testing.T) {
	t.Helper()

	payload := mustCapabilityDetail(t, "legacy verify", "legacy", "verify")
	if payload.Kind != "capability_detail" {
		t.Fatalf("kind = %q, want capability_detail", payload.Kind)
	}
	if payload.Command.Command != "legacy verify" {
		t.Fatalf("command = %q, want legacy verify", payload.Command.Command)
	}
	if len(payload.Command.WhenToUse) == 0 || len(payload.Command.WhenNotToUse) == 0 || len(payload.Command.Examples) == 0 {
		t.Fatalf("expected detailed guidance fields to be populated: %#v", payload.Command)
	}
	if findDetailedFlag(payload.Command.DetailedFlagBehavior, "--dry-run") == nil {
		t.Fatalf("expected verify detailed flags to include --dry-run")
	}
	if findDetailedFlag(payload.Command.DetailedFlagBehavior, "--allow-shell") == nil {
		t.Fatalf("expected verify detailed flags to include --allow-shell")
	}
}

func assertCapabilitiesCITarget(t *testing.T) {
	t.Helper()

	payload := mustCapabilityDetail(t, "ci", "ci")
	if payload.Command.Command != "ci" {
		t.Fatalf("command = %q, want ci", payload.Command.Command)
	}
	if payload.Command.NetworkUse.Summary == "none" {
		t.Fatalf("expected ci targeted detail to describe network use")
	}
	if !strings.Contains(payload.Command.NetworkUse.Summary, "git/gh") {
		t.Fatalf("expected ci targeted detail to describe git/gh subprocess use, got %#v", payload.Command.NetworkUse)
	}
	if !strings.Contains(payload.Command.NetworkUse.FlagDependent, "--copilot") {
		t.Fatalf("expected ci targeted detail to describe optional copilot behavior, got %#v", payload.Command.NetworkUse)
	}
	if !strings.Contains(payload.Command.FileWrites.Summary, ".kit.yaml") {
		t.Fatalf("expected ci targeted detail to describe .kit.yaml cache writes, got %#v", payload.Command.FileWrites)
	}
}
