package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckedInProjectGuidanceMatchesDefaults(t *testing.T) {
	for _, path := range []string{DesignPath, SecurityPath} {
		want, ok := ProjectGuidance(path)
		if !ok || want == "" {
			t.Fatalf("missing canonical guidance for %s", path)
		}
		got, err := os.ReadFile(filepath.Join("..", "..", path))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("checked-in %s drifted from its embedded default", path)
		}
	}
}
