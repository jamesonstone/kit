package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

const projectGuidanceKind = "project-guidance"

// securityPolicyPath follows conventional policy precedence and defaults to
// the root location only when the project has no existing policy.
func securityPolicyPath(projectRoot string) string {
	for _, path := range []string{".github/SECURITY.md", templates.SecurityPath, "docs/SECURITY.md"} {
		if _, err := os.Lstat(filepath.Join(projectRoot, filepath.FromSlash(path))); !os.IsNotExist(err) {
			return path
		}
	}
	return templates.SecurityPath
}

func preserveSecurityPolicy(projectRoot, path string, force bool) string {
	if path == templates.SecurityPath && !force {
		if policy := securityPolicyPath(projectRoot); policy != templates.SecurityPath {
			return policy
		}
	}
	return ""
}

// Inspect current bytes rather than a stale on-disk registry classification.
func currentGuidanceState(projectRoot string, artifact config.RegistryArtifact) string {
	data, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(artifact.Path)))
	if err != nil {
		return "unknown"
	}
	canonical, ok := templates.ProjectGuidance(artifact.Path)
	if ok && (string(data) == canonical || kitWroteUnmodified(artifact, contentHash(string(data)))) {
		return registryArtifactStateManaged
	}
	return registryArtifactStateLocalCustom
}

// Whole-document ownership uses the same installed-hash evidence as rulesets.
// Existing project policies are never adopted merely because their names match.
func planProjectGuidance(plan instructionFileWritePlan, existing string, force bool, cfg *config.Config) instructionFileWritePlan {
	canonical, _ := templates.ProjectGuidance(filepath.ToSlash(plan.relativePath))
	state, _ := cfg.RegistryArtifact(projectGuidanceKind, plan.relativePath)
	if existing == canonical {
		return plan
	}
	if force || kitWroteUnmodified(state, contentHash(existing)) {
		plan.content = canonical
		plan.result = instructionFileUpdated
		return plan
	}
	plan.note = fmt.Sprintf("%s has project-owned content and was preserved; review Kit defaults before explicitly replacing it with `kit reconcile --force --file %s`", plan.relativePath, plan.relativePath)
	return plan
}

func recordProjectGuidance(cfg *config.Config, path, content string) bool {
	canonical, ok := templates.ProjectGuidance(filepath.ToSlash(path))
	if !ok {
		return false
	}
	state, _ := cfg.RegistryArtifact(projectGuidanceKind, path)
	next := config.RegistryArtifact{Kind: projectGuidanceKind, Slug: path, Path: path, State: registryArtifactStateLocalCustom}
	if content == canonical {
		next.State = registryArtifactStateManaged
		next.InstalledHash = contentHash(content)
	}
	if state.Kind == next.Kind && state.Slug == next.Slug && state.Path == next.Path && state.State == next.State && state.InstalledHash == next.InstalledHash && len(state.Sections) == 0 {
		return false
	}
	cfg.Registry.SchemaVersion = registryArtifactSchemaVersion
	cfg.UpsertRegistryArtifact(next)
	return true
}

// Validate presence, readability, and a Markdown heading only. Project policies
// and vulnerability-reporting content need semantic review, not a policy engine.
func auditProjectGuidance(projectRoot string, cfg *config.Config) []reconcileFinding {
	var findings []reconcileFinding
	for _, path := range []string{templates.DesignPath, securityPolicyPath(projectRoot)} {
		absolute := filepath.Join(projectRoot, path)
		data, err := os.ReadFile(absolute)
		issue := ""
		nonBlocking := false
		switch {
		case os.IsNotExist(err):
			issue = "missing project guidance document"
		case err != nil:
			issue = "cannot read project guidance: " + err.Error()
		default:
			content := string(data)
			if !hasMarkdownHeading(content) {
				issue = "project guidance needs nonempty Markdown with a heading"
			} else {
				plan := planProjectGuidance(instructionFileWritePlan{relativePath: path}, content, false, cfg)
				if plan.result == instructionFileUpdated {
					issue = "unchanged Kit guidance differs from the current defaults"
				} else if plan.note != "" {
					issue = "project-owned guidance preserved; review upstream defaults when upgrading"
					nonBlocking = true
				}
			}
		}
		if issue != "" {
			fix := "preview with `kit reconcile --dry-run --diff --file " + path + "`; preserve project policy when reviewing"
			if path != templates.DesignPath && path != templates.SecurityPath {
				fix = "review this project-owned policy in place; Kit preserves it"
			}
			finding := newFinding(reconcileSeverityWarning, absolute, issue, fix)
			finding.NonBlocking = nonBlocking
			findings = append(findings, finding)
		}
	}
	return findings
}

// hasMarkdownHeading checks minimal ATX or Setext structure, not full Markdown syntax.
func hasMarkdownHeading(content string) bool {
	lines := strings.Split(content, "\n")
	for index, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		title := strings.TrimLeft(line, "#")
		if len(title) < len(line) && strings.HasPrefix(title, " ") && strings.TrimSpace(title) != "" {
			return true
		}
		if index > 0 && isSetextUnderline(line) && strings.TrimSpace(lines[index-1]) != "" {
			return true
		}
	}
	return false
}

// isSetextUnderline recognizes a nonempty uniform equals or hyphen underline.
func isSetextUnderline(line string) bool {
	return line != "" && (strings.Trim(line, "=") == "" || strings.Trim(line, "-") == "")
}
