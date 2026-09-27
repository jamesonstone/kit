package context

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (r *resolver) addEvidence(kind, path string, required bool, reason string) []byte {
	relativePath, data, err := secureRead(r.root, path)
	if relativePath == "" {
		relativePath = filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
	}
	if index, ok := r.evidenceIndex[relativePath]; ok {
		item := &r.contract.Evidence[index]
		wasRequired := item.Required
		item.Required = item.Required || required
		item.Reasons = appendUnique(item.Reasons, reason)
		if required && !wasRequired && item.State != "present" {
			r.addDiagnostic("error", "missing-evidence", relativePath, "evidence became required through another selected context")
		}
		return data
	}
	item := EvidenceItem{Kind: kind, Path: relativePath, Required: required, State: "present", Reasons: []string{reason}}
	if err != nil {
		item.State = "missing"
		level := "warning"
		if required {
			level = "error"
		}
		code := "missing-evidence"
		if !errors.Is(err, os.ErrNotExist) {
			code = "invalid-evidence-path"
		}
		r.addDiagnostic(level, code, relativePath, err.Error())
	} else {
		item.Digest = digest(data)
	}
	r.evidenceIndex[relativePath] = len(r.contract.Evidence)
	r.contract.Evidence = append(r.contract.Evidence, item)
	return data
}

// addPathHint records an explicit --path hint. A hint scopes the task, so a
// path the agent is about to create, or a directory, is recorded without
// blocking resolution; hints that escape the project root still fail.
func (r *resolver) addPathHint(hint string) {
	relativePath, _, err := secureRead(r.root, hint)
	if err == nil {
		r.addEvidence("path", hint, true, "explicit path hint")
		return
	}
	state := ""
	switch {
	case errors.Is(err, os.ErrNotExist) && !missingPathEscapes(r.root, hint):
		state = "absent"
	case err.Error() == errEvidenceDirectory:
		state = "directory"
	default:
		r.addEvidence("path", hint, true, "explicit path hint")
		return
	}
	if _, ok := r.evidenceIndex[relativePath]; ok {
		return
	}
	r.evidenceIndex[relativePath] = len(r.contract.Evidence)
	r.contract.Evidence = append(r.contract.Evidence, EvidenceItem{Kind: "path", Path: relativePath, State: state, Reasons: []string{"explicit path hint"}})
	r.addDiagnostic("info", "path-hint-"+state, relativePath, "path hint recorded as "+state+"; nothing to load")
}

// missingPathEscapes reports whether the nearest existing entry for a missing
// hint, starting with the hint itself, resolves outside the project root or
// cannot be resolved, such as a dangling symlink or a planned file beneath a
// symlink that points elsewhere.
func missingPathEscapes(root, value string) bool {
	path := strings.TrimSpace(value)
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, filepath.FromSlash(path))
	}
	for dir := filepath.Clean(path); ; dir = filepath.Dir(dir) {
		if _, err := os.Lstat(dir); err == nil {
			resolved, err := filepath.EvalSymlinks(dir)
			if err != nil {
				return true
			}
			relative, err := filepath.Rel(root, resolved)
			return err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator))
		}
		if parent := filepath.Dir(dir); parent == dir {
			return true
		}
	}
}

const errEvidenceDirectory = "evidence path is a directory"

func secureRead(root, value string) (string, []byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil, fmt.Errorf("evidence path is blank")
	}
	path := value
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, filepath.FromSlash(value))
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.ToSlash(value), nil, err
	}
	relative, err := filepath.Rel(root, abs)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return filepath.ToSlash(value), nil, fmt.Errorf("evidence path escapes the project root")
	}
	relative = filepath.ToSlash(relative)
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return relative, nil, err
	}
	resolvedRelative, err := filepath.Rel(root, resolved)
	if err != nil || resolvedRelative == ".." || strings.HasPrefix(resolvedRelative, ".."+string(filepath.Separator)) {
		return relative, nil, fmt.Errorf("evidence symlink escapes the project root")
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return relative, nil, err
	}
	if info.IsDir() {
		return relative, nil, errors.New(errEvidenceDirectory)
	}
	data, err := os.ReadFile(resolved)
	return relative, data, err
}

func (r *resolver) addDiagnostic(level, code, path, message string) {
	r.contract.Diagnostics = append(r.contract.Diagnostics, Diagnostic{Level: level, Code: code, Path: path, Message: message})
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
