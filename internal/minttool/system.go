package minttool

import (
	"context"
	"debug/buildinfo"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
)

type systemBackend struct {
	verify func(string, string) error // Test seam; production always uses Go build metadata.
}

func (b systemBackend) verifyFile(path, version string) error {
	if b.verify != nil {
		return b.verify(path, version)
	}
	return verifyBinary(path, version)
}

func (systemBackend) inspect(ctx context.Context) (Installation, error) {
	p, err := exec.LookPath("mint")
	if errors.Is(err, exec.ErrNotFound) {
		return Installation{Status: "missing", Method: "none"}, nil
	}
	if err != nil {
		return Installation{}, errors.New("cannot resolve Mint executable")
	}
	p, err = filepath.Abs(p)
	if err != nil {
		return Installation{}, errors.New("cannot resolve Mint path")
	}
	i := Installation{Path: p, Status: "unsupported", Method: "unknown"}
	info, err := buildinfo.ReadFile(p)
	if err != nil {
		i.Reason = "executable has no readable Go build metadata"
		return i, nil
	}
	i.Version = info.Main.Version
	if err = verifyBinary(p, ""); err != nil {
		i.Reason = err.Error()
		return i, nil
	}
	bin, err := goBin(ctx)
	if err != nil {
		i.Reason = err.Error()
		return i, nil
	}
	if p != filepath.Join(bin, executableName()) {
		i.Reason = "resolved executable is outside the current Go bin directory"
		return i, nil
	}
	if err = safeFile(p); err != nil {
		i.Reason = err.Error()
		return i, nil
	}
	i.Status, i.Method = "installed", "go-install"
	return i, nil
}

func executableName() string {
	if runtime.GOOS == "windows" {
		return "mint.exe"
	}
	return "mint"
}

func verifyBinary(path, version string) error {
	b, err := buildinfo.ReadFile(path)
	if err != nil {
		return errors.New("mint build metadata is unavailable")
	}
	return verifyBuildInfo(b, version)
}

func verifyBuildInfo(b *debug.BuildInfo, version string) error {
	if b.Path != Command || b.Main.Path != Module || b.Main.Replace != nil {
		return errors.New("executable is not an official Mint Go module build")
	}
	if version != "" && b.Main.Version != version {
		return errors.New("built Mint version does not match requested release")
	}
	settings := map[string]string{}
	for _, s := range b.Settings {
		settings[s.Key] = s.Value
	}
	if settings["GOOS"] != runtime.GOOS || settings["GOARCH"] != runtime.GOARCH {
		return errors.New("mint executable platform does not match this host")
	}
	return nil
}

func goBin(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "env", "-json", "GOBIN", "GOPATH")
	cmd.Env = environment(map[string]string{"GOWORK": "off", "GOFLAGS": "", "GOTOOLCHAIN": "local"})
	out, err := cmd.Output()
	if err != nil {
		return "", errors.New("go is unavailable; install/update requires a supported Go installation")
	}
	var env map[string]string
	if json.Unmarshal(out, &env) != nil {
		return "", errors.New("cannot read Go installation directory")
	}
	bin := env["GOBIN"]
	if bin == "" {
		paths := filepath.SplitList(env["GOPATH"])
		if len(paths) != 0 && paths[0] != "" {
			bin = filepath.Join(paths[0], "bin")
		}
	}
	if !filepath.IsAbs(bin) {
		return "", errors.New("go bin directory must be absolute")
	}
	if err := safeDirectory(bin); err != nil {
		return "", err
	}
	return filepath.Clean(bin), nil
}

func owned(info os.FileInfo) bool {
	if os.Getuid() < 0 {
		return false
	}
	uid, ok := statID(info, "Uid")
	return ok && uid == uint64(os.Getuid())
}

func statID(info os.FileInfo, name string) (uint64, bool) {
	v := reflect.ValueOf(info.Sys())
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return 0, false
	}
	u := v.Elem().FieldByName(name)
	if !u.IsValid() || !u.CanUint() {
		return 0, false
	}
	return u.Uint(), true
}

func safeDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil || !info.IsDir() || !owned(info) || info.Mode().Perm()&0022 != 0 {
		return errors.New("go bin must be an existing current-user-owned directory without group/world writes")
	}
	real, err := filepath.EvalSymlinks(path)
	if err != nil || real != filepath.Clean(path) {
		return errors.New("go bin symlink paths are unsupported")
	}
	return nil
}

func safeFile(path string) error {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || !owned(info) || info.Mode().Perm()&0022 != 0 {
		return errors.New("mint must be a regular current-user-owned file without group/world writes; symlinks are unsupported")
	}
	return nil
}

func environment(overrides map[string]string) []string {
	env := []string{}
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		if _, overridden := overrides[key]; !overridden {
			env = append(env, item)
		}
	}
	for k, v := range overrides {
		env = append(env, k+"="+v)
	}
	return env
}
