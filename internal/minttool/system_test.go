package minttool

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"testing"
)

func TestBuildInfoVerification(t *testing.T) {
	good := func() *debug.BuildInfo {
		return &debug.BuildInfo{Path: Command, Main: debug.Module{Path: Module, Version: "v0.8.0"}, Settings: []debug.BuildSetting{{Key: "GOOS", Value: runtime.GOOS}, {Key: "GOARCH", Value: runtime.GOARCH}}}
	}
	for _, tt := range []struct {
		name   string
		change func(*debug.BuildInfo)
	}{
		{"module", func(b *debug.BuildInfo) { b.Main.Path = "evil.example/mint" }},
		{"command", func(b *debug.BuildInfo) { b.Path = Module }},
		{"replace", func(b *debug.BuildInfo) { b.Main.Replace = &debug.Module{Path: "local"} }},
		{"version", func(b *debug.BuildInfo) { b.Main.Version = "(devel)" }},
		{"platform", func(b *debug.BuildInfo) { b.Settings[0].Value = "wrong" }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			b := good()
			tt.change(b)
			if verifyBuildInfo(b, "v0.8.0") == nil {
				t.Fatal("accepted invalid build")
			}
		})
	}
	if err := verifyBuildInfo(good(), "v0.8.0"); err != nil {
		t.Fatal(err)
	}
}

func tempBin(t *testing.T) string {
	t.Helper()
	if os.Getuid() < 0 {
		t.Skip("Unix ownership verification required")
	}
	p, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestFilesystemSafety(t *testing.T) {
	bin := tempBin(t)
	if err := safeDirectory(bin); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(bin, "mint")
	if err := os.WriteFile(p, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := safeFile(p); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(bin, "link")
	if err := os.Symlink(p, link); err != nil {
		t.Fatal(err)
	}
	if safeFile(link) == nil {
		t.Fatal("accepted symlink")
	}
	if err := os.Chmod(p, 0777); err != nil {
		t.Fatal(err)
	}
	if safeFile(p) == nil {
		t.Fatal("accepted writable executable")
	}
	if err := os.Chmod(bin, 0777); err != nil {
		t.Fatal(err)
	}
	if safeDirectory(bin) == nil {
		t.Fatal("accepted writable directory")
	}
}

func TestFailedInstallPreservesDestination(t *testing.T) {
	for _, badCandidate := range []bool{false, true} {
		t.Run(map[bool]string{false: "build_failure", true: "invalid_candidate"}[badCandidate], func(t *testing.T) {
			bin := tempBin(t)
			envJSON, _ := json.Marshal(map[string]string{"GOBIN": bin})
			script := "#!/bin/sh\nif [ \"$1\" = env ]; then printf '%s' \"$TEST_GO_JSON\"; exit 0; fi\nexit 1\n"
			if badCandidate {
				script = "#!/bin/sh\nif [ \"$1\" = env ]; then printf '%s' \"$TEST_GO_JSON\"; exit 0; fi\nprintf 'invalid candidate' > \"$GOBIN/mint\"\nchmod 755 \"$GOBIN/mint\"\n"
			}
			goPath := filepath.Join(bin, "go")
			if err := os.WriteFile(goPath, []byte(script), 0755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+"/bin:/usr/bin")
			t.Setenv("TEST_GO_JSON", string(envJSON))
			p := filepath.Join(bin, "mint")
			if err := os.WriteFile(p, []byte("original executable"), 0750); err != nil {
				t.Fatal(err)
			}
			old := Installation{Path: p, Version: "v0.7.0", Status: "installed", Method: "go-install"}
			b := systemBackend{verify: func(path, version string) error {
				if path == p {
					return nil
				}
				return verifyBinary(path, version)
			}}
			_, err := b.install(context.Background(), old, "v0.8.0")
			if err == nil {
				t.Fatal("expected failure")
			}
			content, _ := os.ReadFile(p)
			if string(content) != "original executable" {
				t.Fatal("old executable changed")
			}
			info, _ := os.Stat(p)
			if info.Mode().Perm() != 0750 {
				t.Fatal("old mode changed")
			}
			stages, _ := filepath.Glob(filepath.Join(bin, ".kit-mint-*"))
			if len(stages) != 0 {
				t.Fatal("staging files leaked")
			}
		})
	}
}

func TestMissingGoAndMint(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	i, err := (systemBackend{}).inspect(context.Background())
	if err != nil || i.Status != "missing" {
		t.Fatalf("%+v %v", i, err)
	}
	if _, err = goBin(context.Background()); err == nil {
		t.Fatal("missing Go accepted")
	}
}

func TestInitialInstallRequiresResolvedBin(t *testing.T) {
	bin := tempBin(t)
	t.Setenv("PATH", "/unrelated")
	if binInPath(bin) {
		t.Fatal("accepted bin absent from PATH")
	}
	t.Setenv("PATH", bin)
	if !binInPath(bin) {
		t.Fatal("rejected bin on PATH")
	}
}

func TestInstallRejectsStaleInspection(t *testing.T) {
	for _, replacedDuringVerification := range []bool{false, true} {
		t.Run(map[bool]string{false: "newer_before_snapshot", true: "replaced_during_verification"}[replacedDuringVerification], func(t *testing.T) {
			bin := tempBin(t)
			envJSON, _ := json.Marshal(map[string]string{"GOBIN": bin})
			script := "#!/bin/sh\nif [ \"$1\" = env ]; then printf '%s' \"$TEST_GO_JSON\"; exit 0; fi\nexit 99\n"
			if err := os.WriteFile(filepath.Join(bin, "go"), []byte(script), 0755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin)
			t.Setenv("TEST_GO_JSON", string(envJSON))
			target := filepath.Join(bin, "mint")
			if err := os.WriteFile(target, []byte("v0.9.0"), 0755); err != nil {
				t.Fatal(err)
			}
			verified := false
			b := systemBackend{verify: func(path, version string) error {
				verified = true
				if path != target || version != "v0.7.0" {
					t.Fatalf("unexpected verification: %s %s", path, version)
				}
				if !replacedDuringVerification {
					return errors.New("actual current version is v0.9.0")
				}
				replacement := filepath.Join(bin, "replacement")
				if err := os.WriteFile(replacement, []byte("v0.9.0"), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(replacement, target); err != nil {
					t.Fatal(err)
				}
				return nil
			}}
			old := Installation{Path: target, Version: "v0.7.0", Method: "go-install", Status: "installed"}
			_, err := b.install(context.Background(), old, "v0.8.0")
			if err == nil || !verified {
				t.Fatalf("stale inspection accepted: %v", err)
			}
			bytes, _ := os.ReadFile(target)
			if string(bytes) != "v0.9.0" {
				t.Fatal("concurrent newer install overwritten")
			}
			stages, _ := filepath.Glob(filepath.Join(bin, ".kit-mint-*"))
			if len(stages) != 0 {
				t.Fatal("build started before validating current destination")
			}
		})
	}
}
