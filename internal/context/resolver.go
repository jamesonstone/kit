package context

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

type resolver struct {
	root          string
	contract      Contract
	evidenceIndex map[string]int
	workflowState map[string]string
}

func Resolve(projectRoot string, request Request) Contract {
	request.Workflow = strings.TrimSpace(request.Workflow)
	if request.Workflow == "" {
		request.Workflow = "implementation-delivery"
	}
	request.Feature = strings.TrimSpace(request.Feature)
	request.Paths = normalizeHints(request.Paths)
	contract := Contract{
		SchemaVersion: SchemaVersion,
		Request:       request,
		Workflows:     []SelectedWorkflow{},
		Evidence:      []EvidenceItem{},
		Diagnostics:   []Diagnostic{},
		NextActions:   []string{},
	}
	root, err := filepath.Abs(projectRoot)
	if err != nil {
		contract.Diagnostics = append(contract.Diagnostics, Diagnostic{Level: "error", Code: "invalid-project-root", Message: err.Error()})
		return finalize(contract)
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		contract.Diagnostics = append(contract.Diagnostics, Diagnostic{Level: "error", Code: "invalid-project-root", Message: err.Error()})
		return finalize(contract)
	}
	r := &resolver{root: filepath.Clean(root), contract: contract, evidenceIndex: map[string]int{}, workflowState: map[string]string{}}
	if !config.Exists(r.root) {
		r.addDiagnostic("error", "project-not-initialized", config.ConfigFileName, "Kit project configuration is missing")
		return finalize(r.contract)
	}
	r.resolveWorkflow(request.Workflow, nil)
	if request.Feature != "" {
		r.resolveFeature(request.Feature)
	}
	for _, hint := range request.Paths {
		r.addEvidence("path", hint, true, "explicit path hint")
	}
	return finalize(r.contract)
}

func (r *resolver) resolveWorkflow(slug string, stack []string) {
	if state := r.workflowState[slug]; state == "done" {
		return
	} else if state == "visiting" {
		r.addDiagnostic("error", "workflow-cycle", "", "workflow dependency cycle: "+strings.Join(append(stack, slug), " -> "))
		return
	}
	r.workflowState[slug] = "visiting"
	manifest, path, data, err := loadWorkflow(r.root, slug)
	if path == "" {
		path = filepath.ToSlash(filepath.Join("docs", "references", "workflows", slug+".md"))
	}
	r.addEvidence("workflow", path, true, "selected workflow "+slug)
	if err != nil {
		code := "invalid-workflow"
		if errors.Is(err, os.ErrNotExist) {
			code = "missing-workflow"
		}
		r.addDiagnostic("error", code, path, err.Error())
		r.workflowState[slug] = "done"
		return
	}
	for _, dependency := range manifest.Dependencies {
		r.resolveWorkflow(dependency, append(stack, slug))
	}
	r.contract.Workflows = append(r.contract.Workflows, SelectedWorkflow{
		Slug: slug, Path: path, Description: manifest.Description,
		Dependencies: append([]string{}, manifest.Dependencies...), Digest: digest(data),
	})
	for _, evidence := range manifest.Evidence {
		r.addEvidence(evidence.Kind, evidence.Path, evidence.Required, "workflow "+slug)
	}
	for _, rule := range manifest.Rules {
		r.addRule(rule, slug)
	}
	r.workflowState[slug] = "done"
}

func (r *resolver) addRule(rule WorkflowRule, workflow string) {
	path := filepath.ToSlash(filepath.Join("docs", "references", "rules", rule.Slug+".md"))
	data := r.addEvidence("rule", path, rule.Required, "workflow "+workflow+" rule "+rule.Slug)
	if len(data) == 0 {
		return
	}
	if err := validateArtifactHeader(data, "ruleset", rule.Slug); err != nil {
		level := "warning"
		if rule.Required {
			level = "error"
		}
		r.addDiagnostic(level, "invalid-rule", path, err.Error())
	}
}

func (r *resolver) resolveFeature(reference string) {
	cfg, err := config.Load(r.root)
	if err != nil {
		r.addDiagnostic("error", "invalid-project-config", config.ConfigFileName, err.Error())
		return
	}
	feat, err := feature.Resolve(cfg.SpecsPath(r.root), reference)
	if err != nil {
		r.addDiagnostic("error", "missing-feature", cfg.SpecsDir, err.Error())
		return
	}
	specPath := filepath.ToSlash(filepath.Join(cfg.SpecsDir, feat.DirName, "SPEC.md"))
	data := r.addEvidence("feature-spec", specPath, true, "selected feature "+feat.DirName)
	r.addEvidence("project-index", "docs/PROJECT_PROGRESS_SUMMARY.md", false, "feature discovery index")
	if len(data) == 0 {
		return
	}
	doc, err := document.ParseFile(filepath.Join(r.root, filepath.FromSlash(specPath)), document.TypeSpec)
	if err != nil {
		r.addDiagnostic("error", "invalid-feature-spec", specPath, err.Error())
		return
	}
	for _, diagnostic := range doc.MetadataDiagnostics {
		if diagnostic.Severity == document.MetadataDiagnosticError {
			r.addDiagnostic("error", "invalid-feature-spec", specPath, diagnostic.Message)
		}
	}
	if doc.Metadata == nil {
		return
	}
	for _, relationship := range doc.Metadata.Relationships {
		r.resolveHistoricalSpec(cfg, relationship, specPath)
	}
	for _, reference := range doc.Metadata.References {
		r.resolveFeatureReference(reference, specPath)
	}
}

func (r *resolver) resolveHistoricalSpec(cfg *config.Config, relationship document.MetadataRelationship, source string) {
	feat, err := feature.Resolve(cfg.SpecsPath(r.root), strings.TrimSpace(relationship.Target))
	if err != nil {
		r.addDiagnostic("error", "missing-related-spec", source, fmt.Sprintf("relationship %s target %q is unresolved", relationship.Type, relationship.Target))
		return
	}
	path := filepath.ToSlash(filepath.Join(cfg.SpecsDir, feat.DirName, "SPEC.md"))
	r.addEvidence("historical-spec", path, true, fmt.Sprintf("%s relationship from %s", relationship.Type, source))
}

func (r *resolver) resolveFeatureReference(reference document.MetadataReference, source string) {
	if reference.ReadPolicy == document.ReferenceReadPolicySkip || reference.Status == document.ReferenceStatusStale {
		return
	}
	target := strings.Trim(strings.TrimSpace(reference.Target), "`\"'")
	if !isLocalReference(target) {
		return
	}
	required := reference.ReadPolicy == document.ReferenceReadPolicyMust
	kind := strings.TrimSpace(reference.Type)
	if kind == "" {
		kind = "reference"
	}
	r.addEvidence(kind, target, required, fmt.Sprintf("feature reference %s from %s", reference.Name, source))
}

func isLocalReference(target string) bool {
	if target == "" || strings.Contains(target, "|") || strings.Contains(target, "*") {
		return false
	}
	lower := strings.ToLower(target)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "git@") {
		return false
	}
	return strings.Contains(target, "/") || strings.HasSuffix(target, ".md") || filepath.IsAbs(target)
}

func normalizeHints(paths []string) []string {
	seen := map[string]bool{}
	var normalized []string
	for _, path := range paths {
		path = filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
		if path == "." || path == "" || seen[path] {
			continue
		}
		seen[path] = true
		normalized = append(normalized, path)
	}
	sort.Strings(normalized)
	return normalized
}

func digest(data []byte) string {
	value := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(value[:])
}

func finalize(contract Contract) Contract {
	for index := range contract.Evidence {
		sort.Strings(contract.Evidence[index].Reasons)
	}
	for _, diagnostic := range contract.Diagnostics {
		if diagnostic.Level == "error" {
			contract.Blocked = true
			break
		}
	}
	if contract.Blocked {
		contract.NextActions = []string{"Restore or repair required local evidence, then rerun kit context resolve."}
	} else {
		contract.NextActions = []string{"Load required evidence in order before acting.", "Rerun after a material scope or workflow change."}
	}
	return contract
}
