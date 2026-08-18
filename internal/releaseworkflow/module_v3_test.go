package releaseworkflow_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestModuleAndBuildSurfacesUseV3Identity(t *testing.T) {
	checks := map[string][]string{
		"go.mod": {
			"module github.com/jamesonstone/kit/v3",
		},
		"Makefile": {
			"-X github.com/jamesonstone/kit/v3/pkg/cli.Version=$(VERSION)",
		},
		".goreleaser.yaml": {
			"-X github.com/jamesonstone/kit/v3/pkg/cli.Version={{ .Version }}",
		},
		"README.md": {
			"go install github.com/jamesonstone/kit/v3/cmd/kit@latest",
			"https://github.com/jamesonstone/kit/releases",
		},
		"pkg/cli/upgrade_support.go": {
			`releasesAPIBase   = "https://api.github.com/repos/jamesonstone/kit/releases"`,
			"go install github.com/jamesonstone/kit/v3/cmd/kit@latest",
		},
	}

	for path, required := range checks {
		content := readRepositoryFile(t, path)
		for _, want := range required {
			if !strings.Contains(content, want) {
				t.Errorf("%s missing %q", path, want)
			}
		}
	}
}

func TestGoSelfImportsUseV3Identity(t *testing.T) {
	root := filepath.Join("..", "..")
	moduleRoot := "github.com/jamesonstone/kit"

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			if strings.HasPrefix(importPath, moduleRoot+"/") &&
				!strings.HasPrefix(importPath, moduleRoot+"/v3/") {
				t.Errorf("%s contains unversioned self-import %q", path, importPath)
			}
			if strings.HasPrefix(importPath, moduleRoot+"/v3/v3/") {
				t.Errorf("%s contains duplicated v3 self-import %q", path, importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
