package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
	"github.com/spf13/cobra"
)

func TestInitIssueSelection(t *testing.T) {
	for _, tc := range []struct {
		name, input, want string
		invalid           bool
	}{
		{"default", "\n", "github", false}, {"explicit github", "1\n", "github", false},
		{"linear", "2\nacme\n11111111-1111-4111-8111-111111111111\n\n", "linear", false},
		{"unknown", "bogus\n", "", true}, {"missing config", "2\n\n\n\n", "", true}, {"EOF", "", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			got, err := promptIssueProvider(strings.NewReader(tc.input), &out, config.IssueConfig{})
			if (err != nil) != tc.invalid {
				t.Fatalf("error=%v", err)
			}
			if !tc.invalid && got.EffectiveProvider() != tc.want {
				t.Fatal(got)
			}
			if !strings.Contains(out.String(), "GitHub Issues (default)") || !strings.Contains(out.String(), "Linear") {
				t.Fatal(out.String())
			}
		})
	}
}

func TestNoninteractiveIssueFlags(t *testing.T) {
	for _, args := range [][]string{nil, {"--issue-provider", "linear", "--linear-workspace", "acme", "--linear-team-id", "11111111-1111-4111-8111-111111111111"}, {"--issue-provider", "linear"}, {"--linear-workspace", "acme"}, {"--issue-provider", "unknown"}} {
		cmd := &cobra.Command{}
		bindIssueProviderFlags(cmd)
		cmd.SetIn(strings.NewReader(""))
		cmd.SetOut(&bytes.Buffer{})
		if err := cmd.ParseFlags(args); err != nil {
			t.Fatal(err)
		}
		cfg := defaultInitConfig()
		err := selectInitIssueProvider(cmd, cfg)
		invalid := len(args) > 0 && len(args) != 6
		if (err != nil) != invalid {
			t.Fatalf("%v: %v", args, err)
		}
		if !invalid && len(args) == 0 && cfg.Issues.EffectiveProvider() != "github" {
			t.Fatal(cfg.Issues)
		}
	}
}

func TestReconcilePreservesIssueTargetAndSwitch(t *testing.T) {
	root := t.TempDir()
	cfg := defaultInitConfig()
	cfg.Issues = config.IssueConfig{Provider: "linear", Linear: config.LinearConfig{Workspace: "acme", TeamID: "11111111-1111-4111-8111-111111111111"}}
	for _, provider := range []string{"linear", "github", "linear"} {
		cfg.Issues.Provider = provider
		if err := config.Save(root, cfg); err != nil {
			t.Fatal(err)
		}
		for _, force := range []bool{false, true} {
			got, _, err := initRefreshConfig(root, initRefreshOptions{force: force}, map[string]bool{config.ConfigFileName: true})
			if err != nil || got.Issues != cfg.Issues {
				t.Fatalf("force=%v: %#v %v", force, got, err)
			}
		}
	}
}

func TestLinearReadmeKeepsPRsWithoutGitHubIssueBadge(t *testing.T) {
	block := managedReadmeBadgeBlock("acme/app", "", "PUBLIC", "linear")
	if strings.Contains(block, "img.shields.io/github/issues/") || !strings.Contains(block, "issues-pr/") {
		t.Fatal(block)
	}
	legacy := managedReadmeBadgeBlock("acme/app", "", "PUBLIC")
	if !strings.Contains(legacy, "img.shields.io/github/issues/") {
		t.Fatal(legacy)
	}
}

func TestLinearInitPersistsConfigurationAndTemplate(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	withInitFlags(t, func() {
		initOutputOnly = true
		cmd := &cobra.Command{}
		bindIssueProviderFlags(cmd)
		cmd.SetIn(strings.NewReader(""))
		cmd.SetOut(&bytes.Buffer{})
		args := []string{"--issue-provider", "linear", "--linear-workspace", "acme", "--linear-team-id", "11111111-1111-4111-8111-111111111111"}
		if err := cmd.ParseFlags(args); err != nil {
			t.Fatal(err)
		}
		_ = captureStdout(t, func() {
			if err := runInit(cmd, nil); err != nil {
				t.Fatal(err)
			}
		})
		cfg, err := config.Load(root)
		if err != nil || cfg.Issues.EffectiveProvider() != "linear" || cfg.Issues.Linear.Workspace != "acme" {
			t.Fatalf("%#v %v", cfg, err)
		}
		template := readFile(t, root+"/"+pullRequestTemplatePath)
		if strings.Contains(template, "Closes #") || !strings.Contains(template, "[Linear issue URL]") {
			t.Fatal(template)
		}
	})
}

func TestReconcileSwitchesOnlyOwnedIssueTemplate(t *testing.T) {
	root := t.TempDir()
	github := templates.PullRequestTemplate
	linear := templates.BuildPullRequestTemplate("linear")
	for _, tc := range []struct {
		name, before, after string
		want                instructionFileWriteResult
	}{
		{"github to linear", github, linear, instructionFileUpdated},
		{"linear to github", linear, github, instructionFileUpdated},
		{"custom", github + "\nCustom requirements\n", linear, instructionFileSkipped},
		{"idempotent", linear, linear, instructionFileSkipped},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writeFile(t, root+"/"+pullRequestTemplatePath, tc.before)
			change, err := planRefreshInitScaffoldFile(root, initRefreshOptions{}, nil, pullRequestTemplatePath, tc.after, false, false)
			if err != nil || change.result != tc.want {
				t.Fatalf("%#v %v", change, err)
			}
			want := tc.after
			if tc.want == instructionFileSkipped {
				want = tc.before
			}
			if change.after != want {
				t.Fatalf("after=%q, want %q", change.after, want)
			}
		})
	}
}
