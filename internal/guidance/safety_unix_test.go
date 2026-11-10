//go:build unix

package guidance

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestInspectRejectsNonDirectoryScanTargets(t *testing.T) {
	for _, pipe := range []bool{false, true} {
		root := project(t)
		if err := os.MkdirAll(filepath.Join(root, "docs"), 0755); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(root, "docs/sops")
		if pipe {
			if err := syscall.Mkfifo(path, 0600); err != nil {
				t.Fatal(err)
			}
		} else {
			write(t, root, "docs/sops", "not a directory")
		}
		report, err := Inspect(root)
		if err == nil {
			t.Fatal("non-directory scan target accepted")
		}
		if !strings.Contains(strings.Join(report.Errors, "\n"), "expected a directory") {
			t.Fatalf("%+v", report.Errors)
		}
	}
}
