package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/feature"
)

func TestBuildSkillMinePrompt(t *testing.T) {
	projectRoot := t.TempDir()
	featurePath := filepath.Join(projectRoot, "docs", "specs", "0006-skill-mine-command")
	brainstormPath := filepath.Join(featurePath, "BRAINSTORM.md")
	specPath := filepath.Join(featurePath, "SPEC.md")
	planPath := filepath.Join(featurePath, "PLAN.md")
	tasksPath := filepath.Join(featurePath, "TASKS.md")
	skillsDir := filepath.Join(projectRoot, ".agents", "skills")
	canonicalSkillPath := filepath.Join(skillsDir, "skill-mine-command", "SKILL.md")
	claudeMirrorPath := filepath.Join(projectRoot, ".claude", "skills", "skill-mine-command", "SKILL.md")

	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(brainstormPath, []byte("# BRAINSTORM\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	feat := &feature.Feature{
		Slug: "skill-mine-command",
		Path: featurePath,
	}

	prompt := buildSkillMinePrompt(
		feat,
		brainstormPath,
		specPath,
		planPath,
		tasksPath,
		skillsDir,
		projectRoot,
	)

	checks := []string{
		skillsDir,
		feat.Slug,
		canonicalSkillPath,
		claudeMirrorPath,
		"PROJECT_PROGRESS_SUMMARY.md",
		"CONSTITUTION.md",
		"git diff main",
		"git diff master",
		"Read all existing canonical skill bundles",
		"spec-vs-implementation divergence",
		"description: <one sentence: when to trigger this skill>",
		"must describe when the skill should trigger, not what it does",
		"<skill-name>/",
		"SKILL.md",
		"scripts/        # optional",
		"references/     # optional",
		"assets/         # optional",
		"<procedural knowledge - what to do, in what order, with what constraints>",
		"Duplicate the full skill directory into the Claude mirror",
		"## SKILL AUDIT",
		"ACCURACY",
		"RELEVANCE",
		"COVERAGE",
		"TRIGGER CONDITION",
		"rm -rf " + skillsDir + "/<skill-name>/",
		"rm -rf .claude/skills/<skill-name>/",
		"## Skill Audit Summary",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}

	lower := strings.ToLower(prompt)
	for _, forbidden := range []string{"http://", "https://", "api call"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("expected prompt not to contain %q", forbidden)
		}
	}

	if strings.Contains(prompt, filepath.Join(skillsDir, "skill-mine-command.SKILL.md")) {
		t.Fatalf("expected prompt not to use flat .SKILL.md path")
	}

	if !strings.HasPrefix(prompt, "Mine a reusable skill for feature: skill-mine-command\n\n") {
		t.Fatalf("expected prompt to start with a clear task statement, got %q", prompt[:48])
	}
}
