package minttool

import (
	"context"
	"errors"
	"fmt"
)

// MaintainInstalled updates only an existing supported Mint installation to the
// latest official stable release. Missing Mint is skipped without network reads.
// Lifecycle maintenance always requires the released Linear-compatible minimum;
// it never installs missing Mint or downgrades a newer local stable version.
func MaintainInstalled(ctx context.Context, preview bool) (Report, error) {
	return maintainInstalled(ctx, preview, systemBackend{})
}

func maintainInstalled(ctx context.Context, preview bool, b backend) (Report, error) {
	installed, err := b.inspect(ctx)
	r := Report{Installed: installed, VersionStatus: installed.Status, Compatibility: "not_applicable"}
	if err != nil {
		return r, err
	}
	if installed.Status == "missing" {
		return r, nil
	}
	if installed.Status != "installed" || installed.Method != "go-install" {
		return r, fmt.Errorf("mint maintenance preserved unsupported installation: %s; update it through its installation owner", installed.Reason)
	}
	if !validVersion(installed.Version) {
		r.VersionStatus = "development_or_unversioned"
		return r, errors.New("mint maintenance preserved a development, pseudo-version or unversioned installation; select a published stable installation explicitly")
	}
	opts := Options{RequireLinear: true}
	guard := installedBackend{backend: b, expected: installed}
	r, err = check(ctx, Options{}, guard)
	if err != nil {
		return r, err
	}
	if compare(installed.Version, r.Target) > 0 {
		required, err := minimumRequiredVersion(opts, &r)
		if err != nil {
			return r, errors.New("mint lifecycle compatibility minimum has no verified release")
		}
		if compare(installed.Version, required) < 0 {
			r.Compatibility = "incompatible"
			return r, fmt.Errorf("mint %s or newer is required; newer local installation was preserved", required)
		}
		r.Compatibility = "compatible"
		return r, nil
	}
	return applyUpdatePlan(ctx, opts, guard, preview, r)
}

// installedBackend keeps every reinspection bound to the original installed
// executable, so removal can never turn maintenance into an initial install.
type installedBackend struct {
	backend
	expected Installation
}

func (b installedBackend) inspect(ctx context.Context) (Installation, error) {
	current, err := b.backend.inspect(ctx)
	if err != nil {
		return current, err
	}
	if current.Status != "installed" || current.Method != "go-install" || current.Path != b.expected.Path || current.Version != b.expected.Version {
		return current, errors.New("mint installation changed during maintenance; rerun the check, no missing installation will be recreated")
	}
	return current, nil
}

func (b installedBackend) validate(ctx context.Context, old Installation) error {
	if _, err := b.inspect(ctx); err != nil {
		return err
	}
	return b.backend.validate(ctx, old)
}

func (b installedBackend) install(ctx context.Context, old Installation, version string) (Installation, error) {
	if _, err := b.inspect(ctx); err != nil {
		return old, err
	}
	return b.backend.install(ctx, old, version)
}
