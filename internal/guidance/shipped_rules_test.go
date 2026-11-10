package guidance

import (
	"path/filepath"
	"testing"

	kit "github.com/jamesonstone/kit/v3"
)

// Exercise the complete embedded corpus, including optional rules that this
// repository does not install in its own .kit.yaml registry.
func TestInspectAllShippedRules(t *testing.T) {
	root := project(t)
	files, err := kit.Rules.ReadDir(kit.RulesDir)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".md" {
			continue
		}
		path := kit.RulesDir + "/" + file.Name()
		data, err := kit.Rules.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		write(t, root, path, string(data))
		count++
	}
	if count == 0 {
		t.Fatal("embedded rules corpus is empty")
	}
	report, err := Inspect(root)
	if err != nil {
		t.Fatalf("shipped corpus rejected: %v; %v", err, report.Errors)
	}
	found := 0
	optional := 0
	for _, entry := range report.Entries {
		if entry.Kind != "ruleset" {
			continue
		}
		found++
		if entry.Metadata == nil || entry.Error != "" {
			t.Fatalf("invalid shipped entry %+v", entry)
		}
		if entry.Metadata.RegistryScope == "optional" {
			optional++
		}
	}
	if found != count {
		t.Fatalf("got %d rules, want %d", found, count)
	}
	if optional == 0 {
		t.Fatal("fixture did not exercise shipped optional rules")
	}
}
