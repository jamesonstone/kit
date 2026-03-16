package cli

import (
	"errors"
	"testing"
)

func TestSelectAssetName(t *testing.T) {
	tests := []struct {
		name    string
		version string
		goos    string
		arch    string
		want    string
		wantErr bool
	}{
		{name: "linux amd64", version: "v1.2.3", goos: "linux", arch: "amd64", want: "kit_1.2.3_linux_amd64.tar.gz"},
		{name: "linux arm64", version: "v1.2.3", goos: "linux", arch: "arm64", want: "kit_1.2.3_linux_arm64.tar.gz"},
		{name: "darwin arm64", version: "v1.2.3", goos: "darwin", arch: "arm64", want: "kit_1.2.3_darwin_arm64.tar.gz"},
		{name: "windows amd64", version: "v1.2.3", goos: "windows", arch: "amd64", want: "kit_1.2.3_windows_amd64.zip"},
		{name: "strips v prefix", version: "v9.8.7", goos: "darwin", arch: "amd64", want: "kit_9.8.7_darwin_amd64.tar.gz"},
		{name: "unsupported", version: "v1.2.3", goos: "freebsd", arch: "amd64", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectAssetName(tt.version, tt.goos, tt.arch)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("selectAssetName() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("selectAssetName() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("selectAssetName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseChecksums(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "valid checksum file",
			input: "abc123  kit_1.2.3_linux_amd64.tar.gz\nDEF456  checksums.txt\n",
			want: map[string]string{
				"kit_1.2.3_linux_amd64.tar.gz": "abc123",
				"checksums.txt":                "def456",
			},
		},
		{name: "invalid line", input: "broken line here\n", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseChecksums(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseChecksums() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseChecksums() error = %v", err)
			}
			for name, want := range tt.want {
				if got[name] != want {
					t.Fatalf("parseChecksums()[%q] = %q, want %q", name, got[name], want)
				}
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    int
	}{
		{name: "dev build upgrades", current: "dev", latest: "v1.2.3", want: -1},
		{name: "equal versions", current: "v1.2.3", latest: "v1.2.3", want: 0},
		{name: "major increment", current: "v1.2.3", latest: "v2.0.0", want: -1},
		{name: "minor increment", current: "v1.2.3", latest: "v1.3.0", want: -1},
		{name: "patch increment", current: "v1.2.3", latest: "v1.2.4", want: -1},
		{name: "current newer", current: "v2.0.0", latest: "v1.9.9", want: 1},
		{name: "missing v prefix", current: "1.2.3", latest: "v1.2.3", want: 0},
		{name: "non-semantic current treated as upgradeable", current: "main", latest: "v1.2.3", want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareVersions(tt.current, tt.latest); got != tt.want {
				t.Fatalf("compareVersions() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestBuildExecutablePath(t *testing.T) {
	original := executablePath
	defer func() { executablePath = original }()

	tests := []struct {
		name    string
		mock    func() (string, error)
		want    string
		wantErr bool
	}{
		{name: "success", mock: func() (string, error) { return "/tmp/kit", nil }, want: "/tmp/kit"},
		{name: "error", mock: func() (string, error) { return "", errors.New("boom") }, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executablePath = tt.mock
			got, err := buildExecutablePath()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("buildExecutablePath() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("buildExecutablePath() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("buildExecutablePath() = %q, want %q", got, tt.want)
			}
		})
	}
}
