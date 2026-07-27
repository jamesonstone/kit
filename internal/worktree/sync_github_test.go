package worktree

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestResolveSyncPullRequestsUsesBatchAndExactFallback(t *testing.T) {
	app := NewApp(io.Discard, io.Discard)
	var calls []string
	app.run = func(
		_ context.Context,
		_ string,
		name string,
		args ...string,
	) ([]byte, error) {
		if name != "gh" {
			return nil, fmt.Errorf("unexpected command %s", name)
		}
		head := ""
		for i := range args {
			if args[i] == "--head" && i+1 < len(args) {
				head = args[i+1]
			}
		}
		calls = append(calls, head)
		var prs []SyncPullRequest
		switch head {
		case "":
			prs = []SyncPullRequest{
				mergedSyncPR(80, "topic/batch", "main", strings.Repeat("1", 40)),
			}
		case "topic/batch":
			prs = []SyncPullRequest{
				mergedSyncPR(80, "topic/batch", "main", strings.Repeat("1", 40)),
				mergedSyncPR(82, "other-head", "main", strings.Repeat("3", 40)),
			}
		case "topic/fallback":
			prs = []SyncPullRequest{
				mergedSyncPR(81, "topic/fallback", "main", strings.Repeat("2", 40)),
				mergedSyncPR(83, "other-head", "main", strings.Repeat("4", 40)),
			}
		}
		return json.Marshal(prs)
	}

	result, err := app.resolveSyncPullRequests(
		context.Background(),
		t.TempDir(),
		"example/project",
		[]string{"topic/batch", "topic/fallback"},
	)
	if err != nil {
		t.Fatalf("resolveSyncPullRequests() error = %v", err)
	}
	if len(result["topic/batch"]) != 1 ||
		len(result["topic/fallback"]) != 1 ||
		!reflect.DeepEqual(calls, []string{"", "topic/batch", "topic/fallback"}) {
		t.Fatalf("result=%#v calls=%#v", result, calls)
	}
}
