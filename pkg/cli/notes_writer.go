package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/feature"
)

func applyNotesShortcuts(options *notesOptions) {
	privateRequested := options.private ||
		strings.EqualFold(strings.TrimSpace(options.section), "private") ||
		strings.EqualFold(strings.TrimSpace(options.sensitivity), "private")
	if privateRequested {
		options.private = true
		options.section = "private"
		options.sensitivity = "private"
	}
	options.source = normalizeNoteField(options.source, "manual")
	options.status = normalizeNoteField(options.status, "active")
	options.sensitivity = normalizeNoteField(options.sensitivity, "internal")
	options.section = strings.ToLower(strings.TrimSpace(options.section))
}

func validateNotesOptions(options notesOptions) error {
	switch effectiveNoteSection(options) {
	case "inbox", "references", "responses", "private":
		return nil
	default:
		return fmt.Errorf("--section must be inbox, references, responses, or private")
	}
}

func createFeatureNoteFile(projectRoot, featureDirName string, options notesOptions) (string, error) {
	section := effectiveNoteSection(options)
	dir := filepath.Join(featureNotesPath(projectRoot, featureDirName), section)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create notes section: %w", err)
	}

	title := strings.TrimSpace(options.title)
	if title == "" {
		title = defaultNoteTitle(options)
	}
	name := noteFileName(notesNow(), title)
	path := uniqueNotePath(dir, name)
	if err := os.WriteFile(path, []byte(buildNoteTemplate(featureDirName, title, options)), 0644); err != nil {
		return "", fmt.Errorf("failed to create note: %w", err)
	}
	return path, nil
}

func buildNoteTemplate(featureDirName, title string, options notesOptions) string {
	capturedAt := notesNow().Format(time.RFC3339)
	return fmt.Sprintf(`---
kind: note
source: %s
status: %s
sensitivity: %s
captured_at: %s
feature: %s
---

# %s

## Context

## Raw Notes

## Durable Takeaways

Promote durable, non-sensitive decisions into canonical project docs before
treating them as implementation requirements.
`, options.source, options.status, options.sensitivity, capturedAt, featureDirName, title)
}

func noteFileName(now time.Time, title string) string {
	slug := noteFilenameSlug(title)
	if slug == "" {
		slug = "note"
	}
	return fmt.Sprintf("%s-%s.md", now.Format("2006-01-02-150405"), slug)
}

func uniqueNotePath(dir, name string) string {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	ext := filepath.Ext(name)
	path := filepath.Join(dir, name)
	for i := 2; ; i++ {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
		path = filepath.Join(dir, fmt.Sprintf("%s-%d%s", base, i, ext))
	}
}

func effectiveNoteSection(options notesOptions) string {
	if options.private {
		return "private"
	}
	return strings.ToLower(strings.TrimSpace(options.section))
}

func defaultNoteTitle(options notesOptions) string {
	if effectiveNoteSection(options) == "private" {
		return "Private note"
	}
	return "Feature note"
}

func normalizeNoteField(value, fallback string) string {
	normalized := feature.NormalizeSlug(value)
	if normalized == "" {
		return fallback
	}
	return normalized
}

func noteFilenameSlug(value string) string {
	normalized := feature.NormalizeSlug(value)
	parts := strings.Split(normalized, "-")
	if len(parts) > 8 {
		parts = parts[:8]
	}
	return strings.Join(parts, "-")
}

func relativeProjectPath(projectRoot, path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	rel, err := filepath.Rel(projectRoot, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

func printNotesResult(out io.Writer, result notesResult, copied bool) error {
	if _, err := fmt.Fprintf(out, "Feature notes ready: %s\n", result.NotesPath); err != nil {
		return err
	}
	if result.CreatedFeature {
		if _, err := fmt.Fprintf(out, "  ✓ Created feature directory: %s\n", result.Feature); err != nil {
			return err
		}
	}
	if result.NotePath != "" {
		if _, err := fmt.Fprintf(out, "  ✓ Created note: %s\n", result.NotePath); err != nil {
			return err
		}
		if result.Private {
			if _, err := fmt.Fprintln(out, "  ✓ Private note contents are ignored by git"); err != nil {
				return err
			}
		}
	}
	if copied {
		if _, err := fmt.Fprintln(out, "  ✓ Copied notes path to clipboard"); err != nil {
			return err
		}
	}
	return nil
}
