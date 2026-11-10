// Package guidance provides a read-only inventory of current local guidance.
package guidance

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
)

type Entry struct {
	Kind     string    `json:"kind"`
	Path     string    `json:"path"`
	State    string    `json:"state"`
	Active   bool      `json:"active"`
	Metadata *Metadata `json:"metadata,omitempty"`
	Error    string    `json:"error,omitempty"`
}

type Report struct {
	SchemaVersion int               `json:"schema_version"`
	ProjectRoot   string            `json:"project_root"`
	Config        config.Inspection `json:"config"`
	Entries       []Entry           `json:"entries"`
	Errors        []string          `json:"errors"`
}

func Inspect(path string) (Report, error) {
	report := Report{SchemaVersion: 1, Entries: []Entry{}, Errors: []string{}}
	project, err := Resolve(path)
	if err != nil {
		return report, err
	}
	report.ProjectRoot = project
	root, err := os.OpenRoot(project)
	if err != nil {
		return report, err
	}
	defer func() { _ = root.Close() }()
	data, err := root.ReadFile(config.ConfigFileName)
	if err != nil {
		return report, fmt.Errorf("read .kit.yaml: %w", err)
	}
	cfg, inspection, err := config.ParseWithInspection(data)
	report.Config = inspection
	if err != nil {
		return report, err
	}
	for _, finding := range inspection.Findings {
		if finding.Severity == config.FindingError {
			report.Errors = append(report.Errors, finding.Message)
		}
	}
	if inspection.SchemaState == config.SchemaStateNewer {
		return report, fmt.Errorf("unsupported project configuration")
	}
	seen := map[string]bool{}
	add := func(kind, path string, required bool) {
		if filepath.IsLocal(path) {
			path = filepath.Clean(path)
		}
		key := kind + ":" + path
		if seen[key] {
			return
		}
		entry := inspectEntry(root, kind, path)
		if entry.State == "missing" && !required {
			return
		}
		seen[key] = true
		report.Entries = append(report.Entries, entry)
	}
	for _, path := range cfg.Agents {
		add("agent", path, true)
	}
	add("constitution", cfg.ConstitutionPath, true)
	for _, path := range []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"} {
		add("agent", path, false)
	}
	for _, path := range []string{"DESIGN.md", ".github/SECURITY.md", "SECURITY.md", "docs/SECURITY.md", "docs/references/testing.md", "docs/sops/README.md"} {
		add("guidance", path, false)
	}
	for _, artifact := range cfg.Registry.Artifacts {
		if artifact.Kind == "ruleset" || artifact.Kind == "sop" || artifact.Kind == "project-guidance" {
			kind := artifact.Kind
			if kind == "project-guidance" {
				kind = "guidance"
			}
			add(kind, artifact.Path, true)
		}
	}
	scan(root, "docs/references/rules", "ruleset", add, &report)
	scan(root, "docs/sops", "sop", add, &report)
	sort.Slice(report.Entries, func(i, j int) bool {
		a, b := report.Entries[i], report.Entries[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Path < b.Path
	})
	for _, entry := range report.Entries {
		if entry.Error != "" {
			report.Errors = append(report.Errors, entry.Path+": "+entry.Error)
		}
	}
	sort.Strings(report.Errors)
	if len(report.Errors) > 0 {
		return report, fmt.Errorf("project guidance has %d error(s)", len(report.Errors))
	}
	return report, nil
}

func inspectEntry(root *os.Root, kind, path string) Entry {
	entry := Entry{Kind: kind, Path: filepath.ToSlash(path), State: "invalid"}
	safe, err := safePath(root, path)
	if os.IsNotExist(err) {
		entry.State = "missing"
		return entry
	}
	if err == nil {
		var info os.FileInfo
		info, err = root.Stat(safe)
		if err == nil && !info.Mode().IsRegular() {
			err = fmt.Errorf("expected a regular file")
		}
	}
	if err == nil {
		// Opening also verifies access for guidance whose contents are not needed.
		var file *os.File
		file, err = root.Open(safe)
		if err == nil {
			err = file.Close()
		}
	}
	if err == nil && (kind == "ruleset" || kind == "sop") {
		var data []byte
		data, err = root.ReadFile(safe)
		if err == nil {
			entry.Metadata, err = parseMetadata(data, kind)
			if err == nil {
				want := strings.TrimSuffix(filepath.Base(path), ".md")
				if kind == "sop" {
					want = filepath.Base(filepath.Dir(path))
				}
				if entry.Metadata.Slug != want {
					err = fmt.Errorf("metadata slug %q does not match path slug %q", entry.Metadata.Slug, want)
				}
			}
		}
	}
	if err != nil {
		entry.Error = err.Error()
		return entry
	}
	entry.State = "present"
	if entry.Metadata != nil {
		entry.State = entry.Metadata.Status
		entry.Active = entry.State == "active"
	}
	return entry
}

func scan(root *os.Root, dir, kind string, add func(string, string, bool), report *Report) {
	safe, err := safePath(root, dir)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		report.Errors = append(report.Errors, dir+": "+err.Error())
		return
	}
	info, err := root.Stat(safe)
	if err == nil && !info.IsDir() {
		err = fmt.Errorf("expected a directory")
	}
	if err != nil {
		report.Errors = append(report.Errors, dir+": "+err.Error())
		return
	}
	file, err := root.Open(safe)
	if err != nil {
		report.Errors = append(report.Errors, dir+": "+err.Error())
		return
	}
	entries, err := file.ReadDir(-1)
	_ = file.Close()
	if err != nil {
		report.Errors = append(report.Errors, dir+": "+err.Error())
		return
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if kind == "ruleset" && strings.HasSuffix(entry.Name(), ".md") {
			add(kind, path, true)
		}
		if kind == "sop" && (entry.IsDir() || entry.Type()&os.ModeSymlink != 0) {
			add(kind, filepath.Join(path, "SOP.md"), true)
		}
	}
}
