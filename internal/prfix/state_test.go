package prfix

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStateStoreDeduplicatesWatchersAndPersistsOnlyFingerprints(t *testing.T) {
	store := StateStore{root: t.TempDir()}
	target := Target{Repository{"Acme", "App"}, 9}
	release, err := store.Acquire(target, "ABC")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Acquire(target, "ABC"); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("duplicate watcher error = %v", err)
	}
	release()
	secondRelease, err := store.Acquire(target, "ABC")
	if err != nil {
		t.Fatalf("released watcher was not reusable: %v", err)
	}
	secondRelease()

	result := AwaitResult{ExpectedHead: "ABC", ObservedHead: "ABC", State: "completed", Reason: "secret status body"}
	feedback := []Feedback{{Fingerprint: "sha256:one", Body: "secret review body", Task: "secret task"}}
	if err := store.Save(target, "ABC", result, feedback); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(store.root, watcherKey(target, "ABC")+".json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"secret status body", "secret review body", "secret task"} {
		if strings.Contains(string(content), forbidden) {
			t.Errorf("persisted state contains %q", forbidden)
		}
	}
	var persisted PersistedState
	if err := json.Unmarshal(content, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.StatusFingerprint == "" || len(persisted.FeedbackFingerprints) != 1 ||
		persisted.FeedbackFingerprints[0] != "sha256:one" {
		t.Fatalf("persisted state = %#v", persisted)
	}
}
