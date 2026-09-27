package document

import "testing"

func TestStaleConstrainingReferenceHasNoContradictoryWarnings(t *testing.T) {
	diagnostics := referencePolicyDiagnostics("references[0]", MetadataReference{
		Relation: ReferenceRelationConstrains, ReadPolicy: ReferenceReadPolicySkip, Status: ReferenceStatusStale,
	})
	if len(diagnostics) != 0 {
		t.Fatalf("stale constraining reference produced %#v", diagnostics)
	}
}
