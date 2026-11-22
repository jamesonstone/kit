package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
	"github.com/spf13/cobra"
)

func TestInitIssueSelection(t *testing.T) {
	for _, tc := range []struct {
		name, input, want, prefix string
		invalid                   bool
	}{
		{"default", "\n", "github", "GH-", false}, {"explicit github", "1\n", "github", "GH-", false},
		{"linear", "2\nacme\n11111111-1111-4111-8111-111111111111\n\nLAB2\n", "linear", "LAB2-", false},
		{"linear without key", "2\nacme\n11111111-1111-4111-8111-111111111111\n\n\n", "linear", "", false},
		{"unknown", "bogus\n", "", "", true}, {"missing config", "2\n\n\n\n\n", "", "", true}, {"EOF", "", "", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			got, err := promptIssueProvider(strings.NewReader(tc.input), &out, config.IssueConfig{})
			if (err != nil) != tc.invalid {
				t.Fatalf("error=%v", err)
			}
			if !tc.invalid && (got.EffectiveProvider() != tc.want || got.BranchPrefix() != tc.prefix) {
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

func TestInitDropsPrefilledKeyAfterChangingTeam(t *testing.T) {
	initial := config.IssueConfig{Linear: config.LinearConfig{
		Workspace: "acme", TeamID: "11111111-1111-4111-8111-111111111111", TeamKey: "LAB2",
	}}
	for _, tc := range []struct{ team, prefix string }{
		{"", "LAB2-"},
		{"22222222-2222-4222-8222-222222222222", ""},
	} {
		got, err := promptIssueProvider(strings.NewReader("2\n\n"+tc.team+"\n\n\n"), &bytes.Buffer{}, initial)
		if err != nil || got.BranchPrefix() != tc.prefix {
			t.Fatalf("team %q: %#v %v", tc.team, got, err)
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
		args := []string{"--issue-provider", "linear", "--linear-workspace", "acme", "--linear-team-id", "11111111-1111-4111-8111-111111111111", "--linear-team-key", "LAB2"}
		if err := cmd.ParseFlags(args); err != nil {
			t.Fatal(err)
		}
		_ = captureStdout(t, func() {
			if err := runInit(cmd, nil); err != nil {
				t.Fatal(err)
			}
		})
		cfg, err := config.Load(root)
		if err != nil || cfg.Issues.EffectiveProvider() != "linear" || cfg.Issues.Linear.Workspace != "acme" || cfg.Issues.BranchPrefix() != "LAB2-" {
			t.Fatalf("%#v %v", cfg, err)
		}
		template := readFile(t, root+"/"+pullRequestTemplatePath)
		if strings.Contains(template, "Closes #") || !strings.Contains(template, "[Linear issue URL]") {
			t.Fatal(template)
		}
	})
}

func previousPullRequestTemplate() string {
	return `## Description

-

## How to Test

1.

## Ticket

Closes #[ticket number]
`
}

func previousLinearPullRequestTemplate() string {
	return strings.Replace(previousPullRequestTemplate(), "Closes #[ticket number]", "<!-- Full canonical Linear URL; preserve in the squash commit footer. Follow delivery for lifecycle. -->\n[Linear issue URL]", 1)
}

const priorEvidencePullRequestTemplate = `## Description

-

## How to Test

1.

## Evidence

- Commands: command and real output or excerpt
- Decisions and assumptions: defaults, scope calls, conflicting requirements reinterpreted
- Checks: for each check, the property it verified
- Not checked:
- Fresh-context review:

## Ticket

Closes #[ticket number]
`

func TestReconcileSwitchesOnlyOwnedIssueTemplate(t *testing.T) {
	root := t.TempDir()
	github := templates.PullRequestTemplate
	linear := templates.BuildPullRequestTemplate("linear")
	priorLinear := strings.Replace(priorEvidencePullRequestTemplate, "Closes #[ticket number]", "<!-- Full canonical Linear URL; preserve in the squash commit footer. Follow delivery for lifecycle. -->\n[Linear issue URL]", 1)
	for _, tc := range []struct {
		name, before, after string
		want                instructionFileWriteResult
	}{
		{"github to linear", github, linear, instructionFileUpdated},
		{"linear to github", linear, github, instructionFileUpdated},
		{"previous github", previousPullRequestTemplate(), github, instructionFileUpdated},
		{"previous linear", previousLinearPullRequestTemplate(), linear, instructionFileUpdated},
		{"prior evidence github", priorEvidencePullRequestTemplate, github, instructionFileUpdated},
		{"prior evidence linear", priorLinear, linear, instructionFileUpdated},
		{"custom prior evidence", priorEvidencePullRequestTemplate + "\nCustom review requirements\n", github, instructionFileSkipped},
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

func TestProjectIssueScopeMatchesReconcileTemplate(t *testing.T) {
	setupInitHome(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	globalPath, err := config.GlobalConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	const globalIssues = "issues:\n  provider: linear\n  issue_template: true\n  linear:\n    workspace: global-example\n    team_id: 11111111-1111-4111-8111-111111111111\n"
	anchored := "defaults: &defaults\n  " + strings.ReplaceAll(strings.TrimSuffix(globalIssues, "\n"), "\n", "\n  ") + "\n<<: *defaults\n"
	for _, tc := range []struct {
		name, global, project, provider, workspace string
	}{
		{name: "omitted", global: globalIssues, provider: "github"},
		{name: "global YAML merge", global: anchored, provider: "github"},
		{name: "global tagged key", global: strings.Replace(globalIssues, "issues:", "!!binary aXNzdWVz:", 1), provider: "github"},
		{name: "local Linear", global: globalIssues, project: "issues:\n  provider: linear\n  linear:\n    workspace: project-example\n    team_id: 22222222-2222-4222-8222-222222222222\n", provider: "linear", workspace: "project-example"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writeFile(t, globalPath, tc.global)
			root := t.TempDir()
			writeFile(t, filepath.Join(root, config.ConfigFileName), "schema_version: 2\n"+tc.project)
			cfg, err := config.Load(root)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Issues.EffectiveProvider() != tc.provider || cfg.Issues.Linear.Workspace != tc.workspace || cfg.Issues.IssueTemplateEnabled() {
				t.Fatalf("load issue settings = %+v", cfg.Issues)
			}
			opts := initRefreshOptions{dryRun: true, files: []string{pullRequestTemplatePath, issueTemplatePath}}
			plan, err := buildInitRefreshPlan(t.Context(), root, opts)
			if err != nil {
				t.Fatal(err)
			}
			want := templates.BuildPullRequestTemplate(tc.provider)
			for _, change := range plan.changes {
				switch change.relativePath {
				case pullRequestTemplatePath:
					if change.after != want {
						t.Fatal("PR template disagrees with the repository issue provider")
					}
				case issueTemplatePath:
					if change.result != instructionFileSkipped {
						t.Fatal("inherited a global issue template opt-in")
					}
				}
			}
			if len(plan.changes) != 2 {
				t.Fatalf("expected both template entries in plan, got %d", len(plan.changes))
			}
		})
	}
	t.Run("incomplete local target", func(t *testing.T) {
		writeFile(t, globalPath, globalIssues)
		root := t.TempDir()
		writeFile(t, filepath.Join(root, config.ConfigFileName), "schema_version: 2\nissues:\n  provider: linear\n")
		if _, err := config.Load(root); err == nil || !strings.Contains(err.Error(), "issues.linear.workspace") {
			t.Fatalf("incomplete repository target = %v", err)
		}
	})
}
