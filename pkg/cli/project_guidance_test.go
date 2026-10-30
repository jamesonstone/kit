package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func TestInitProjectGuidance(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(map[bool]string{false: "fresh", true: "existing-policy"}[existing], func(t *testing.T) {
			root := t.TempDir()
			setupInitHome(t)
			setWorkingDirectory(t, root)
			custom := "# Private reporting\n\nUse the existing private channel; support the current release.\n"
			if existing {
				writeFile(t, filepath.Join(root, templates.SecurityPath), custom)
			}
			withInitFlags(t, func() {
				initOutputOnly = true
				captureStdout(t, func() {
					if err := runInitForTest(initCmd, nil); err != nil {
						t.Fatal(err)
					}
				})
			})
			cfg, err := config.Load(root)
			if err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{templates.DesignPath, templates.SecurityPath} {
				want, _ := templates.ProjectGuidance(path)
				if existing && path == templates.SecurityPath {
					want = custom
				}
				if got := readFile(t, filepath.Join(root, path)); got != want {
					t.Fatalf("%s was not preserved/scaffolded", path)
				}
				state, ok := cfg.RegistryArtifact(projectGuidanceKind, path)
				if !ok {
					t.Fatalf("missing ownership for %s", path)
				}
				if existing && path == templates.SecurityPath {
					if state.State != registryArtifactStateLocalCustom {
						t.Fatal(state)
					}
				} else if !kitWroteUnmodified(state, contentHash(want)) {
					t.Fatal(state)
				}
			}
			for _, path := range instructionFiles(cfg) {
				content := readFile(t, filepath.Join(root, path))
				for _, ref := range []string{"`DESIGN.md`", "`SECURITY.md`"} {
					if !strings.Contains(content, ref) {
						t.Fatalf("%s missing %s", path, ref)
					}
				}
				if strings.Contains(content, "Never hardcode credentials") {
					t.Fatal("provider duplicates security guidance")
				}
			}
		})
	}
}

func TestReconcileProjectGuidanceLifecycle(t *testing.T) {
	for _, path := range []string{templates.DesignPath, templates.SecurityPath} {
		t.Run(path, func(t *testing.T) {
			root := t.TempDir()
			setupInitHome(t)
			cfg := config.Default()
			if err := config.Save(root, cfg); err != nil {
				t.Fatal(err)
			}
			canonical, _ := templates.ProjectGuidance(path)
			refresh := func(opts initRefreshOptions) string {
				t.Helper()
				return captureStdout(t, func() {
					if _, err := runInitRefreshWithSnapshot(root, opts); err != nil {
						t.Fatal(err)
					}
				})
			}
			opts := initRefreshOptions{files: []string{path}}
			dry := opts
			dry.dryRun = true
			dry.diff = true
			before := readFile(t, filepath.Join(root, ".kit.yaml"))
			output := refresh(dry)
			if !strings.Contains(output, "diff --git a/"+path+" b/"+path) {
				t.Fatal(output)
			}
			assertFileDoesNotExist(t, filepath.Join(root, path))
			if readFile(t, filepath.Join(root, ".kit.yaml")) != before {
				t.Fatal("dry run changed registry")
			}
			refresh(opts)
			if readFile(t, filepath.Join(root, path)) != canonical {
				t.Fatal("reconcile did not create canonical guidance")
			}
			cfg, err := config.Load(root)
			if err != nil {
				t.Fatal(err)
			}
			state, ok := cfg.RegistryArtifact(projectGuidanceKind, path)
			if !ok || !kitWroteUnmodified(state, contentHash(canonical)) {
				t.Fatalf("targeted refresh lost ownership: %+v", state)
			}
			// Simulate a default shipped by an earlier release, with installed evidence.
			old := "# Earlier default\n\nEarlier Kit engineering guidance.\n"
			writeFile(t, filepath.Join(root, path), old)
			state.InstalledHash = contentHash(old)
			cfg.UpsertRegistryArtifact(state)
			if err := config.Save(root, cfg); err != nil {
				t.Fatal(err)
			}
			refresh(opts)
			if readFile(t, filepath.Join(root, path)) != canonical {
				t.Fatal("unchanged installed default did not upgrade")
			}
			custom := "# Project policy\n\nPrivate reporting and project-specific constraints.\n"
			writeFile(t, filepath.Join(root, path), custom)
			output = refresh(opts)
			if !strings.Contains(output, "preserved") || readFile(t, filepath.Join(root, path)) != custom {
				t.Fatal("customization was not preserved")
			}
			// A whole-project --force does not grant replacement of project policies.
			refresh(initRefreshOptions{force: true, outputOnly: true})
			if readFile(t, filepath.Join(root, path)) != custom {
				t.Fatal("blanket force replaced project guidance")
			}
			forced := opts
			forced.force = true
			dry = forced
			dry.dryRun = true
			dry.diff = true
			before = readFile(t, filepath.Join(root, ".kit.yaml"))
			output = refresh(dry)
			if !strings.Contains(output, "-# Project policy") {
				t.Fatal(output)
			}
			if readFile(t, filepath.Join(root, path)) != custom || readFile(t, filepath.Join(root, ".kit.yaml")) != before {
				t.Fatal("forced preview wrote files")
			}
			refresh(forced)
			if readFile(t, filepath.Join(root, path)) != canonical {
				t.Fatal("explicit replacement did not converge")
			}
		})
	}
}

func TestProjectGuidanceAudit(t *testing.T) {
	root := setupCoherentProjectForCheck(t)
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{templates.DesignPath, templates.SecurityPath} {
		canonical, _ := templates.ProjectGuidance(path)
		writeFile(t, filepath.Join(root, path), canonical)
	}
	for _, test := range []struct {
		name, content    string
		remove, blocking bool
	}{
		{name: "missing", remove: true, blocking: true},
		{name: "empty", content: "", blocking: true},
		{name: "no-heading", content: "Project policy", blocking: true},
		{name: "custom", content: "# Local policy\n\nPrivate reporting follows project policy.\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(root, templates.SecurityPath)
			if test.remove {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			} else {
				writeFile(t, path, test.content)
			}
			findings := auditProjectGuidance(root, cfg)
			if len(findings) != 1 || findings[0].NonBlocking == test.blocking {
				t.Fatalf("unexpected findings: %+v", findings)
			}
			var out bytes.Buffer
			err := checkProjectContractTo(&out, root, cfg)
			if (err != nil) != test.blocking {
				t.Fatalf("check error %v; output %s", err, out.String())
			}
		})
	}
}

func TestProjectGuidanceLinkedTargetPreserved(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	external := filepath.Join(t.TempDir(), "security.md")
	custom := "# External policy\n\nProject-owned private channel.\n"
	writeFile(t, external, custom)
	if err := os.Symlink(external, filepath.Join(root, templates.SecurityPath)); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	if err := config.Save(root, cfg); err != nil {
		t.Fatal(err)
	}
	captureStdout(t, func() {
		if _, err := runInitRefreshWithSnapshot(root, initRefreshOptions{force: true, files: []string{templates.SecurityPath}}); err != nil {
			t.Fatal(err)
		}
	})
	if readFile(t, external) != custom {
		t.Fatal("external linked policy was overwritten")
	}
}

func TestProjectGuidanceRegistryStatus(t *testing.T) {
	root := t.TempDir()
	cfg := config.Default()
	for _, path := range []string{templates.DesignPath, templates.SecurityPath} {
		content, _ := templates.ProjectGuidance(path)
		writeFile(t, filepath.Join(root, path), content)
		recordProjectGuidance(cfg, path, content)
	}
	summary := &statusKitManagedSummary{}
	recordStatusLocalRegistry(root, cfg, summary)
	if summary.Registry.Total != 2 || summary.Registry.Managed != 2 {
		t.Fatal(summary.Registry)
	}
	custom := "# Private project policy\n"
	writeFile(t, filepath.Join(root, templates.SecurityPath), custom)
	recordProjectGuidance(cfg, templates.SecurityPath, custom)
	summary = &statusKitManagedSummary{}
	recordStatusLocalRegistry(root, cfg, summary)
	if summary.Registry.LocalCustom != 1 || len(summary.Items) != 1 || summary.Items[0].Kind != "registry-project-guidance" {
		t.Fatalf("unexpected status %+v", summary)
	}
}
