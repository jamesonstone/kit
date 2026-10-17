package minttool

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func (systemBackend) validate(ctx context.Context, old Installation) error {
	bin, err := goBin(ctx)
	if err != nil {
		return err
	}
	target := filepath.Join(bin, executableName())
	if old.Status != "missing" {
		if old.Path != target || old.Method != "go-install" {
			return errors.New("mint installation changed; check again")
		}
		return safeFile(target)
	}
	if !binInPath(bin) {
		return errors.New("go bin is absent from PATH; configure PATH explicitly before initial Mint install")
	}
	if _, err = os.Lstat(target); !errors.Is(err, os.ErrNotExist) {
		return errors.New("mint destination already exists; refusing initial install")
	}
	return nil
}

func (b systemBackend) install(ctx context.Context, old Installation, version string) (Installation, error) {
	if !validVersion(version) {
		return old, errors.New("invalid Mint release tag")
	}
	bin, err := goBin(ctx)
	if err != nil {
		return old, err
	}
	target := filepath.Join(bin, executableName())
	var before os.FileInfo
	if old.Status != "missing" {
		if old.Path != target || old.Method != "go-install" {
			return old, errors.New("mint installation changed; check again")
		}
		if err = safeFile(target); err != nil {
			return old, err
		}
		before, err = os.Lstat(target)
		if err != nil {
			return old, errors.New("cannot inspect existing Mint")
		}
		if err = b.verifyFile(target, old.Version); err != nil {
			return old, errors.New("mint version changed since inspection; check again")
		}
		if !sameSnapshot(target, before) {
			return old, errors.New("mint changed during version verification; check again")
		}
	} else if _, err = os.Lstat(target); !errors.Is(err, os.ErrNotExist) {
		return old, errors.New("mint destination already exists; refusing initial install")
	}
	if old.Status == "missing" && !binInPath(bin) {
		return old, errors.New("go bin is absent from PATH; configure PATH explicitly before initial Mint install")
	}
	stage, err := os.MkdirTemp(bin, ".kit-mint-")
	if err != nil {
		return old, errors.New("cannot create Mint staging directory; check Go bin permissions")
	}
	defer func() { _ = os.RemoveAll(stage) }()
	cmd := exec.CommandContext(ctx, "go", "install", Command+"@"+version)
	cmd.Dir = stage
	cmd.Env = environment(map[string]string{
		"GOBIN": stage, "GOENV": "off", "GOWORK": "off", "GOFLAGS": "", "GO111MODULE": "on",
		"GOTOOLCHAIN": "local", "GOPROXY": "https://proxy.golang.org", "GOSUMDB": "sum.golang.org",
		"GONOSUMDB": "", "GONOPROXY": "", "GOPRIVATE": "", "GOINSECURE": "", "GOEXPERIMENT": "",
		"GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH, "CGO_ENABLED": "0",
	})
	// Do not relay build output: Go or credential helpers can include private environment data.
	if err = cmd.Run(); err != nil {
		return old, errors.New("verified Go install failed; existing Mint unchanged (check Go version, proxy access and permissions)")
	}
	candidate := filepath.Join(stage, executableName())
	if err = b.verifyFile(candidate, version); err != nil {
		return old, err
	}
	if err = safeFile(candidate); err != nil {
		return old, err
	}
	if err = safeDirectory(bin); err != nil {
		return old, err
	}
	mode := os.FileMode(0755)
	if before != nil {
		if !sameSnapshot(target, before) {
			return old, errors.New("mint changed during build; refusing replacement")
		}
		if err = safeFile(target); err != nil {
			return old, err
		}
		mode = before.Mode().Perm()
		group, ok := statID(before, "Gid")
		if !ok || uint64(int(group)) != group {
			return old, errors.New("cannot verify existing Mint group ownership")
		}
		if err = os.Chown(candidate, -1, int(group)); err != nil {
			return old, errors.New("cannot preserve Mint group ownership; existing installation unchanged")
		}
	} else if _, err = os.Lstat(target); !errors.Is(err, os.ErrNotExist) {
		return old, errors.New("mint destination appeared during build; refusing replacement")
	}
	if err = os.Chmod(candidate, mode); err != nil {
		return old, errors.New("cannot preserve Mint executable permissions")
	}
	if err = os.Rename(candidate, target); err != nil {
		return old, errors.New("cannot atomically replace Mint; existing installation unchanged")
	}
	return Installation{Path: target, Version: version, Method: "go-install", Status: "installed"}, nil
}

func sameSnapshot(path string, before os.FileInfo) bool {
	after, err := os.Lstat(path)
	return err == nil && os.SameFile(before, after) && before.Size() == after.Size() && before.ModTime().Equal(after.ModTime())
}

func binInPath(bin string) bool {
	for _, path := range filepath.SplitList(os.Getenv("PATH")) {
		if filepath.IsAbs(path) && filepath.Clean(path) == bin {
			return true
		}
	}
	return false
}
