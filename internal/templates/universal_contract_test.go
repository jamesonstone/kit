package templates

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

// criticalInvariants maps each universal invariant to one marker that must stay
// directly visible in the always-loaded contract. Markers pin meaning, not
// surrounding prose; reword freely as long as each invariant stays stated.
var criticalInvariants = map[string]string{
	"human-only authorship":     "Only the human user is author, committer, and assignee",
	"human assignee":            "Assign every created or reused issue and pull request to the human",
	"primary checkout":          "Treat the clone's primary checkout as read-only",
	"pull-request delivery":     "deliver it through a ready pull request",
	"no lane-choice question":   "Never ask the user to choose between lanes",
	"protected branches":        "Never commit directly to the default or a protected branch",
	"merge boundary":            "Merge is its own boundary",
	"merge readiness":           "Merge only a current `MERGE_READY` head",
	"no exact-head reauth":      "never request exact-head reauthorization",
	"infrastructure approval":   "get one explicit approval",
	"hard delete":               "A hard delete",
	"slack read-only":           "Slack is read-only by default",
	"draft is not send":         "A draft is not a send",
	"external actions":          "Ask before any other outward-facing or irreversible action",
	"secrets":                   "Never stage secrets",
	"aws identity":              "run `kit aws verify`",
	"human pause":               "A direct human pause, hold, or revocation stops the affected action",
	"source size":               "300 physical lines or less",
	"delegation safety":         "Delegated agents never mutate Git or GitHub",
	"truthful checks":           "Never present a pending, skipped, unavailable, or unrun check as passing",
	"recoverable change report": "needed to find or undo every change",
}

func TestUniversalContractCarriesCriticalInvariants(t *testing.T) {
	contract := UniversalContractBlock()
	for name, marker := range criticalInvariants {
		if !strings.Contains(contract, marker) {
			t.Errorf("universal contract lost invariant %s (marker %q)", name, marker)
		}
	}
}

func TestUniversalContractStaysSmall(t *testing.T) {
	if words := len(strings.Fields(universalContractBody)); words > 1500 {
		t.Fatalf("universal contract has %d words; move task-specific guidance to a contextual rule", words)
	}
}

func TestVendorAdaptersRenderOneContract(t *testing.T) {
	block := UniversalContractBlock()
	for path, title := range map[string]string{
		"AGENTS.md":                       "AGENTS",
		"CLAUDE.md":                       "CLAUDE",
		".github/copilot-instructions.md": "GitHub Copilot Repository Instructions",
	} {
		got := InstructionFileForVersion(path, config.InstructionScaffoldVersionMemory)
		if want := "# " + title + "\n\n" + block; got != want {
			t.Errorf("%s is not the title plus the shared contract block:\n%s", path, got)
		}
		if strings.Count(got, UniversalContractBeginMarker) != 1 || strings.Count(got, UniversalContractEndMarker) != 1 {
			t.Errorf("%s must carry exactly one managed contract block", path)
		}
	}
}

func TestContractContextualRulesExist(t *testing.T) {
	section := universalContractBody[strings.Index(universalContractBody, "## Contextual Rules"):]
	matches := regexp.MustCompile("`([a-z0-9/-]+\\.md)`").FindAllStringSubmatch(section, -1)
	if len(matches) < 10 {
		t.Fatalf("contextual rules section names only %d rule files", len(matches))
	}
	for _, match := range matches {
		name := match[1]
		path := filepath.Join("..", "..", "docs", "references", "rules", name)
		if strings.Contains(name, "/") {
			path = filepath.Join("..", "..", filepath.FromSlash(name))
		}
		if _, err := os.Stat(path); err != nil {
			t.Errorf("contextual rule %q named by the contract does not exist: %v", name, err)
		}
	}
}

func TestV3SupportDocsDoNotRestateContract(t *testing.T) {
	for _, file := range InstructionSupportFiles(config.InstructionScaffoldVersionMemory) {
		for name, marker := range criticalInvariants {
			if strings.Contains(file.Content, marker) {
				t.Errorf("%s restates contract invariant %s; derive it from the contract instead", file.RelativePath, name)
			}
		}
	}
}

func TestCheckedInAgentSurfacesMatchGenerator(t *testing.T) {
	block := UniversalContractBlock()
	for _, path := range []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"} {
		got, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		// Project guidance may live outside the block; only the block is generated.
		if strings.Count(string(got), UniversalContractBeginMarker) != 1 || !strings.Contains(string(got), block) {
			t.Errorf("checked-in %s managed block drifted from the universal contract; regenerate it", path)
		}
	}
	readme, err := os.ReadFile(filepath.Join("..", "..", "docs", "agents", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(readme) != memoryInstructionSupportContent("docs/agents/README.md") {
		t.Error("checked-in docs/agents/README.md drifted from the generator")
	}
}

func TestV3ConstitutionBaselineDoesNotRestateContract(t *testing.T) {
	baseline := ConstitutionBaselineSectionFor(config.InstructionScaffoldVersionMemory)
	for name, marker := range criticalInvariants {
		if strings.Contains(baseline, marker) {
			t.Errorf("v3 Constitution baseline restates contract invariant %s", name)
		}
	}
	if !strings.Contains(Constitution, baseline) {
		t.Fatal("fresh Constitution does not carry the v3 baseline")
	}
}
