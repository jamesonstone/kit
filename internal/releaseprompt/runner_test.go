package releaseprompt

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type fakeRepositoryState struct {
	Top           string
	Remote        string
	DefaultBranch string
}

type fakeRunner struct {
	repositories map[string]fakeRepositoryState
	lookPaths    map[string]bool
	ghMetadata   ghRepository
	ghCalls      int
}

func (f *fakeRunner) Run(_ context.Context, dir, name string, args ...string) ([]byte, error) {
	dir = filepath.Clean(dir)
	if name == "gh" {
		f.ghCalls++
		if f.ghMetadata.NameWithOwner == "" {
			return nil, fmt.Errorf("gh unavailable")
		}
		return json.Marshal(f.ghMetadata)
	}
	if name != "git" {
		return nil, fmt.Errorf("unexpected command %s", name)
	}
	state, ok := f.repositories[dir]
	if !ok {
		return nil, fmt.Errorf("not a repository")
	}
	command := strings.Join(args, " ")
	switch command {
	case "rev-parse --show-toplevel":
		if state.Top == "" {
			state.Top = dir
		}
		return []byte(state.Top + "\n"), nil
	case "remote get-url origin":
		if state.Remote == "" {
			return nil, fmt.Errorf("remote unavailable")
		}
		return []byte(state.Remote + "\n"), nil
	case "symbolic-ref --short refs/remotes/origin/HEAD":
		if state.DefaultBranch == "" {
			return nil, fmt.Errorf("default branch unavailable")
		}
		return []byte("origin/" + state.DefaultBranch + "\n"), nil
	default:
		return nil, fmt.Errorf("unexpected git command %s", command)
	}
}

func (f *fakeRunner) LookPath(name string) (string, error) {
	if f.lookPaths[name] {
		return "/usr/bin/" + name, nil
	}
	return "", fmt.Errorf("not found")
}

func addFakeRepository(f *fakeRunner, path, remote, branch string) {
	if f.repositories == nil {
		f.repositories = map[string]fakeRepositoryState{}
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		resolved = path
	}
	resolved = filepath.Clean(resolved)
	f.repositories[resolved] = fakeRepositoryState{Top: resolved, Remote: remote, DefaultBranch: branch}
}
