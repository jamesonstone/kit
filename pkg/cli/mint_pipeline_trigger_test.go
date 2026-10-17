package cli

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMintPushTriggerFormsAndDefaultBranch(t *testing.T) {
	root, _ := mintRemoteFixture(t)
	mintPreparationGit(t, root, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/trunk")
	for _, tc := range []struct {
		name, on string
		valid    bool
	}{
		{"scalar", "push", true}, {"sequence", "[pull_request, push]", true},
		{"mapping", "\n  push:", true}, {"exact default", "\n  push:\n    branches: [trunk]", true},
		{"multiple exact branches", "\n  push:\n    branches: [release, trunk]", true},
		{"substring event", "[pull_request, push_extra]", false},
		{"missing default", "\n  push:\n    branches: [release]", false},
		{"substring branch", "\n  push:\n    branches: [trunk-release]", false},
		{"wildcard only", "\n  push:\n    branches: ['**']", false},
		{"negative exclusion", "\n  push:\n    branches: [trunk, '!trunk']", false},
		{"ignored branch", "\n  push:\n    branches-ignore: [trunk]", false},
		{"tags only", "\n  push:\n    tags: ['v*']", false},
		{"tags ignored", "\n  push:\n    tags-ignore: ['v*']", false},
		{"reusable", "workflow_call", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var doc yaml.Node
			if err := yaml.Unmarshal([]byte("on: "+tc.on+"\njobs: {}\n"), &doc); err != nil {
				t.Fatal(err)
			}
			err := validateMintPushTrigger(root, &doc)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v error=%v", tc.valid, err)
			}
		})
	}
	t.Run("unknown default", func(t *testing.T) {
		var doc yaml.Node
		if err := yaml.Unmarshal([]byte("on:\n  push:\n    branches: [main]\n"), &doc); err != nil {
			t.Fatal(err)
		}
		if err := validateMintPushTrigger(t.TempDir(), &doc); err == nil || !strings.Contains(err.Error(), "INCOMPLETE") {
			t.Fatalf("unknown metadata accepted: %v", err)
		}
	})
}
