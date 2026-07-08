package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/verify"
)

func TestBuildLoopReflectVerdictUsesRawCommandEvidence(t *testing.T) {
	projectRoot := t.TempDir()
	feat := writeReflectVerdictSpec(t, projectRoot, "0001-alpha", []string{"pkg/cli/loop_reflect_verdict.go"})
	now := time.Unix(1700001800, 0).UTC()
	runner := fakeReflectEvidenceRunner{
		"make test": {
			ExitCode: 1,
			Stdout:   "agent summary says tests_pass=true\n",
			Stderr:   "FAIL ./...\n",
		},
		"make lint": {
			ExitCode: 0,
		},
		"git merge-base HEAD origin/main": {
			ExitCode: 0,
			Stdout:   "base\n",
		},
		"git diff --name-only base...HEAD": {
			ExitCode: 0,
			Stdout:   "docs/specs/0001-alpha/SPEC.md\npkg/cli/loop_reflect_verdict.go\n",
		},
		"git diff --name-only": {
			ExitCode: 0,
		},
		"git diff --name-only --cached": {
			ExitCode: 0,
		},
		"git ls-files --others --exclude-standard": {
			ExitCode: 0,
		},
		"git log --format=%H%x00%ct -- docs/specs/0001-alpha/SPEC.md": {
			ExitCode: 0,
			Stdout:   "readyhash\x001700000000\n",
		},
		"git show readyhash:docs/specs/0001-alpha/SPEC.md": {
			ExitCode: 0,
			Stdout:   validV2SpecWithPhase("0001-alpha", "ready"),
		},
		"git log --format=%H readyhash..HEAD -- docs/specs/0001-alpha/SPEC.md pkg/cli/loop_reflect_verdict.go": {
			ExitCode: 0,
			Stdout:   "rework1\n",
		},
	}

	verdict, err := buildLoopReflectVerdict(context.Background(), reflectVerdictOptions{
		ProjectRoot: projectRoot,
		Feature:     feat,
		Runner:      runner,
		Now:         now,
	})
	if err != nil {
		t.Fatalf("buildLoopReflectVerdict() error = %v", err)
	}
	if verdict.TestsPass {
		t.Fatal("TestsPass = true, want false from failing make test exit code")
	}
	if verdict.LintDelta != 0 {
		t.Fatalf("LintDelta = %d, want 0", verdict.LintDelta)
	}
	if verdict.ScopeDrift != "none" {
		t.Fatalf("ScopeDrift = %q, want none", verdict.ScopeDrift)
	}
	if verdict.CycleTimeMin != 30 {
		t.Fatalf("CycleTimeMin = %d, want 30", verdict.CycleTimeMin)
	}
	if verdict.ReworkCount != 1 {
		t.Fatalf("ReworkCount = %d, want 1", verdict.ReworkCount)
	}
	if verdict.PromptVersion != "" {
		t.Fatalf("PromptVersion = %q, want empty", verdict.PromptVersion)
	}
}

func TestParseLintIssueCount(t *testing.T) {
	output := strings.Join([]string{
		"pkg/cli/a.go:12:6: missing error check (errcheck)",
		"pkg/cli/b.go:21:2: shadow: declaration of \"err\" shadows declaration at line 18 (govet)",
		"make: *** [lint] Error 1",
	}, "\n")

	if got := parseLintIssueCount(output); got != 2 {
		t.Fatalf("parseLintIssueCount() = %d, want 2", got)
	}
}

func TestClassifyReflectScopeDriftTiers(t *testing.T) {
	declared := []string{"pkg/a.go", "pkg/b.go"}
	tests := []struct {
		name    string
		touched []string
		want    string
	}{
		{name: "none", touched: []string{"pkg/a.go", "pkg/b.go"}, want: "none"},
		{name: "minor", touched: []string{"pkg/a.go", "pkg/b.go", "pkg/c.go", "pkg/d.go"}, want: "minor"},
		{name: "major unlisted", touched: []string{"pkg/a.go", "pkg/b.go", "pkg/c.go", "pkg/d.go", "pkg/e.go"}, want: "major"},
		{name: "major missed declared", touched: []string{"pkg/a.go"}, want: "major"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := classifyReflectScopeDrift(declared, tt.touched)
			if err != nil {
				t.Fatalf("classifyReflectScopeDrift() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("classifyReflectScopeDrift() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteLoopReflectVerdictFailsClosedWithoutReadyBoundary(t *testing.T) {
	projectRoot := t.TempDir()
	feat := writeReflectVerdictSpec(t, projectRoot, "0001-alpha", []string{"pkg/cli/loop_reflect_verdict.go"})
	runner := fakeReflectEvidenceRunner{
		"make test": {ExitCode: 0},
		"make lint": {ExitCode: 0},
		"git merge-base HEAD origin/main": {
			ExitCode: 0,
			Stdout:   "base\n",
		},
		"git diff --name-only base...HEAD": {
			ExitCode: 0,
			Stdout:   "pkg/cli/loop_reflect_verdict.go\n",
		},
		"git diff --name-only":                     {ExitCode: 0},
		"git diff --name-only --cached":            {ExitCode: 0},
		"git ls-files --others --exclude-standard": {ExitCode: 0},
		"git log --format=%H%x00%ct -- docs/specs/0001-alpha/SPEC.md": {
			ExitCode: 0,
			Stdout:   "laterhash\x001700001000\n",
		},
		"git show laterhash:docs/specs/0001-alpha/SPEC.md": {
			ExitCode: 0,
			Stdout:   validV2SpecWithPhase("0001-alpha", "implement"),
		},
	}

	_, err := writeLoopReflectVerdict(context.Background(), reflectVerdictOptions{
		ProjectRoot: projectRoot,
		Feature:     feat,
		Runner:      runner,
		Now:         time.Unix(1700001800, 0).UTC(),
	})
	if err == nil || !strings.Contains(err.Error(), "phase: ready boundary") {
		t.Fatalf("expected missing ready boundary error, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(feat.Path, reflectVerdictFileName)); !os.IsNotExist(statErr) {
		t.Fatalf("REFLECT.json should not be written on fail-closed path, stat err=%v", statErr)
	}
}

func TestDeclaredFilesFromSpecExpectedFileLines(t *testing.T) {
	content := strings.Join([]string{
		"## TASK CHECKLIST",
		"",
		"- [ ] T001. Expected files: `pkg/a.go`, `docs/specs/0001-alpha/SPEC.md`.",
		"- [ ] T002. expected-file lines also work: `pkg/b.go`.",
	}, "\n")
	got := declaredFilesFromSpec(content)
	want := []string{"docs/specs/0001-alpha/SPEC.md", "pkg/a.go", "pkg/b.go"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("declaredFilesFromSpec() = %#v, want %#v", got, want)
	}
}

type fakeReflectEvidenceRunner map[string]verify.CommandResult

func (runner fakeReflectEvidenceRunner) Run(_ context.Context, projectRoot string, commandID string, argv []string) verify.CommandResult {
	key := strings.Join(argv, " ")
	result, ok := runner[key]
	if !ok {
		return verify.CommandResult{
			CommandID: commandID,
			Argv:      append([]string(nil), argv...),
			Raw:       key,
			CWD:       projectRoot,
			ExitCode:  -1,
			Status:    "fail",
			Error:     "unexpected command: " + key,
		}
	}
	result.CommandID = commandID
	result.Argv = append([]string(nil), argv...)
	result.Raw = key
	result.CWD = projectRoot
	if result.Status == "" {
		if result.ExitCode == 0 {
			result.Status = "pass"
		} else {
			result.Status = "fail"
		}
	}
	return result
}

func writeReflectVerdictSpec(t *testing.T, projectRoot string, dirName string, expectedFiles []string) *feature.Feature {
	t.Helper()
	featurePath := filepath.Join(projectRoot, "docs", "specs", dirName)
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("MkdirAll(feature) error = %v", err)
	}
	var expected []string
	for _, file := range expectedFiles {
		expected = append(expected, "`"+file+"`")
	}
	content := validV2SpecWithPhase(dirName, "reflect") + "\n- Expected files: " + strings.Join(expected, ", ") + "\n"
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), content)
	_, slug, ok := strings.Cut(dirName, "-")
	if !ok {
		slug = dirName
	}
	return &feature.Feature{
		Slug:    slug,
		DirName: dirName,
		Path:    featurePath,
	}
}
