package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePlugin(t *testing.T, root, id, manifest, instructions string) {
	t.Helper()
	dir := filepath.Join(root, RootDir, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ManifestName), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if instructions != "" {
		if err := os.WriteFile(filepath.Join(dir, "instructions.md"), []byte(instructions), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoadEnabledPassesConfigPathWithoutReadingIt(t *testing.T) {
	root := t.TempDir()
	manifest := "schema_version: 1\nid: layer2context\ninstructions: instructions.md\nconfig: config.yaml\nname: ignored\n"
	writePlugin(t, root, "layer2context", manifest, "# Roster\n\nRead the config file.\n")
	blocks, err := LoadEnabled(root, []string{"layer2context"})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || blocks[0].ConfigPath != ".kit/plugins/layer2context/config.yaml" || blocks[0].Body != "# Roster\n\nRead the config file." {
		t.Fatalf("%#v", blocks)
	}
	if _, err := os.Stat(filepath.Join(root, RootDir, "layer2context", "config.yaml")); !os.IsNotExist(err) {
		t.Fatal("config file was created")
	}
	none, err := LoadEnabled(root, nil)
	if err != nil || none != nil {
		t.Fatalf("empty load %#v %v", none, err)
	}
}

func TestLoadEnabledRejectsUnsafePlugins(t *testing.T) {
	root := t.TempDir()
	writePlugin(t, root, "fixture", "schema_version: 1\nid: fixture\ninstructions: instructions.md\n", "ok\n")
	cases := []struct {
		id       string
		manifest string
		body     string
	}{
		{id: "missing"},
		{id: "fixture", manifest: "schema_version: 2\nid: fixture\ninstructions: instructions.md\n", body: "ok\n"},
		{id: "fixture", manifest: "schema_version: 1\nid: other\ninstructions: instructions.md\n", body: "ok\n"},
		{id: "fixture", manifest: "schema_version: 1\nid: fixture\ninstructions: ../secret.md\n", body: "ok\n"},
		{id: "fixture", manifest: "schema_version: 1\nid: fixture\ninstructions: instructions.md\n", body: "<!-- BEGIN KIT-PLUGIN fixture -->\n"},
	}
	for _, tc := range cases {
		if tc.manifest != "" {
			writePlugin(t, root, "fixture", tc.manifest, tc.body)
		}
		if _, err := LoadEnabled(root, []string{tc.id}); err == nil {
			t.Fatalf("accepted %#v", tc)
		}
	}
	if _, err := LoadEnabled(root, []string{"Bad ID"}); err == nil {
		t.Fatal("accepted a bad id")
	}
	writePlugin(t, root, "fixture", "schema_version: 1\nid: fixture\ninstructions: instructions.md\n", "ok\n")
	if _, err := LoadEnabled(root, []string{"fixture", "fixture"}); err == nil || !strings.Contains(err.Error(), "more than once") {
		t.Fatal(err)
	}
}

func TestLoadEnabledRejectsMarkerConfigPath(t *testing.T) {
	root := t.TempDir()
	writePlugin(t, root, "fixture", "schema_version: 1\nid: fixture\ninstructions: instructions.md\nconfig: KIT-PLUGIN.yaml\n", "ok\n")
	if _, err := LoadEnabled(root, []string{"fixture"}); err == nil || !strings.Contains(err.Error(), "config path") {
		t.Fatal(err)
	}
}

func TestLoadEnabledRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, RootDir, "fixture")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "secret.md")
	if err := os.WriteFile(outside, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "instructions.md")); err != nil {
		t.Fatal(err)
	}
	manifest := "schema_version: 1\nid: fixture\ninstructions: instructions.md\n"
	if err := os.WriteFile(filepath.Join(dir, ManifestName), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadEnabled(root, []string{"fixture"}); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatal(err)
	}
}

func TestLoadEnabledRejectsOversizedInstructions(t *testing.T) {
	root := t.TempDir()
	writePlugin(t, root, "fixture", "schema_version: 1\nid: fixture\ninstructions: instructions.md\n", strings.Repeat("a", maxInstructionBytes+1))
	if _, err := LoadEnabled(root, []string{"fixture"}); err == nil {
		t.Fatal("accepted oversized instructions")
	}
}
