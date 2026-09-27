package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/templates"
)

func TestUpsertConstitutionBaselineIsIdempotentAndPreservesCustomConstraints(t *testing.T) {
	const customConstraint = "Keep custom constraints intact."
	input := `# CONSTITUTION

## PRINCIPLES

Correctness first.

## CONSTRAINTS

` + customConstraint + `

## NON-GOALS

No hidden behavior.
`

	once, changed := upsertConstitutionBaseline(input, templates.ConstitutionBaselineSection)
	if !changed {
		t.Fatal("first upsert changed = false, want true")
	}
	if !strings.Contains(once, customConstraint) {
		t.Fatalf("first upsert removed custom constraint:\n%s", once)
	}

	twice, changed := upsertConstitutionBaseline(once, templates.ConstitutionBaselineSection)
	if changed {
		t.Fatalf("second upsert changed = true, want idempotent no-op:\nonce:\n%s\ntwice:\n%s", once, twice)
	}
	if twice != once || !strings.Contains(twice, customConstraint) {
		t.Fatalf("second upsert changed custom content:\nonce:\n%s\ntwice:\n%s", once, twice)
	}
}

func TestRefreshKeepsFreshInitConstitutionBaseline(t *testing.T) {
	fresh := templates.Constitution
	if !strings.Contains(fresh, templates.ConstitutionBaselineSection) {
		t.Fatal("fresh Constitution does not embed the shared baseline section")
	}

	refreshed, changed := upsertConstitutionBaseline(fresh, templates.ConstitutionBaselineSection)
	if changed || refreshed != strings.TrimRight(fresh, "\n")+"\n" {
		t.Fatalf("refresh changed a freshly initialized Constitution:\n%s", refreshed)
	}

	// A v3 project still carrying the pre-contract baseline converges to the
	// pointer baseline instead of keeping restated universal rules.
	legacyBaseline := "### " + templates.ConstitutionBaselineHeading + "\n\n<!-- BEGIN KIT-MANAGED BASELINE RULES -->\n- Keep every version-control-eligible handwritten implementation/source and test file at 300 physical lines or less.\n<!-- END KIT-MANAGED BASELINE RULES -->"
	stale := strings.Replace(fresh, templates.ConstitutionBaselineSection, legacyBaseline, 1)
	restored, changed := upsertConstitutionBaseline(stale, templates.ConstitutionBaselineSection)
	if !changed || !strings.Contains(restored, templates.ConstitutionBaselineSection) || strings.Contains(restored, "300 physical lines") {
		t.Fatalf("refresh did not converge to the v3 baseline:\n%s", restored)
	}
}

func TestKitConstitutionCarriesSharedBaseline(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "docs", "CONSTITUTION.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), templates.ConstitutionBaselineSection) {
		t.Fatal("docs/CONSTITUTION.md baseline diverges from templates.ConstitutionBaselineSection; run kit init --refresh")
	}
}
