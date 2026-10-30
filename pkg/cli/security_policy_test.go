package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func TestConventionalSecurityPoliciesPreserved(t *testing.T) {
	for _, policy := range []string{".github/SECURITY.md", templates.SecurityPath, "docs/SECURITY.md"} {
		t.Run(policy, func(t *testing.T) {
			root := t.TempDir()
			setupInitHome(t)
			stubRulesetRegistry(t, registryRulesetForTest("delivery", []string{"git"}))
			setWorkingDirectory(t, root)
			custom := "# Vulnerability reporting\n\nReport privately through the project-owned channel.\n"
			writeFile(t, filepath.Join(root, filepath.FromSlash(policy)), custom)
			withInitFlags(t, func() {
				initOutputOnly = true
				captureStdout(t, func() {
					if err := runInitForTest(initCmd, nil); err != nil {
						t.Fatal(err)
					}
				})
			})
			assertPolicy := func() {
				t.Helper()
				if got := securityPolicyPath(root); got != policy {
					t.Fatalf("resolved %s, want %s", got, policy)
				}
				if got := readFile(t, filepath.Join(root, filepath.FromSlash(policy))); got != custom {
					t.Fatal("policy changed")
				}
				if policy != templates.SecurityPath {
					assertFileDoesNotExist(t, filepath.Join(root, templates.SecurityPath))
				}
				cfg, err := config.Load(root)
				if err != nil {
					t.Fatal(err)
				}
				if policy != templates.SecurityPath {
					if _, ok := cfg.RegistryArtifact(projectGuidanceKind, templates.SecurityPath); ok {
						t.Fatal("adopted absent root placeholder")
					}
					if _, ok := cfg.RegistryArtifact(projectGuidanceKind, policy); ok {
						t.Fatal("adopted project policy as generated guidance")
					}
				}
				findings := auditProjectGuidance(root, cfg)
				for _, finding := range findings {
					if !finding.NonBlocking {
						t.Fatalf("false blocking audit: %+v", finding)
					}
					if strings.Contains(finding.Issue, "missing") {
						t.Fatal(finding)
					}
				}
				registry, err := rulesetRegistryFetcher(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if !requiredStructurePresent(root, cfg, registry, nil) {
					t.Fatal("existing policy did not satisfy project structure")
				}
				for _, path := range instructionFiles(cfg) {
					content := readFile(t, filepath.Join(root, path))
					a, b, c := strings.Index(content, "`.github/SECURITY.md`"), strings.Index(content, "`SECURITY.md`"), strings.Index(content, "`docs/SECURITY.md`")
					if a < 0 || b < a || c < b {
						t.Fatalf("%s lacks policy precedence", path)
					}
				}
			}
			assertPolicy()
			before := readFile(t, filepath.Join(root, ".kit.yaml"))
			output := captureStdout(t, func() {
				if _, err := runInitRefreshWithSnapshot(root, initRefreshOptions{dryRun: true, diff: true, files: []string{templates.SecurityPath}}); err != nil {
					t.Fatal(err)
				}
			})
			if policy != templates.SecurityPath && !strings.Contains(output, "preserved project-owned security policy at "+policy) {
				t.Fatal(output)
			}
			if readFile(t, filepath.Join(root, ".kit.yaml")) != before {
				t.Fatal("preview changed ownership")
			}
			assertPolicy()
			captureStdout(t, func() {
				if _, err := runInitRefreshWithSnapshot(root, initRefreshOptions{outputOnly: true}); err != nil {
					t.Fatal(err)
				}
			})
			assertPolicy()
			if policy != templates.SecurityPath {
				// Explicit root replacement leaves the existing non-root policy intact.
				output = captureStdout(t, func() {
					if _, err := runInitRefreshWithSnapshot(root, initRefreshOptions{force: true, dryRun: true, diff: true, files: []string{templates.SecurityPath}}); err != nil {
						t.Fatal(err)
					}
				})
				if !strings.Contains(output, "diff --git a/SECURITY.md b/SECURITY.md") {
					t.Fatal(output)
				}
				assertFileDoesNotExist(t, filepath.Join(root, templates.SecurityPath))
				captureStdout(t, func() {
					if _, err := runInitRefreshWithSnapshot(root, initRefreshOptions{force: true, outputOnly: true, files: []string{templates.SecurityPath}}); err != nil {
						t.Fatal(err)
					}
				})
				if readFile(t, filepath.Join(root, templates.SecurityPath)) != templates.Security {
					t.Fatal("explicit root target was not created")
				}
				if readFile(t, filepath.Join(root, filepath.FromSlash(policy))) != custom {
					t.Fatal("non-root policy overwritten")
				}
			}
		})
	}
}

func TestSecurityPolicyPrecedence(t *testing.T) {
	root := t.TempDir()
	if got := securityPolicyPath(root); got != templates.SecurityPath {
		t.Fatal(got)
	}
	for _, path := range []string{"docs/SECURITY.md", templates.SecurityPath, ".github/SECURITY.md"} {
		writeFile(t, filepath.Join(root, filepath.FromSlash(path)), "# Existing policy\n")
		if got := securityPolicyPath(root); got != path {
			t.Fatalf("resolved %s, want %s", got, path)
		}
	}
	// A present but unreadable/dangling policy must not be hidden by a new file.
	if err := os.Remove(filepath.Join(root, ".github/SECURITY.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing.md", filepath.Join(root, ".github/SECURITY.md")); err != nil {
		t.Fatal(err)
	}
	if got := securityPolicyPath(root); got != ".github/SECURITY.md" {
		t.Fatal(got)
	}
}
