package legacy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "pkg", "cli", "testdata", "migration", filepath.FromSlash(path)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestReleasedOutputIsRecognizedAsUnmodified(t *testing.T) {
	for _, path := range []string{"v1/AGENTS.md", "v2/AGENTS.md", "v2/docs/agents/RLM.md", "v3-precontract/CLAUDE.md", "v3-precontract/docs/references/workflows/implementation-delivery.md"} {
		rel := strings.SplitN(path, "/", 2)[1]
		if doc := Classify(rel, fixture(t, path)); !doc.Unmodified() {
			t.Errorf("%s: modified=%v project=%d", path, doc.ModifiedKeys(), len(doc.Project))
		}
	}
}

func TestClassifySeparatesEditedKitAndProjectSections(t *testing.T) {
	content := fixture(t, "v2/AGENTS.md")
	content = strings.Replace(content, "## Purpose\n", "## Purpose\n\n- local edit\n", 1)
	content += "\n## Team Notes\n\n- ours\n"
	doc := Classify("AGENTS.md", content)
	if got := doc.ModifiedKeys(); len(got) != 1 || got[0] != "PURPOSE" {
		t.Fatalf("modified = %v", got)
	}
	if len(doc.Project) != 1 || doc.Project[0].Key != "TEAM NOTES" {
		t.Fatalf("project = %#v", doc.Project)
	}
	if doc.Unmodified() {
		t.Fatal("edited document reported unmodified")
	}
}

func TestSectionsNormalizeLineEndingsAndTrailingSpace(t *testing.T) {
	content := fixture(t, "v2/docs/agents/RLM.md")
	noisy := strings.ReplaceAll(content, "\n", "  \r\n")
	if !Classify("docs/agents/RLM.md", noisy).Unmodified() {
		t.Fatal("whitespace-only differences changed ownership")
	}
}

func TestRuleKnownIgnoresStatusOnly(t *testing.T) {
	content := fixture(t, "v3-precontract/docs/references/rules/work-lane-gating.md")
	if !RuleKnown("work-lane-gating", content) {
		t.Fatal("shipped rule not recognized")
	}
	if !RuleKnown("work-lane-gating", strings.Replace(content, "status: active", "status: optional", 1)) {
		t.Fatal("status change affected recognition")
	}
	if RuleKnown("work-lane-gating", content+"\n- edit\n") || RuleKnown("delivery", content) {
		t.Fatal("edited or mismatched rule recognized")
	}
}

func TestProgressSummaryGeneratedStructure(t *testing.T) {
	generated := "# PROJECT PROGRESS SUMMARY\n\n## FEATURE PROGRESS TABLE\n\n| x |\n\n## PROJECT INTENT\n\ni\n\n## GLOBAL CONSTRAINTS\n\nc\n\n## FEATURE SUMMARIES\n\n### a\n\n- s\n\n## LAST UPDATED\n\nnow\n"
	if !ProgressSummaryGenerated(generated) {
		t.Fatal("rollup output not recognized")
	}
	for _, edited := range []string{generated + "\n## MAINTENANCE RULE\n\nx\n", strings.Replace(generated, "# PROJECT PROGRESS SUMMARY", "# Project Progress Summary", 1)} {
		if ProgressSummaryGenerated(edited) {
			t.Fatalf("hand-maintained summary recognized as generated:\n%s", edited)
		}
	}
}

func TestFingerprintDataCoversEveryRetiredGeneratedPath(t *testing.T) {
	paths := strings.Join(GeneratedPaths(), "\n")
	for _, want := range []string{"docs/agents/README.md", "docs/references/workflows/pull-request-merge.md", "docs/references/worktrees.md", "docs/references/tooling.md"} {
		if !strings.Contains(paths, want) {
			t.Errorf("fingerprints do not cover %s", want)
		}
	}
}

func TestSectionsIgnoreHeadingsInsideFences(t *testing.T) {
	content := "# T\n\n## Real\n\n```md\n## Not a section\n```\n\n~~~\n## Also not\n~~~\n"
	sections := Sections(content)
	if len(sections) != 2 || sections[1].Key != "REAL" || !strings.Contains(sections[1].Raw, "Also not") {
		t.Fatalf("sections = %#v", sections)
	}
}
