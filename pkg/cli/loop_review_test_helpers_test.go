package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func setupLoopReviewProject(t *testing.T, agentScript string) string {
	t.Helper()
	projectRoot := t.TempDir()
	agentPath := filepath.Join(projectRoot, "agent.sh")
	if err := os.WriteFile(agentPath, []byte(agentScript), 0o755); err != nil {
		t.Fatalf("WriteFile(agent) error = %v", err)
	}
	cfg := config.Default()
	cfg.Loop.Agent = config.LoopAgentConfig{Command: agentPath}
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	return projectRoot
}

func installLoopReviewGitFake(t *testing.T) func() {
	t.Helper()
	return installReviewLoopFakes(t, nil, fakeReviewLoopRunner{
		output: func(_ string, name string, args ...string) ([]byte, error) {
			if name != "git" {
				return nil, fmt.Errorf("unexpected command: %s %v", name, args)
			}
			return installLoopReviewGitOutput(args...)
		},
	})
}

func installLoopReviewPRFake(t *testing.T, status reviewLoopCheckStatus) func() {
	t.Helper()
	return installReviewLoopFakes(t, nil, fakeReviewLoopRunner{
		output: func(_ string, name string, args ...string) ([]byte, error) {
			if name == "gh" {
				return reviewLoopPRPayload("abc123"), nil
			}
			if name != "git" {
				return nil, fmt.Errorf("unexpected command: %s %v", name, args)
			}
			return installLoopReviewGitOutput(args...)
		},
		outputAllowError: func(_ string, name string, args ...string) ([]byte, error) {
			if name != "gh" {
				return nil, fmt.Errorf("unexpected command: %s %v", name, args)
			}
			switch status {
			case reviewLoopCheckPending:
				return []byte(`[{"name":"CodeRabbit","state":"PENDING","bucket":"pending"}]`), nil
			case reviewLoopCheckComplete:
				return []byte(`[{"name":"CodeRabbit","state":"SUCCESS","bucket":"pass"}]`), nil
			default:
				return []byte(`[]`), nil
			}
		},
	})
}

func installLoopReviewGitOutput(args ...string) ([]byte, error) {
	joined := strings.Join(args, " ")
	switch joined {
	case "rev-parse --verify origin/main":
		return []byte("origin/main\n"), nil
	case "diff --name-only origin/main...HEAD":
		return []byte("internal/app.go\n"), nil
	case "diff --name-only", "diff --cached --name-only":
		return nil, nil
	case "diff --stat origin/main...HEAD":
		return []byte(" internal/app.go | 2 +-\n"), nil
	case "diff --stat", "diff --cached --stat":
		return nil, nil
	default:
		return nil, fmt.Errorf("unexpected git args: %s", joined)
	}
}
