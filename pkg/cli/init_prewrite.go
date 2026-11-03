package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/templates"
)

type prewriteMode int

const (
	prewriteInit prewriteMode = iota
	prewriteReconcile
)

type issueSurface struct {
	path    string
	content string
	warning string
}

func gatePrewrite(root string, changes []initRefreshFileChange, mode prewriteMode, interactive, force, dryRun bool, in io.Reader, out io.Writer) ([]initRefreshFileChange, map[string]string, error) {
	surfaces := issueTemplateSurfaces(root)
	notes := map[string]string{}
	for i, change := range changes {
		if !needsPrewriteConfirm(change, mode, surfaces) {
			continue
		}
		if force {
			notes[change.relativePath] = "force"
			continue
		}
		if dryRun && interactive {
			notes[change.relativePath] = "would ask before writing"
			continue
		}
		if interactive && !dryRun {
			ok, err := confirmPrompt(in, out, fmt.Sprintf("Write %s?", change.relativePath))
			if err != nil {
				return nil, nil, err
			}
			if ok {
				continue
			}
		}
		notes[change.relativePath] = "existing file; pass --force to write it"
		if change.result == instructionFileCreated {
			notes[change.relativePath] = "existing issue setup; pass --force to add the template"
		}
		changes[i].result = instructionFileSkipped
		changes[i].after = changes[i].before
	}
	return changes, notes, nil
}

func needsPrewriteConfirm(change initRefreshFileChange, mode prewriteMode, surfaces []issueSurface) bool {
	issueCreate := change.relativePath == issueTemplatePath && change.result == instructionFileCreated && len(surfaces) > 0
	if issueCreate {
		return true
	}
	if mode == prewriteReconcile {
		return false
	}
	switch change.result {
	case instructionFileUpdated, instructionFileMerged, instructionFileRemoved:
		return true
	default:
		return false
	}
}

func confirmPrompt(in io.Reader, out io.Writer, question string) (bool, error) {
	if _, err := fmt.Fprintf(out, "%s [y/N]: ", question); err != nil {
		return false, err
	}
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && line == "" {
		return false, fmt.Errorf("read write confirmation: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func printSurfaceWarnings(out io.Writer, surfaces []issueSurface) error {
	for _, surface := range surfaces {
		if surface.path == issueTemplatePath && templates.IsShippedIssueTemplate(surface.content) {
			continue
		}
		if _, err := fmt.Fprintln(out, "Warning: "+surface.warning); err != nil {
			return err
		}
	}
	return nil
}

func prewriteWarnings(changes []initRefreshFileChange, surfaces []issueSurface) []string {
	var warnings []string
	for _, change := range changes {
		switch change.result {
		case instructionFileUpdated, instructionFileMerged, instructionFileRemoved:
			warnings = append(warnings, "existing file will change: "+change.relativePath)
		}
	}
	for _, surface := range surfaces {
		if surface.path == issueTemplatePath && templates.IsShippedIssueTemplate(surface.content) {
			continue
		}
		warnings = append(warnings, surface.warning)
	}
	return warnings
}

func issueTemplateSurfaces(root string) []issueSurface {
	var found []issueSurface
	for _, rel := range []string{".github/ISSUE_TEMPLATE.md", "ISSUE_TEMPLATE.md", "docs/ISSUE_TEMPLATE.md"} {
		if data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel))); err == nil {
			found = append(found, issueSurface{path: rel, content: string(data), warning: "existing issue template: " + rel})
		}
	}
	for _, dir := range []string{".github/ISSUE_TEMPLATE", "ISSUE_TEMPLATE", "docs/ISSUE_TEMPLATE"} {
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			rel := dir + "/" + entry.Name()
			data, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
			content := ""
			if readErr == nil {
				content = string(data)
			}
			found = append(found, issueSurface{path: rel, content: content, warning: issueSurfaceWarning(rel, entry.Name())})
		}
	}
	sort.Slice(found, func(i, j int) bool { return found[i].path < found[j].path })
	return found
}

func gateReconcilePlan(root string, changes []initRefreshFileChange, force, dryRun bool) ([]initRefreshFileChange, map[string]string, error) {
	interactive := streamsHaveInteractiveTerminal(os.Stdin, os.Stdout)
	return gatePrewrite(root, changes, prewriteReconcile, interactive, force, dryRun, os.Stdin, os.Stdout)
}

func issueSurfaceWarning(rel, name string) string {
	lower := strings.ToLower(name)
	switch {
	case lower == "config.yml" || lower == "config.yaml":
		return "existing issue template config: " + rel
	case strings.HasSuffix(lower, ".yml") || strings.HasSuffix(lower, ".yaml"):
		return "existing issue form: " + rel
	default:
		return "existing issue template: " + rel
	}
}

func orgIssueTemplateWarning(root string, changes []initRefreshFileChange) string {
	writes := false
	for _, change := range changes {
		if change.relativePath == issueTemplatePath && (change.result == instructionFileCreated || change.result == instructionFileUpdated) {
			writes = true
		}
	}
	if !writes {
		return ""
	}
	owner := gitHubRemoteOwner(root)
	if owner == "" {
		return "Kit cannot detect issue templates inherited from an organization .github repository."
	}
	switch strings.ToLower(readmeGitHubOwnerType(root, owner)) {
	case "organization":
		return "GitHub owner " + owner + " is an organization. Its .github repository can supply issue templates, and a local template overrides them. Kit does not download those inherited templates."
	case "":
		return "Kit cannot detect issue templates inherited from an organization .github repository."
	default:
		return ""
	}
}

func gitHubRemoteOwner(root string) string {
	output, err := readmeCommandRunner.Output(root, "git", "remote", "get-url", "origin")
	if err != nil {
		return ""
	}
	owner, _, err := parseGitHubRemoteURL(strings.TrimSpace(string(output)))
	if err != nil {
		return ""
	}
	return owner
}

func printPrewritePlan(out io.Writer, changes []initRefreshFileChange, notes map[string]string) error {
	if _, err := fmt.Fprintln(out, "Pre-write plan:"); err != nil {
		return err
	}
	for _, change := range changes {
		note := notes[change.relativePath]
		if note == "" && change.relativePath == issueTemplatePath && change.result == instructionFileSkipped && change.before == "" {
			note = "opt-in; pass --issue-template"
		}
		line := fmt.Sprintf("  %s %s", prewriteLabel(change.result), change.relativePath)
		if note != "" {
			line += " (" + note + ")"
		}
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	return nil
}

func printPrewriteWarnings(out io.Writer, warnings []string, org string) error {
	for _, warning := range warnings {
		if _, err := fmt.Fprintln(out, "Warning: "+warning); err != nil {
			return err
		}
	}
	if org != "" {
		if _, err := fmt.Fprintln(out, "Warning: "+org); err != nil {
			return err
		}
	}
	return nil
}

func prewriteLabel(result instructionFileWriteResult) string {
	switch result {
	case instructionFileCreated:
		return "create"
	case instructionFileUpdated, instructionFileMerged:
		return "overwrite"
	case instructionFileRemoved:
		return "remove"
	default:
		return "skip"
	}
}

func recountInitRefreshStats(changes []initRefreshFileChange) initRefreshStats {
	var stats initRefreshStats
	for _, change := range changes {
		stats.recordFileChange(change)
	}
	return stats
}
