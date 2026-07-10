package improve

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type benchmarkDefinition struct {
	Suite    Suite               `json:"suite"`
	Tasks    []Task              `json:"tasks"`
	Fixtures []fixtureDefinition `json:"fixtures"`
}

type fixtureDefinition struct {
	Path  string            `json:"path"`
	Files map[string]string `json:"files"`
}

func benchmarkProvenance(opts RunOptions, suite Suite, tasks []Task) (BenchmarkProvenance, error) {
	binaryPath, err := resolveBinaryPath(opts.KitBinary)
	if err != nil {
		return BenchmarkProvenance{}, err
	}
	binaryHash, err := hashFile(binaryPath)
	if err != nil {
		return BenchmarkProvenance{}, fmt.Errorf("hash Kit binary: %w", err)
	}
	runnerBinary := strings.TrimSpace(opts.RunnerBinary)
	if runnerBinary == "" {
		runnerBinary = binaryPath
	}
	runnerPath, err := resolveBinaryPath(runnerBinary)
	if err != nil {
		return BenchmarkProvenance{}, err
	}
	runnerHash, err := hashFile(runnerPath)
	if err != nil {
		return BenchmarkProvenance{}, fmt.Errorf("hash benchmark runner binary: %w", err)
	}
	definitionHash, err := hashBenchmarkDefinition(opts.ProjectRoot, suite, tasks)
	if err != nil {
		return BenchmarkProvenance{}, err
	}
	return BenchmarkProvenance{
		SuiteDefinitionSHA256: definitionHash,
		RunnerBinaryPath:      runnerPath,
		RunnerBinarySHA256:    runnerHash,
		KitBinaryPath:         binaryPath,
		KitBinarySHA256:       binaryHash,
		KitVersion:            opts.KitVersion,
		HarnessGitCommit:      opts.GitCommit,
	}, nil
}

func resolveBinaryPath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "kit"
	}
	if !strings.ContainsRune(value, filepath.Separator) {
		resolved, err := exec.LookPath(value)
		if err != nil {
			return "", fmt.Errorf("resolve Kit binary %q: %w", value, err)
		}
		value = resolved
	}
	absolute, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("resolve Kit binary path: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return "", fmt.Errorf("inspect Kit binary %q: %w", absolute, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("kit binary %q is a directory", absolute)
	}
	return absolute, nil
}

func hashBenchmarkDefinition(projectRoot string, suite Suite, tasks []Task) (string, error) {
	uniqueFixtures := map[string]struct{}{}
	for _, task := range tasks {
		uniqueFixtures[task.Fixture] = struct{}{}
	}
	fixturePaths := make([]string, 0, len(uniqueFixtures))
	for fixture := range uniqueFixtures {
		fixturePaths = append(fixturePaths, fixture)
	}
	sort.Strings(fixturePaths)

	definition := benchmarkDefinition{Suite: suite, Tasks: tasks}
	for _, fixture := range fixturePaths {
		files, err := snapshotDir(filepath.Join(projectRoot, fixture))
		if err != nil {
			return "", fmt.Errorf("hash benchmark fixture %q: %w", fixture, err)
		}
		definition.Fixtures = append(definition.Fixtures, fixtureDefinition{Path: fixture, Files: files})
	}
	data, err := json.Marshal(definition)
	if err != nil {
		return "", fmt.Errorf("encode benchmark definition: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
