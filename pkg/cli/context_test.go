package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	stdreflect "reflect"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/config"
	contextcontract "github.com/jamesonstone/kit/v3/internal/context"
)

func TestRunContextResolveEmitsReadyDeterministicJSONWithoutWrites(t *testing.T) {
	root := setupContextCLIProject(t, false)
	setWorkingDirectory(t, root)
	before := contextCLISnapshot(t, root)

	first := resolveContextJSON(t, &contextResolveOptions{workflow: "test", jsonOutput: true})
	second := resolveContextJSON(t, &contextResolveOptions{workflow: "test", jsonOutput: true})
	if !stdreflect.DeepEqual(first, second) {
		t.Fatalf("context JSON changed across resolutions:\n%#v\n%#v", first, second)
	}
	if first.SchemaVersion != "kit.context/v1" || first.Blocked {
		t.Fatalf("unexpected resolved contract: %#v", first)
	}
	after := contextCLISnapshot(t, root)
	if !stdreflect.DeepEqual(before, after) {
		t.Fatalf("read-only resolution changed project files:\nbefore=%#v\nafter=%#v", before, after)
	}
}

func TestRunContextResolveReturnsBlockedJSONAndExitTwo(t *testing.T) {
	root := setupContextCLIProject(t, true)
	setWorkingDirectory(t, root)
	var output bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&output)
	err := runContextResolve(cmd, &contextResolveOptions{workflow: "test", jsonOutput: true})
	var exitErr *cliExitError
	if !errors.As(err, &exitErr) || exitErr.code != 2 || !exitErr.silent {
		t.Fatalf("blocked error = %#v, want silent exit code 2", err)
	}
	var contract contextcontract.Contract
	if decodeErr := json.Unmarshal(output.Bytes(), &contract); decodeErr != nil {
		t.Fatalf("blocked JSON is invalid: %v\n%s", decodeErr, output.String())
	}
	if !contract.Blocked {
		t.Fatalf("blocked result = %#v", contract)
	}
}

func setupContextCLIProject(t *testing.T, missingEvidence bool) string {
	t.Helper()
	root := t.TempDir()
	if err := config.Save(root, config.Default()); err != nil {
		t.Fatal(err)
	}
	requiredPath := "evidence.md"
	if missingEvidence {
		requiredPath = "missing.md"
	}
	writeFile(t, filepath.Join(root, "docs/references/workflows/test.md"), `---
kind: workflow
slug: test
description: CLI resolution fixture
rules:
  - slug: local
    required: true
evidence:
  - kind: source
    path: `+requiredPath+`
    required: true
---
# Test workflow
`)
	writeFile(t, filepath.Join(root, "docs/references/rules/local.md"), `---
kind: ruleset
slug: local
description: Local test rule
status: active
---
# Ruleset: local
`)
	if !missingEvidence {
		writeFile(t, filepath.Join(root, requiredPath), "evidence\n")
	}
	return root
}

func resolveContextJSON(t *testing.T, options *contextResolveOptions) contextcontract.Contract {
	t.Helper()
	var output bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&output)
	if err := runContextResolve(cmd, options); err != nil {
		t.Fatalf("runContextResolve() error = %v", err)
	}
	var contract contextcontract.Contract
	if err := json.Unmarshal(output.Bytes(), &contract); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, output.String())
	}
	return contract
}

func contextCLISnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(root, path)
		result[filepath.ToSlash(relative)] = string(content)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
