package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

func requiredRulesetSections() []string {
	return []string{"Purpose", "Applies When", "Rules", "Anti-Patterns", "Verification", "Examples"}
}

func validRulesetStatus(value string) bool {
	switch value {
	case document.ReferenceStatusActive, document.ReferenceStatusOptional, document.ReferenceStatusStale:
		return true
	default:
		return false
	}
}

func validRulesetReadPolicy(value string) bool {
	switch value {
	case document.ReferenceReadPolicyMust,
		document.ReferenceReadPolicyConditional,
		document.ReferenceReadPolicyEvidence,
		document.ReferenceReadPolicySkip:
		return true
	default:
		return false
	}
}

func rulesetReference(slug, readPolicy string) document.MetadataReference {
	return document.MetadataReference{
		ID:         rulesetReferenceIDPrefix + slug,
		Name:       "Ruleset: " + slug,
		Type:       rulesetReferenceType,
		Target:     rulesetTarget(slug),
		Relation:   document.ReferenceRelationGuides,
		ReadPolicy: readPolicy,
		UsedFor:    "load durable " + slug + " rules only when relevant to the current decision",
		Status:     rulesetReferenceStatus,
	}
}

func rulesetLinkTargetDoc(feat *feature.Feature) (string, document.DocumentType, error) {
	candidates := []struct {
		name    string
		docType document.DocumentType
	}{
		{name: "SPEC.md", docType: document.TypeSpec},
		{name: "PLAN.md", docType: document.TypePlan},
		{name: "BRAINSTORM.md", docType: document.TypeBrainstorm},
		{name: "TASKS.md", docType: document.TypeTasks},
	}
	for _, candidate := range candidates {
		path := filepath.Join(feat.Path, candidate.name)
		if document.Exists(path) {
			return path, candidate.docType, nil
		}
	}
	return "", "", fmt.Errorf("feature %q has no document that can hold ruleset references", feat.Slug)
}

func featureRulesetReferenceErrors(projectRoot string, doc *document.Document) []string {
	var errors []string
	for _, reference := range doc.References() {
		if !isRulesetReference(reference) {
			continue
		}
		path, ok := rulesetReferencePath(projectRoot, reference.Target)
		if !ok {
			errors = append(errors, fmt.Sprintf("%s: ruleset reference %q must target docs/references/rules/<slug>.md", doc.Path, reference.Name))
			continue
		}
		if !document.Exists(path) {
			errors = append(errors, fmt.Sprintf("%s: ruleset reference %q points to missing file %s", doc.Path, reference.Name, filepath.ToSlash(strings.TrimSpace(reference.Target))))
			continue
		}
		ruleset, err := parseRulesetFile(path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to read ruleset reference %q: %v", doc.Path, reference.Name, err))
			continue
		}
		expectedSlug := strings.TrimSuffix(filepath.Base(path), ".md")
		if issues := validateRulesetDocument(ruleset, expectedSlug); len(issues) > 0 {
			errors = append(errors, fmt.Sprintf("%s: ruleset reference %q points to invalid ruleset: %s", doc.Path, reference.Name, strings.Join(issues, "; ")))
		}
	}
	return errors
}

func rulesetReferencePath(projectRoot, target string) (string, bool) {
	cleanTarget := filepath.Clean(filepath.FromSlash(strings.TrimSpace(target)))
	if cleanTarget == "." || strings.TrimSpace(target) == "" {
		return "", false
	}
	var absPath string
	var relPath string
	if filepath.IsAbs(cleanTarget) {
		absPath = cleanTarget
		rel, err := filepath.Rel(projectRoot, absPath)
		if err != nil {
			return "", false
		}
		relPath = rel
	} else {
		relPath = cleanTarget
		absPath = filepath.Join(projectRoot, relPath)
	}
	relSlash := filepath.ToSlash(filepath.Clean(relPath))
	if !strings.HasPrefix(relSlash, rulesetDirRelPath+"/") || !strings.HasSuffix(relSlash, ".md") {
		return "", false
	}
	return absPath, true
}

func isRulesetReference(reference document.MetadataReference) bool {
	return strings.EqualFold(strings.TrimSpace(reference.Type), rulesetReferenceType) ||
		strings.HasPrefix(filepath.ToSlash(strings.TrimSpace(reference.Target)), rulesetDirRelPath+"/")
}

func auditRulesets(projectRoot string) []reconcileFinding {
	dir := filepath.Join(projectRoot, filepath.FromSlash(rulesetDirRelPath))
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return []reconcileFinding{newFinding(
			reconcileSeverityError,
			dir,
			"failed to read ruleset directory",
			templateSource(projectRoot),
			"fix docs/references/rules/ permissions before validating rulesets",
			[]string{fmt.Sprintf("ls -la %s", dir)},
		)}
	}

	var findings []reconcileFinding
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		ruleset, err := parseRulesetFile(path)
		if err != nil {
			findings = append(findings, newFinding(
				reconcileSeverityError,
				path,
				"failed to read ruleset document",
				templateSource(projectRoot),
				"make the ruleset readable and retry validation",
				[]string{fmt.Sprintf("sed -n '1,220p' %s", path)},
			))
			continue
		}
		expectedSlug := strings.TrimSuffix(entry.Name(), ".md")
		for _, issue := range validateRulesetDocument(ruleset, expectedSlug) {
			findings = append(findings, newFinding(
				reconcileSeverityError,
				path,
				"ruleset document issue: "+issue,
				templateSource(projectRoot),
				"update the ruleset front matter and required sections to match the Kit ruleset contract",
				[]string{fmt.Sprintf("sed -n '1,220p' %s", path)},
			))
		}
	}
	return findings
}

func auditRulesetReferences(projectRoot string, path string, doc *document.Document) []reconcileFinding {
	var findings []reconcileFinding
	for _, issue := range featureRulesetReferenceErrors(projectRoot, doc) {
		findings = append(findings, newFinding(
			reconcileSeverityError,
			path,
			issue,
			templateSource(projectRoot),
			"create the referenced ruleset with `kit rules add <slug>` or update the feature reference target",
			[]string{
				fmt.Sprintf("sed -n '1,90p' %s", path),
				fmt.Sprintf("ls %s", filepath.Join(projectRoot, filepath.FromSlash(rulesetDirRelPath))),
			},
		))
	}
	return findings
}

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
