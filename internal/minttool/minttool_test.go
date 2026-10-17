package minttool

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeBackend struct {
	installed Installation
	latest    string
	failure   error
	calls     int
}

func (f *fakeBackend) inspect(context.Context) (Installation, error) { return f.installed, nil }
func (f *fakeBackend) validate(context.Context, Installation) error  { return nil }
func (f *fakeBackend) release(_ context.Context, version string) (string, error) {
	if version != "" {
		return version, nil
	}
	return f.latest, nil
}
func (f *fakeBackend) install(_ context.Context, old Installation, version string) (Installation, error) {
	f.calls++
	if f.failure != nil {
		return old, f.failure
	}
	return Installation{Version: version, Method: "go-install", Status: "installed"}, nil
}

func TestCheckAndUpdate(t *testing.T) {
	for _, tt := range []struct {
		name, version, method, status, latest string
		opts                                  Options
		wantErr                               bool
		calls                                 int
	}{
		{"missing", "", "none", "missing", "v0.8.0", Options{}, false, 1},
		{"current", "v0.8.0", "go-install", "installed", "v0.8.0", Options{}, false, 0},
		{"old", "v0.7.0", "go-install", "installed", "v0.8.0", Options{}, false, 1},
		{"dev", "(devel)", "go-install", "installed", "v0.8.0", Options{}, true, 0},
		{"pseudo", "v0.8.1-0.20260101000000-abc", "go-install", "installed", "v0.8.0", Options{}, true, 0},
		{"unknown", "v0.7.0", "unknown", "unsupported", "v0.8.0", Options{}, true, 0},
		{"downgrade", "v0.9.0", "go-install", "installed", "v0.8.0", Options{}, true, 0},
		{"target", "v0.7.0", "go-install", "installed", "v0.9.0", Options{TargetVersion: "v0.8.0"}, false, 1},
		{"unstable", "v0.7.0", "go-install", "installed", "v0.8.0", Options{TargetVersion: "v0.9.0-rc.1"}, true, 0},
		{"required", "v0.7.0", "go-install", "installed", "v0.8.0", Options{RequireVersion: "v0.9.0"}, true, 0},
		{"unreleased", "v0.8.0", "go-install", "installed", "v0.8.0", Options{RequireLinear: true}, true, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			b := &fakeBackend{installed: Installation{Version: tt.version, Method: tt.method, Status: tt.status}, latest: tt.latest}
			_, _ = check(context.Background(), tt.opts, b)
			if b.calls != 0 {
				t.Fatal("check installed Mint")
			}
			r, err := update(context.Background(), tt.opts, b)
			if (err != nil) != tt.wantErr || b.calls != tt.calls {
				t.Fatalf("report=%+v err=%v calls=%d", r, err, b.calls)
			}
		})
	}
}

func TestCompatibilityAndFailure(t *testing.T) {
	b := &fakeBackend{installed: Installation{Version: "v0.7.0", Method: "go-install", Status: "installed"}, latest: "v0.8.0"}
	r, err := check(context.Background(), Options{RequireVersion: "v0.8.0"}, b)
	if err == nil || r.Compatibility != "incompatible" {
		t.Fatalf("%+v %v", r, err)
	}
	b.installed.Version = "v0.8.0"
	r, err = check(context.Background(), Options{RequireVersion: "v0.8.0"}, b)
	if err != nil || r.Compatibility != "compatible" {
		t.Fatalf("%+v %v", r, err)
	}
	b.installed.Version = "v0.7.0"
	b.failure = errors.New("mock build failure")
	r, err = update(context.Background(), Options{}, b)
	if err == nil || r.Updated || r.Installed.Version != "v0.7.0" {
		t.Fatalf("%+v %v", r, err)
	}
}

func TestUpdatePreview(t *testing.T) {
	for _, tt := range []struct {
		status, method, version string
		opts                    Options
		wantErr                 bool
	}{
		{"missing", "none", "", Options{}, false},
		{"installed", "go-install", "v0.7.0", Options{RequireVersion: "v0.8.0"}, false},
		{"unsupported", "unknown", "v0.7.0", Options{}, true},
		{"installed", "go-install", "v0.9.0", Options{}, true},
		{"installed", "go-install", "v0.7.0", Options{RequireLinear: true}, true},
	} {
		b := &fakeBackend{installed: Installation{Status: tt.status, Method: tt.method, Version: tt.version}, latest: "v0.8.0"}
		r, err := updatePlan(context.Background(), tt.opts, b, true)
		if (err != nil) != tt.wantErr || b.calls != 0 || r.Updated {
			t.Fatalf("%+v %v calls=%d", r, err, b.calls)
		}
	}
}

func TestReleaseValidation(t *testing.T) {
	for _, body := range []string{
		`{"tag_name":"v0.8.0","published_at":"2026-01-01"}`,
		`{"tag_name":"v0.8.0","published_at":"2026-01-01","draft":true}`,
		`{"tag_name":"v0.8.0","published_at":"2026-01-01","prerelease":true}`,
		`{"tag_name":"v0.8.0"}`, `{"tag_name":"v0.8.0-rc.1","published_at":"x"}`, `broken`,
	} {
		_, err := decodeRelease(strings.NewReader(body), "v0.8.0")
		if (err == nil) != (body == `{"tag_name":"v0.8.0","published_at":"2026-01-01"}`) {
			t.Errorf("release %s: %v", body, err)
		}
	}
	if _, err := decodeRelease(strings.NewReader(`{"tag_name":"v0.8.0","published_at":"x"}`), "v0.7.0"); err == nil {
		t.Fatal("accepted wrong tag")
	}
	for _, invalid := range []string{"v01.2.3", "v1.2", "1.2.3", "v1.2.3+build", "v18446744073709551616.0.0"} {
		if validVersion(invalid) {
			t.Errorf("accepted %s", invalid)
		}
	}
}
