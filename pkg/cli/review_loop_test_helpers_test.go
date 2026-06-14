package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func withFakeEditor(t *testing.T) {
	t.Helper()
	previousWait := awaitEditorLaunchConfirmation
	previousRunner := editorInputRunner
	t.Cleanup(func() {
		awaitEditorLaunchConfirmation = previousWait
		editorInputRunner = previousRunner
	})
	awaitEditorLaunchConfirmation = func(_ *os.File, _ io.Writer) error {
		return nil
	}
	editorInputRunner = func(_ freeTextInputConfig, _ string, initialContent string) (string, bool, error) {
		return initialContent, true, nil
	}
}

func installReviewLoopFakes(
	t *testing.T,
	clock reviewLoopClockSource,
	runner reviewLoopCommandRunner,
) func() {
	t.Helper()
	previousClock := reviewLoopClock
	previousRunner := reviewLoopRunner
	previousLoader := reviewLoopLoadReviewTasks
	if clock != nil {
		reviewLoopClock = clock
	}
	if runner != nil {
		reviewLoopRunner = runner
	}
	return func() {
		reviewLoopClock = previousClock
		reviewLoopRunner = previousRunner
		reviewLoopLoadReviewTasks = previousLoader
	}
}

type fakeReviewLoopRunner struct {
	output           func(dir string, name string, args ...string) ([]byte, error)
	outputAllowError func(dir string, name string, args ...string) ([]byte, error)
}

func (f fakeReviewLoopRunner) Output(dir string, name string, args ...string) ([]byte, error) {
	if f.output != nil {
		return f.output(dir, name, args...)
	}
	return nil, fmt.Errorf("unexpected command: %s %v", name, args)
}

func (f fakeReviewLoopRunner) OutputAllowError(dir string, name string, args ...string) ([]byte, error) {
	if f.outputAllowError != nil {
		return f.outputAllowError(dir, name, args...)
	}
	return f.Output(dir, name, args...)
}

type fakeReviewLoopClock struct {
	now    time.Time
	sleeps []time.Duration
}

func (c *fakeReviewLoopClock) Now() time.Time {
	return c.now
}

func (c *fakeReviewLoopClock) Sleep(duration time.Duration) {
	c.sleeps = append(c.sleeps, duration)
	c.now = c.now.Add(duration)
}

func reviewLoopPRPayload(sha string) []byte {
	return []byte(fmt.Sprintf(`{
		"number": 67,
		"url": "https://github.com/Patient-Driven-Care/cortex/pull/67",
		"title": "feat(GH-67): review loop",
		"body": "Closes #67",
		"headRefOid": %q
	}`, sha))
}

func containsReviewLoopMutation(args []string) bool {
	joined := strings.ToLower(strings.Join(args, " "))
	mutations := []string{
		" git add ",
		" git commit ",
		" git push ",
		" pr comment ",
		" pr review ",
		" resolve",
	}
	for _, mutation := range mutations {
		if strings.Contains(" "+joined+" ", mutation) {
			return true
		}
	}
	return false
}
