package cli

import (
	"strings"
	"testing"
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
