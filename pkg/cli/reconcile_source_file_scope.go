package cli

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"strings"
)

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
	if !sourceFileMetadataInScope(relativePath, info) || bytes.IndexByte(data, 0) >= 0 || generatedSourceContent(data) {
		return false
	}
	return filepath.Ext(relativePath) != "" || bytes.HasPrefix(data, []byte("#!"))
}

func sourceFileMetadataInScope(relativePath string, info fs.FileInfo) bool {
	extension := filepath.Ext(relativePath)
	return recognizedSourceExtension(extension) || extension == "" && info.Mode()&0o111 != 0
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
