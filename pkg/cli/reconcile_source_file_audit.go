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

const sourceFileLineLimit = 300

type sourceFileAuditSummary struct {
	CandidateCount int
	EligibleCount  int
	ViolationCount int
	Complete       bool
}

type sourceFileAuditResult struct {
	Summary  sourceFileAuditSummary
	Findings []reconcileFinding
}

func auditSourceFileSizes(projectRoot string) []reconcileFinding {
	return inspectSourceFileSizes(projectRoot).Findings
}

func inspectSourceFileSizes(projectRoot string) sourceFileAuditResult {
	paths, err := sourceFileAuditCandidates(projectRoot)
	if err != nil {
		return sourceFileAuditResult{Findings: []reconcileFinding{newFinding(
			reconcileSeverityError,
			filepath.Join(projectRoot, ".git"),
			fmt.Sprintf("source-file-size audit unavailable: %v", err),
			sourceFileSizeRuleSource(projectRoot),
			"restore version-control-eligible file enumeration, then rerun whole-project reconcile; do not claim a clean line audit until enumeration succeeds",
			[]string{"git status --short --branch", "git ls-files --cached --others --exclude-standard"},
		)}}
	}

	result := sourceFileAuditResult{Summary: sourceFileAuditSummary{
		CandidateCount: len(paths),
		Complete:       true,
	}}
	for _, relativePath := range paths {
		finding, ok, eligible := auditSourceFileSize(projectRoot, relativePath)
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

func auditSourceFileSize(projectRoot, relativePath string) (reconcileFinding, bool, bool) {
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
	data, err := os.ReadFile(absPath)
	if err != nil {
		return sourceFileReadFinding(projectRoot, absPath, err), true, false
	}
	if !sourceFileContentInScope(relativePath, info, data) {
		return reconcileFinding{}, false, false
	}
	lineCount := physicalLineCount(data)
	if lineCount <= sourceFileLineLimit {
		return reconcileFinding{}, false, true
	}

	finding := newFinding(
		reconcileSeverityWarning,
		absPath,
		fmt.Sprintf("version-control-eligible handwritten source/test file exceeds 300 physical lines (%d)", lineCount),
		sourceFileSizeRuleSource(projectRoot),
		"split the file by semantic responsibility until every resulting handwritten source/test file is at most 300 physical lines; preserve behavior, stable public entry points, and language-native test discovery, and use responsibility-based filenames",
		[]string{
			fmt.Sprintf("awk 'END { print NR }' %s", shellQuoteArgument(absPath)),
			fmt.Sprintf("git diff -- %s", shellQuoteArgument(relativePath)),
		},
	)
	finding.AllowsCodeChanges = true
	return finding, true, true
}

func sourceFileAuditEvidence(summary *sourceFileAuditSummary) string {
	if summary == nil {
		return ""
	}
	state := "complete"
	if !summary.Complete {
		state = "incomplete; clean result prohibited"
	}
	return fmt.Sprintf(
		"source-file-size audit: %s (%d version-control-eligible candidates; %d eligible handwritten source/test files checked; %d above 300 physical lines)",
		state,
		summary.CandidateCount,
		summary.EligibleCount,
		summary.ViolationCount,
	)
}

func sourceFileReadFinding(projectRoot, absPath string, err error) reconcileFinding {
	return newFinding(
		reconcileSeverityError,
		absPath,
		fmt.Sprintf("source-file-size audit could not read candidate file: %v", err),
		sourceFileSizeRuleSource(projectRoot),
		"restore file readability and rerun whole-project reconcile before claiming a clean source-file-size audit",
		[]string{fmt.Sprintf("ls -l %s", shellQuoteArgument(absPath))},
	)
}

func sourceFileSizeRuleSource(projectRoot string) string {
	return filepath.Join(projectRoot, "docs", "references", "rules", "source-file-size.md")
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

func sourceFilePathExcluded(relativePath string) bool {
	cleanPath := filepath.ToSlash(filepath.Clean(relativePath))
	if cleanPath == ".kit.yaml" {
		return true
	}
	for _, part := range strings.Split(cleanPath, "/") {
		switch strings.ToLower(part) {
		case ".git", ".kit", "docs", "node_modules", "third_party", "third-party", "vendor":
			return true
		}
	}
	base := strings.ToLower(filepath.Base(cleanPath))
	return strings.Contains(base, ".min.") || strings.Contains(base, ".bundle.")
}

func sourceFileContentInScope(relativePath string, info fs.FileInfo, data []byte) bool {
	if bytes.IndexByte(data, 0) >= 0 || generatedSourceContent(data) {
		return false
	}
	if recognizedSourceExtension(filepath.Ext(relativePath)) {
		return true
	}
	return filepath.Ext(relativePath) == "" && info.Mode()&0o111 != 0 && bytes.HasPrefix(data, []byte("#!"))
}

func generatedSourceContent(data []byte) bool {
	if len(data) > 4096 {
		data = data[:4096]
	}
	lines := bytes.Split(data, []byte{'\n'})
	if len(lines) > 20 {
		lines = lines[:20]
	}
	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if !generatedMarkerComment(trimmed) {
			continue
		}
		lower := bytes.ToLower(trimmed)
		if bytes.Contains(lower, []byte("@generated")) ||
			(bytes.Contains(lower, []byte("do not edit")) &&
				(bytes.Contains(lower, []byte("code generated")) ||
					bytes.Contains(lower, []byte("automatically generated")) ||
					bytes.Contains(lower, []byte("file was generated")))) {
			return true
		}
	}
	return false
}

func generatedMarkerComment(line []byte) bool {
	for _, prefix := range [][]byte{
		[]byte("//"), []byte("#"), []byte("/*"), []byte("*"), []byte("<!--"), []byte("--"),
	} {
		if bytes.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func recognizedSourceExtension(extension string) bool {
	switch strings.ToLower(extension) {
	case ".astro", ".bash", ".bat", ".c", ".cc", ".clj", ".cljs", ".cljc", ".cmd",
		".cpp", ".cs", ".css", ".cts", ".cxx", ".dart", ".ejs", ".erl", ".ex", ".exs",
		".fish", ".fs", ".fsx", ".go", ".gql", ".graphql", ".h", ".hbs", ".hh", ".hpp",
		".hrl", ".hs", ".htm", ".html", ".hxx", ".java", ".jinja", ".jinja2", ".jl",
		".js", ".jsx", ".kt", ".kts", ".less", ".lhs", ".lua", ".m", ".mm", ".mts",
		".mustache", ".nix", ".php", ".pl", ".proto", ".ps1", ".py", ".pyi", ".r", ".rb",
		".rs", ".sass", ".scala", ".scss", ".sh", ".sol", ".sql", ".svelte", ".swift",
		".tf", ".tmpl", ".tpl", ".ts", ".tsx", ".vb", ".vue", ".xml", ".zig", ".zsh":
		return true
	default:
		return false
	}
}

func physicalLineCount(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	count := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		count++
	}
	return count
}
