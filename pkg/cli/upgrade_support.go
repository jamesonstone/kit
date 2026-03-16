package cli

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	releasesAPIBase   = "https://api.github.com/repos/jamesonstone/kit/releases"
	manualInstallHint = "go install github.com/jamesonstone/kit/cmd/kit@latest"
)

var (
	upgradeHTTPClient = &http.Client{Timeout: 30 * time.Second}
	executablePath    = os.Executable
)

type githubRelease struct {
	TagName    string         `json:"tag_name"`
	Prerelease bool           `json:"prerelease"`
	Assets     []releaseAsset `json:"assets"`
}

type releaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

func latestStableRelease(current string) (githubRelease, error) {
	var latest githubRelease
	if err := fetchJSON(releasesAPIBase+"/latest", current, &latest); err != nil {
		return githubRelease{}, err
	}
	if !latest.Prerelease && validVersion(latest.TagName) {
		return latest, nil
	}
	for page := 1; ; page++ {
		var releases []githubRelease
		url := fmt.Sprintf("%s?per_page=20&page=%d", releasesAPIBase, page)
		if err := fetchJSON(url, current, &releases); err != nil {
			return githubRelease{}, err
		}
		if len(releases) == 0 {
			return githubRelease{}, fmt.Errorf("no stable GitHub release found")
		}
		for _, release := range releases {
			if !release.Prerelease && validVersion(release.TagName) {
				return release, nil
			}
		}
	}
}

func fetchJSON(url, current string, target any) error {
	body, err := downloadBytes(url, current)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to decode GitHub response: %w", err)
	}
	return nil
}

func downloadBytes(url, current string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("User-Agent", "kit/"+current)
	resp, err := upgradeHTTPClient.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, fmt.Errorf("network timeout while checking for updates. install manually with `%s`", manualInstallHint)
		}
		return nil, fmt.Errorf("network error while checking for updates: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusForbidden, http.StatusTooManyRequests:
		return nil, fmt.Errorf("GitHub API rate limit exceeded for unauthenticated requests. install manually with `%s`", manualInstallHint)
	default:
		return nil, fmt.Errorf("GitHub request failed: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read download: %w", err)
	}
	return body, nil
}

func selectAssetName(version, goos, arch string) (string, error) {
	if (goos != "linux" && goos != "darwin" && goos != "windows") ||
		(arch != "amd64" && arch != "arm64") {
		return "", fmt.Errorf("unsupported platform %s/%s", goos, arch)
	}
	trimmed := strings.TrimPrefix(version, "v")
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("kit_%s_%s_%s.%s", trimmed, goos, arch, ext), nil
}

func parseChecksums(raw string) (map[string]string, error) {
	checksums := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid checksum line: %q", line)
		}
		checksums[fields[1]] = strings.ToLower(fields[0])
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read checksums: %w", err)
	}
	return checksums, nil
}

func compareVersions(current, latest string) int {
	if current == "dev" {
		return -1
	}
	cur, curOK := parseVersion(current)
	lat, latOK := parseVersion(latest)
	if !latOK {
		return 0
	}
	if !curOK {
		return -1
	}
	for i := range cur {
		if cur[i] < lat[i] {
			return -1
		}
		if cur[i] > lat[i] {
			return 1
		}
	}
	return 0
}

func parseVersion(version string) ([3]int, bool) {
	var parsed [3]int
	parts := strings.Split(strings.TrimPrefix(version, "v"), ".")
	if len(parts) != 3 {
		return parsed, false
	}
	for i, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil {
			return parsed, false
		}
		parsed[i] = value
	}
	return parsed, true
}

func validVersion(version string) bool { _, ok := parseVersion(version); return ok }

func displayVersion(version string) string {
	if version == "dev" || strings.HasPrefix(version, "v") || !validVersion(version) {
		return version
	}
	return "v" + version
}

func findAssetURL(assets []releaseAsset, name string) (string, bool) {
	for _, asset := range assets {
		if asset.Name == name {
			return asset.URL, true
		}
	}
	return "", false
}

func confirmUpgrade(cmd *cobra.Command) (bool, error) {
	if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Update now? [y/N]"); err != nil {
		return false, err
	}
	input, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(input))
	return answer == "y" || answer == "yes", nil
}

func extractBinary(archiveBytes []byte, goos string) ([]byte, error) {
	if goos == "windows" {
		reader, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
		if err != nil {
			return nil, fmt.Errorf("failed to read zip archive: %w", err)
		}
		for _, file := range reader.File {
			if path.Base(file.Name) != "kit.exe" {
				continue
			}
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open kit.exe in archive: %w", err)
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
		return nil, fmt.Errorf("kit.exe not found in release archive")
	}
	gzr, err := gzip.NewReader(bytes.NewReader(archiveBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to read tar.gz archive: %w", err)
	}
	defer gzr.Close()
	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("kit not found in release archive")
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar archive: %w", err)
		}
		if path.Base(header.Name) == "kit" {
			return io.ReadAll(tr)
		}
	}
}
