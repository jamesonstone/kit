package minttool

import (
	"context"
	"errors"
	"testing"
)

type lifecycleBackend struct {
	fakeBackend
	releases, inspections, validations int
	disappearAt                        int
	releaseError                       error
}

func (b *lifecycleBackend) inspect(context.Context) (Installation, error) {
	b.inspections++
	if b.disappearAt > 0 && b.inspections >= b.disappearAt {
		return Installation{Status: "missing", Method: "none"}, nil
	}
	return b.installed, nil
}

func (b *lifecycleBackend) release(ctx context.Context, version string) (string, error) {
	b.releases++
	if b.releaseError != nil {
		return "", b.releaseError
	}
	return b.fakeBackend.release(ctx, version)
}

func (b *lifecycleBackend) validate(context.Context, Installation) error {
	b.validations++
	return nil
}

func (b *lifecycleBackend) install(ctx context.Context, old Installation, version string) (Installation, error) {
	i, err := b.fakeBackend.install(ctx, old, version)
	if err == nil {
		i.Path = old.Path
		b.installed = i
	}
	return i, err
}

func TestMaintainInstalledLifecycle(t *testing.T) {
	for _, tt := range []struct {
		name, version, method, status, latest string
		preview, wantErr                      bool
		installs, releases                    int
		wantStatus                            string
	}{
		{"old", "v0.8.0", "go-install", "installed", "v0.9.0", false, false, 1, 1, "current"},
		{"current", "v0.9.0", "go-install", "installed", "v0.9.0", false, false, 0, 1, "current"},
		{"newer", "v0.10.0", "go-install", "installed", "v0.9.0", false, false, 0, 1, "newer"},
		{"missing", "", "none", "missing", "", false, false, 0, 0, "missing"},
		{"unsupported", "v0.8.0", "unknown", "unsupported", "", false, true, 0, 0, "unsupported"},
		{"development", "(devel)", "go-install", "installed", "", false, true, 0, 0, "development_or_unversioned"},
		{"pseudo", "v0.9.1-0.20261020000000-abcdef", "go-install", "installed", "", false, true, 0, 0, "development_or_unversioned"},
		{"old preview", "v0.8.0", "go-install", "installed", "v0.9.0", true, false, 0, 1, "outdated"},
		{"missing preview", "", "none", "missing", "", true, false, 0, 0, "missing"},
		{"below minimum", "v0.7.0", "go-install", "installed", "v0.8.0", false, true, 0, 1, "outdated"},
		{"newer below minimum", "v0.8.0", "go-install", "installed", "v0.7.0", false, true, 0, 1, "newer"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			b := &lifecycleBackend{fakeBackend: fakeBackend{installed: Installation{Path: "/fixture/mint", Version: tt.version, Method: tt.method, Status: tt.status}, latest: tt.latest}}
			r, err := maintainInstalled(context.Background(), tt.preview, b)
			if (err != nil) != tt.wantErr || b.calls != tt.installs || b.releases != tt.releases || r.VersionStatus != tt.wantStatus {
				t.Fatalf("report=%+v error=%v installs=%d releases=%d", r, err, b.calls, b.releases)
			}
			if tt.preview && r.Updated {
				t.Fatal("preview updated")
			}
			if tt.name == "old preview" && (!r.UpdateAvailable || b.validations != 1) {
				t.Fatal("preview did not validate update")
			}
		})
	}
}

func TestMaintainInstalledRepeatabilityAndFailures(t *testing.T) {
	newBackend := func() *lifecycleBackend {
		return &lifecycleBackend{fakeBackend: fakeBackend{installed: Installation{Path: "/fixture/mint", Version: "v0.8.0", Method: "go-install", Status: "installed"}, latest: "v0.9.0"}}
	}
	b := newBackend()
	if _, err := maintainInstalled(context.Background(), false, b); err != nil {
		t.Fatal(err)
	}
	r, err := maintainInstalled(context.Background(), false, b)
	if err != nil || r.Updated || b.calls != 1 {
		t.Fatalf("repeat report=%+v err=%v installs=%d", r, err, b.calls)
	}
	b = newBackend()
	b.failure = errors.New("build failed")
	r, err = maintainInstalled(context.Background(), false, b)
	if err == nil || r.Updated || r.Installed.Version != "v0.8.0" || b.installed.Version != "v0.8.0" {
		t.Fatalf("failed update changed installation: %+v %v", r, err)
	}
	b = newBackend()
	b.releaseError = errors.New("offline")
	if _, err = maintainInstalled(context.Background(), false, b); err == nil || b.calls != 0 {
		t.Fatal("release failure installed")
	}
}

func TestMaintainInstalledNeverRecreatesRemovedMint(t *testing.T) {
	for _, disappearAt := range []int{2, 3, 4} {
		b := &lifecycleBackend{fakeBackend: fakeBackend{installed: Installation{Path: "/fixture/mint", Version: "v0.8.0", Method: "go-install", Status: "installed"}, latest: "v0.9.0"}, disappearAt: disappearAt}
		r, err := maintainInstalled(context.Background(), false, b)
		if err == nil || r.Updated || b.calls != 0 {
			t.Fatalf("removal at inspection%d recreated Mint: %+v %v installs=%d", disappearAt, r, err, b.calls)
		}
	}
}
