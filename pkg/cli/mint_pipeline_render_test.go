package cli

import (
	stdreflect "reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMintReleasePipelineTrustAndPermissions(t *testing.T) {
	data := renderMintReleasePipeline("main")
	var w struct {
		On map[string]struct {
			Branches []string `yaml:"branches"`
		} `yaml:"on"`
		Permissions map[string]string `yaml:"permissions"`
		Concurrency struct {
			Group  string `yaml:"group"`
			Cancel bool   `yaml:"cancel-in-progress"`
		} `yaml:"concurrency"`
		Jobs map[string]struct {
			If          string            `yaml:"if"`
			Permissions map[string]string `yaml:"permissions"`
			Steps       []struct {
				Uses string         `yaml:"uses"`
				With map[string]any `yaml:"with"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(data, &w); err != nil {
		t.Fatal(err)
	}
	if len(w.On) != 1 || !stdreflect.DeepEqual(w.On["push"].Branches, []string{"main"}) {
		t.Fatalf("unsafe triggers: %+v", w.On)
	}
	if !stdreflect.DeepEqual(w.Permissions, map[string]string{"contents": "read"}) {
		t.Fatalf("top permissions: %+v", w.Permissions)
	}
	if len(w.Jobs) != 1 {
		t.Fatal("expected exactly one publisher")
	}
	j := w.Jobs["publish"]
	if !stdreflect.DeepEqual(j.Permissions, map[string]string{"contents": "write"}) {
		t.Fatalf("publisher permissions: %+v", j.Permissions)
	}
	for _, guard := range []string{"github.event_name == 'push'", "github.event.repository.default_branch", "github.event.deleted == false"} {
		if !strings.Contains(j.If, guard) {
			t.Fatalf("missing trust guard %s", guard)
		}
	}
	if w.Concurrency.Group != "mint-release-${{ github.repository }}" || w.Concurrency.Cancel {
		t.Fatal("release publisher concurrency is unsafe")
	}
	if len(j.Steps) != 5 || j.Steps[0].Uses != "actions/checkout@"+mintPipelineCheckoutRef || j.Steps[0].With["fetch-depth"] != 0 || j.Steps[0].With["ref"] != "${{ github.sha }}" {
		t.Fatalf("checkout: %+v", j.Steps)
	}
	for i, want := range []string{"release-resolve", "changelog", "", "release-publish"} {
		if got := j.Steps[i+1].With["command"]; (want != "" && got != want) || got == "version" {
			t.Fatalf("step %d command=%v", i+1, got)
		}
	}
	m := j.Steps[4]
	if m.Uses != "jamesonstone/mint@"+mintPipelineActionRef {
		t.Fatalf("unpinned Mint: %s", m.Uses)
	}
	for key, want := range map[string]any{"command": "release-publish", "commitish": "${{ github.sha }}", "owner": "${{ github.repository_owner }}", "repo": "${{ github.event.repository.name }}", "github-token": "${{ github.token }}"} {
		if m.With[key] != want {
			t.Errorf("%s=%v, want %v", key, m.With[key], want)
		}
	}
	if strings.Contains(string(data), "workflow_dispatch") || strings.Contains(string(data), "pull_request") {
		t.Fatal("unsafe write trigger")
	}
}

func TestMintReleasePipelineBranchEncoding(t *testing.T) {
	for _, branch := range []string{"main", "release/stable", "true", "yes", "feature/'quoted'", "rélease"} {
		var w struct {
			On struct {
				Push struct {
					Branches []string `yaml:"branches"`
				} `yaml:"push"`
			} `yaml:"on"`
		}
		if err := yaml.Unmarshal(renderMintReleasePipeline(branch), &w); err != nil {
			t.Fatal(err)
		}
		if !stdreflect.DeepEqual(w.On.Push.Branches, []string{branch}) {
			t.Fatalf("branch not encoded safely: %q", branch)
		}
	}
}
