// Package minttool checks and explicitly updates the locally resolved Mint tool.
package minttool

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const Module = "github.com/jamesonstone/mint"
const Command = Module + "/cmd/mint"

// LinearMinimumVersion is verified by v0.9.0, tag commit eac91490f772adbd2ab030cb6eba3910c1bca93c.
const LinearMinimumVersion = "v0.9.0"

type Options struct {
	TargetVersion  string
	RequireVersion string
	RequireLinear  bool
}

type Installation struct {
	Path    string `json:"path,omitempty"`
	Version string `json:"version,omitempty"`
	Method  string `json:"method"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
}

type Report struct {
	Installed       Installation `json:"installed"`
	Latest          string       `json:"latest,omitempty"`
	Target          string       `json:"target,omitempty"`
	Compatibility   string       `json:"compatibility"`
	Updated         bool         `json:"updated"`
	UpdateAvailable bool         `json:"update_available"`
	VersionStatus   string       `json:"version_status"`
}

type backend interface {
	inspect(context.Context) (Installation, error)
	release(context.Context, string) (string, error)
	validate(context.Context, Installation) error
	install(context.Context, Installation, string) (Installation, error)
}

// Check reads local executable metadata and official release metadata. It never installs.
func Check(ctx context.Context, opts Options) (Report, error) {
	return check(ctx, opts, systemBackend{})
}

// Update installs a published version only after validating installation ownership.
func Update(ctx context.Context, opts Options) (Report, error) {
	return update(ctx, opts, systemBackend{})
}

// PreviewUpdate validates the proposed update without building or replacing Mint.
func PreviewUpdate(ctx context.Context, opts Options) (Report, error) {
	return updatePlan(ctx, opts, systemBackend{}, true)
}

func check(ctx context.Context, opts Options, b backend) (Report, error) {
	r := Report{Compatibility: "not_requested"}
	for _, version := range []string{opts.TargetVersion, opts.RequireVersion} {
		if version != "" && !validVersion(version) {
			return r, errors.New("version must be a stable vX.Y.Z tag")
		}
	}
	var err error
	r.Installed, err = b.inspect(ctx)
	if err != nil {
		return r, err
	}
	r.Latest, err = b.release(ctx, "")
	if err != nil {
		return r, err
	}
	if !validVersion(r.Latest) {
		return r, errors.New("latest Mint release is not a stable version")
	}
	r.Target = r.Latest
	if opts.TargetVersion != "" {
		r.Target, err = b.release(ctx, opts.TargetVersion)
		if err != nil {
			return r, err
		}
		if r.Target != opts.TargetVersion {
			return r, errors.New("mint release tag mismatch")
		}
	}
	r.VersionStatus = r.Installed.Status
	if r.Installed.Status == "installed" && validVersion(r.Installed.Version) {
		switch compare(r.Target, r.Installed.Version) {
		case -1:
			r.VersionStatus = "newer"
		case 0:
			r.VersionStatus = "current"
		case 1:
			r.VersionStatus = "outdated"
			r.UpdateAvailable = true
		}
	} else if r.Installed.Status == "installed" {
		r.VersionStatus = "development_or_unversioned"
	} else if r.Installed.Status == "missing" {
		r.UpdateAvailable = true
	}
	if opts.RequireLinear && LinearMinimumVersion == "" {
		r.Compatibility = "unknown_unreleased"
		return r, errors.New("linear compatibility has no verified published Mint minimum version")
	}
	required := opts.RequireVersion
	if opts.RequireLinear && (required == "" || compare(LinearMinimumVersion, required) > 0) {
		required = LinearMinimumVersion
	}
	if required != "" {
		r.Compatibility = "incompatible"
		if !validVersion(r.Installed.Version) || compare(r.Installed.Version, required) < 0 {
			return r, fmt.Errorf("mint %s or newer is required", required)
		}
		r.Compatibility = "compatible"
	}
	return r, nil
}

func update(ctx context.Context, opts Options, b backend) (Report, error) {
	return updatePlan(ctx, opts, b, false)
}

func updatePlan(ctx context.Context, opts Options, b backend, preview bool) (Report, error) {
	// A version requirement can be satisfied by this update, not necessarily the old binary.
	r, err := check(ctx, Options{TargetVersion: opts.TargetVersion}, b)
	if err != nil {
		return r, err
	}
	if opts.RequireLinear && LinearMinimumVersion == "" {
		r.Compatibility = "unknown_unreleased"
		return r, errors.New("linear compatibility has no verified published Mint minimum version")
	}
	required := opts.RequireVersion
	if required != "" && !validVersion(required) {
		return r, errors.New("required version must be a stable vX.Y.Z tag")
	}
	if opts.RequireLinear && (required == "" || compare(LinearMinimumVersion, required) > 0) {
		required = LinearMinimumVersion
	}
	if required != "" {
		if !validVersion(required) {
			return r, errors.New("required version must be a stable vX.Y.Z tag")
		}
		if compare(r.Target, required) < 0 {
			r.Compatibility = "incompatible"
			return r, errors.New("target is below required Mint version")
		}
		r.Compatibility = "compatible"
	}
	if r.Installed.Status != "missing" && r.Installed.Method != "go-install" {
		return r, fmt.Errorf("unsupported Mint installation: %s", r.Installed.Reason)
	}
	if r.Installed.Status != "missing" {
		if !validVersion(r.Installed.Version) {
			return r, errors.New("refusing to replace an unversioned, development, or pseudo-version Mint")
		}
		order := compare(r.Target, r.Installed.Version)
		if order < 0 {
			return r, errors.New("refusing to downgrade Mint")
		}
		if order == 0 {
			return r, nil
		}
	}
	if err = b.validate(ctx, r.Installed); err != nil {
		return r, err
	}
	if preview {
		return r, nil
	}
	installed, err := b.install(ctx, r.Installed, r.Target)
	if err != nil {
		return r, err
	}
	r.Installed = installed
	r.Updated = true
	r.UpdateAvailable = false
	r.VersionStatus = "current"
	if opts.RequireVersion != "" || opts.RequireLinear {
		r.Compatibility = "compatible"
	}
	return r, nil
}

var stable = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

func validVersion(s string) bool {
	if !stable.MatchString(s) {
		return false
	}
	for _, p := range strings.Split(s[1:], ".") {
		if _, err := strconv.ParseUint(p, 10, 64); err != nil {
			return false
		}
	}
	return true
}

func compare(a, b string) int {
	x, y := strings.Split(a[1:], "."), strings.Split(b[1:], ".")
	for i := range x {
		u, _ := strconv.ParseUint(x[i], 10, 64)
		v, _ := strconv.ParseUint(y[i], 10, 64)
		if u < v {
			return -1
		}
		if u > v {
			return 1
		}
	}
	return 0
}
