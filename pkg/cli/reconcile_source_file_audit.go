package cli

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type sourceFileAuditSummary struct {
	Limit          int
	CandidateCount int
	EligibleCount  int
	ViolationCount int
	Complete       bool
}

type sourceFileAuditResult struct {
	Summary  sourceFileAuditSummary
	Findings []reconcileFinding
}

func auditSourceFileSizes(projectRoot string, limit int) []reconcileFinding {
	return inspectSourceFileSizes(projectRoot, limit).Findings
}

// inspectSourceFileSizes enforces a project-configured line limit; a zero
// limit (the default) disables the audit.
func inspectSourceFileSizes(projectRoot string, limit int) sourceFileAuditResult {
	if limit <= 0 {
		return sourceFileAuditResult{}
	}
	paths, err := sourceFileAuditCandidates(projectRoot)
	if err != nil {
		return sourceFileAuditResult{Summary: sourceFileAuditSummary{Limit: limit}, Findings: []reconcileFinding{newFinding(
			reconcileSeverityError,
			filepath.Join(projectRoot, ".git"),
			fmt.Sprintf("source-file-size audit unavailable: %v", err),
			"restore version-control-eligible file enumeration, then rerun whole-project reconcile; do not claim a clean line audit until enumeration succeeds",
		)}}
	}

	result := sourceFileAuditResult{Summary: sourceFileAuditSummary{
		Limit:          limit,
		CandidateCount: len(paths),
		Complete:       true,
	}}
	for _, relativePath := range paths {
		finding, ok, eligible := auditSourceFileSize(projectRoot, relativePath, limit)
		if eligible {
			result.Summary.EligibleCount++
		}
		if ok {
			result.Findings = append(result.Findings, finding)
			if finding.Severity == reconcileSeverityError {
				result.Summary.Complete = false
			}
			if finding.AllowsCodeChanges {
				result.Summary.ViolationCount++
			}
		}
	}
	return result
}

func auditSourceFileSize(projectRoot, relativePath string, limit int) (reconcileFinding, bool, bool) {
	if sourceFilePathExcluded(relativePath) {
		return reconcileFinding{}, false, false
	}
	absPath := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
	info, err := os.Lstat(absPath)
	if os.IsNotExist(err) || (err == nil && !info.Mode().IsRegular()) {
		return reconcileFinding{}, false, false
	}
	if err != nil {
		return sourceFileReadFinding(projectRoot, absPath, err), true, false
	}
	if !sourceFileMetadataInScope(relativePath, info) {
		return reconcileFinding{}, false, false
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return sourceFileReadFinding(projectRoot, absPath, err), true, false
	}
	if !sourceFileContentInScope(relativePath, info, data) {
		return reconcileFinding{}, false, false
	}
	lineCount := physicalLineCount(data)
	if lineCount <= limit {
		return reconcileFinding{}, false, true
	}

	finding := newFinding(
		reconcileSeverityWarning,
		absPath,
		fmt.Sprintf("version-control-eligible handwritten source/test file exceeds %d physical lines (%d)", limit, lineCount),
		fmt.Sprintf("split the file by semantic responsibility until every resulting handwritten source/test file is at most %d physical lines; preserve behavior, stable public entry points, and language-native test discovery, and use responsibility-based filenames", limit),
	)
	finding.AllowsCodeChanges = true
	return finding, true, true
}

func sourceFileAuditEvidence(summary *sourceFileAuditSummary) string {
	if summary == nil || summary.Limit <= 0 {
		return ""
	}
	state := "complete"
	if !summary.Complete {
		state = "incomplete; clean result prohibited"
	}
	return fmt.Sprintf(
		"source-file-size audit: %s (%d version-control-eligible candidates; %d eligible handwritten source/test files checked; %d above %d physical lines)",
		state,
		summary.CandidateCount,
		summary.EligibleCount,
		summary.ViolationCount,
		summary.Limit,
	)
}

func sourceFileReadFinding(projectRoot, absPath string, err error) reconcileFinding {
	return newFinding(
		reconcileSeverityError,
		absPath,
		fmt.Sprintf("source-file-size audit could not read candidate file: %v", err),
		"restore file readability and rerun whole-project reconcile before claiming a clean source-file-size audit",
	)
}

func sourceFileAuditCandidates(projectRoot string) ([]string, error) {
	if _, err := os.Lstat(filepath.Join(projectRoot, ".git")); err == nil {
		return gitSourceFileAuditCandidates(projectRoot)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("inspect .git: %w", err)
	}
	return filesystemSourceFileAuditCandidates(projectRoot)
}

func gitSourceFileAuditCandidates(projectRoot string) ([]string, error) {
	cmd := exec.Command("git", "-C", projectRoot, "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list Git candidates: %w", err)
	}
	return normalizedSourceAuditPaths(bytes.Split(output, []byte{0})), nil
}

func filesystemSourceFileAuditCandidates(projectRoot string) ([]string, error) {
	var paths [][]byte
	err := filepath.WalkDir(projectRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(projectRoot, path)
		if err != nil {
			return err
		}
		if relativePath == "." {
			return nil
		}
		relativePath = filepath.ToSlash(relativePath)
		if entry.IsDir() && sourceFilePathExcluded(relativePath) {
			return filepath.SkipDir
		}
		if !entry.IsDir() {
			paths = append(paths, []byte(relativePath))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk project candidates: %w", err)
	}
	return normalizedSourceAuditPaths(paths), nil
}

func normalizedSourceAuditPaths(rawPaths [][]byte) []string {
	seen := make(map[string]bool, len(rawPaths))
	for _, rawPath := range rawPaths {
		if len(rawPath) == 0 {
			continue
		}
		path := filepath.Clean(filepath.FromSlash(string(rawPath)))
		if filepath.IsAbs(path) || path == ".." || strings.HasPrefix(path, ".."+string(filepath.Separator)) {
			continue
		}
		seen[filepath.ToSlash(path)] = true
	}
	paths := make([]string, 0, len(seen))
	for path := range seen {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}
