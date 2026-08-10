package prfix

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type StateStore struct{ root string }

type PersistedState struct {
	SchemaVersion        int      `json:"schema_version"`
	Repository           string   `json:"repository"`
	PullRequest          int      `json:"pull_request"`
	HeadSHA              string   `json:"head_sha"`
	StatusFingerprint    string   `json:"status_fingerprint,omitempty"`
	FeedbackFingerprints []string `json:"feedback_fingerprints,omitempty"`
	UpdatedAt            string   `json:"updated_at"`
}

func NewStateStore() (StateStore, error) {
	root := strings.TrimSpace(os.Getenv("KIT_STATE_HOME"))
	if root == "" {
		root = strings.TrimSpace(os.Getenv("XDG_STATE_HOME"))
	}
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return StateStore{}, fmt.Errorf("resolve user home for PR feedback state: %w", err)
		}
		root = filepath.Join(home, ".local", "state")
	}
	return StateStore{root: filepath.Join(root, "kit", "pr-feedback")}, nil
}

func (store StateStore) Acquire(target Target, head string) (func(), error) {
	if err := os.MkdirAll(store.root, 0o700); err != nil {
		return nil, fmt.Errorf("create PR feedback state directory: %w", err)
	}
	path := filepath.Join(store.root, watcherKey(target, head)+".lock")
	// #nosec G304 -- the filename is a SHA-256 watcher key under the selected state root.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if os.IsExist(err) {
		return nil, fmt.Errorf("an await watcher already exists for %s#%d at head %s", target.Slug(), target.Number, head)
	}
	if err != nil {
		return nil, fmt.Errorf("acquire PR feedback watcher: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return nil, err
	}
	return func() { _ = os.Remove(path) }, nil
}

func (store StateStore) Save(target Target, head string, result AwaitResult, feedback []Feedback) error {
	if err := os.MkdirAll(store.root, 0o700); err != nil {
		return fmt.Errorf("create PR feedback state directory: %w", err)
	}
	fingerprints := make([]string, 0, len(feedback))
	for _, item := range feedback {
		fingerprints = append(fingerprints, item.Fingerprint)
	}
	state := PersistedState{
		SchemaVersion: 1, Repository: target.Slug(), PullRequest: target.Number, HeadSHA: head,
		StatusFingerprint: statusFingerprint(result), FeedbackFingerprints: fingerprints,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	path := filepath.Join(store.root, watcherKey(target, head)+".json")
	temporary, err := os.CreateTemp(store.root, ".state-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func watcherKey(target Target, head string) string {
	value := fmt.Sprintf("%s#%d@%s", strings.ToLower(target.Slug()), target.Number, head)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value)))
}

func statusFingerprint(result AwaitResult) string {
	value := strings.Join([]string{result.ExpectedHead, result.ObservedHead,
		result.ProviderState, result.ProviderDescription, result.State}, "\x00")
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(value)))
}
