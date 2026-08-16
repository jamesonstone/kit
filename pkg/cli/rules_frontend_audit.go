package cli

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/document"
	"github.com/jamesonstone/kit/v3/internal/feature"
)

func auditActiveFrontendRulesetAdvisory(projectRoot string, feat *feature.Feature) []reconcileFinding {
	if feat == nil || feat.Paused || feat.Phase == feature.PhaseComplete {
		return nil
	}
	if !featureHasActiveFrontendProfileDependency(feat.Path) {
		return nil
	}
	if featureHasActiveRulesetForApplicability(projectRoot, feat, frontendRulesetAppliesTo) {
		return nil
	}

	return []reconcileFinding{newFinding(
		reconcileSeverityWarning,
		feat.Path,
		"active frontend feature has no active frontend ruleset reference",
		templateSource(projectRoot),
		"create or link a frontend ruleset with `kit rules add frontend-ui` and `kit rules link "+feat.Slug+" frontend-ui --read-policy conditional` if durable frontend rules apply",
		[]string{
			fmt.Sprintf("rg -n \"type: %s|%s|%s\" %s", rulesetReferenceType, rulesetDirRelPath, frontendProfileReferenceMarker, feat.Path),
			fmt.Sprintf("find %s -maxdepth 1 -type f -name '*.md' -print", filepath.Join(projectRoot, filepath.FromSlash(rulesetDirRelPath))),
		},
	)}
}

func featureHasActiveFrontendProfileDependency(featurePath string) bool {
	for _, source := range rulesetReferenceSources(featurePath) {
		if !document.Exists(source.path) {
			continue
		}
		doc, err := document.ParseFile(source.path, source.docType)
		if err != nil {
			continue
		}
		for _, reference := range doc.References() {
			if strings.EqualFold(reference.Name, "Frontend profile") &&
				strings.EqualFold(reference.Type, "profile") &&
				reference.Target == "--profile=frontend" &&
				reference.Status == document.ReferenceStatusActive {
				return true
			}
		}
	}
	return false
}

func canonicalFrontendProfileReferences(_ string) []document.MetadataReference {
	return []document.MetadataReference{{
		ID:         "frontend-profile",
		Name:       "Frontend profile",
		Type:       "profile",
		Target:     "--profile=frontend",
		Relation:   document.ReferenceRelationGuides,
		ReadPolicy: document.ReferenceReadPolicyConditional,
		UsedFor:    "apply frontend-specific coding-agent guidance",
		Status:     document.ReferenceStatusActive,
	}}
}

func featureHasActiveRulesetForApplicability(projectRoot string, feat *feature.Feature, appliesTo string) bool {
	for _, source := range rulesetReferenceSources(feat.Path) {
		if !document.Exists(source.path) {
			continue
		}
		doc, err := document.ParseFile(source.path, source.docType)
		if err != nil {
			continue
		}
		for _, reference := range doc.References() {
			if !activeRulesetReference(reference) {
				continue
			}
			path, ok := rulesetReferencePath(projectRoot, reference.Target)
			if !ok || !document.Exists(path) {
				continue
			}
			ruleset, err := parseRulesetFile(path)
			if err != nil || len(validateRulesetDocument(ruleset, strings.TrimSuffix(filepath.Base(path), ".md"))) > 0 {
				continue
			}
			if ruleset.Metadata.Status == document.ReferenceStatusActive && slices.Contains(ruleset.Metadata.AppliesTo, appliesTo) {
				return true
			}
		}
	}
	return false
}

func activeRulesetReference(reference document.MetadataReference) bool {
	return isRulesetReference(reference) &&
		reference.Status == document.ReferenceStatusActive &&
		reference.ReadPolicy != document.ReferenceReadPolicySkip
}

func rulesetReferenceSources(featurePath string) []struct {
	path    string
	docType document.DocumentType
} {
	return []struct {
		path    string
		docType document.DocumentType
	}{
		{path: filepath.Join(featurePath, "BRAINSTORM.md"), docType: document.TypeBrainstorm},
		{path: filepath.Join(featurePath, "SPEC.md"), docType: document.TypeSpec},
		{path: filepath.Join(featurePath, "PLAN.md"), docType: document.TypePlan},
		{path: filepath.Join(featurePath, "TASKS.md"), docType: document.TypeTasks},
	}
}
