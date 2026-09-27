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

	once, changed := upsertConstitutionBaseline(input)
	if !changed {
		t.Fatal("first upsert changed = false, want true")
	}
	if !strings.Contains(once, customConstraint) {
		t.Fatalf("first upsert removed custom constraint:\n%s", once)
	}

	twice, changed := upsertConstitutionBaseline(once)
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

	refreshed, changed := upsertConstitutionBaseline(fresh)
	if changed || refreshed != strings.TrimRight(fresh, "\n")+"\n" {
		t.Fatalf("refresh changed a freshly initialized Constitution:\n%s", refreshed)
	}

	// A project refreshed by an older Kit may lack baseline bullets; refresh
	// must restore the complete fresh-init baseline rather than a subset.
	var stale []string
	for _, line := range strings.Split(fresh, "\n") {
		if strings.Contains(line, "deletion") {
			continue
		}
		stale = append(stale, line)
	}
	restored, changed := upsertConstitutionBaseline(strings.Join(stale, "\n"))
	if !changed {
		t.Fatal("refresh left a stale baseline unchanged")
	}
	if !strings.Contains(restored, templates.ConstitutionBaselineSection) {
		t.Fatalf("refresh did not restore the fresh-init baseline:\n%s", restored)
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
