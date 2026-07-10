package improve

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	EvalDir     = "docs/evals/kit-improve"
	ArtifactDir = ".kit/improve"
)

func LoadSuite(projectRoot, name string) (Suite, []Task, error) {
	if strings.TrimSpace(name) == "" {
		name = "default"
	}
	suitePath := filepath.Join(projectRoot, EvalDir, "suites", name+".yaml")
	data, err := os.ReadFile(suitePath)
	if err != nil {
		return Suite{}, nil, err
	}
	var suite Suite
	if err := decodeStrictYAML(data, &suite); err != nil {
		return Suite{}, nil, err
	}
	if err := validateSuite(suite); err != nil {
		return Suite{}, nil, err
	}
	tasks, err := loadTasks(projectRoot)
	if err != nil {
		return Suite{}, nil, err
	}
	selected := selectTasks(suite, tasks)
	if len(selected) < suite.MinimumTasks {
		return Suite{}, nil, fmt.Errorf("suite %q selected %d tasks, minimum is %d", suite.ID, len(selected), suite.MinimumTasks)
	}
	return suite, selected, nil
}

func loadTasks(projectRoot string) ([]Task, error) {
	dir := filepath.Join(projectRoot, EvalDir, "tasks")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	for _, entry := range entries {
		if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml")) {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var task Task
		if err := decodeStrictYAML(data, &task); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if err := validateTask(task); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })
	return tasks, nil
}

func decodeStrictYAML(data []byte, out any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	return decoder.Decode(out)
}

func validateSuite(suite Suite) error {
	if suite.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported suite schema_version %d", suite.SchemaVersion)
	}
	if strings.TrimSpace(suite.ID) == "" {
		return fmt.Errorf("suite id is required")
	}
	if suite.Repeat <= 0 {
		return fmt.Errorf("suite repeat must be positive")
	}
	return nil
}

func validateTask(task Task) error {
	if task.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported task schema_version %d", task.SchemaVersion)
	}
	required := map[string]string{
		"id":                task.ID,
		"title":             task.Title,
		"category":          task.Category,
		"fixture":           task.Fixture,
		"expected_behavior": task.ExpectedBehavior,
		"oracle":            task.Oracle,
		"mutation_policy":   task.MutationPolicy,
	}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("task %q missing %s", task.ID, field)
		}
	}
	if len(task.Commands) == 0 && strings.TrimSpace(task.InputPrompt) == "" {
		return fmt.Errorf("task %q requires commands or input_prompt", task.ID)
	}
	if len(task.Assertions) == 0 {
		return fmt.Errorf("task %q requires assertions", task.ID)
	}
	for index, assertion := range task.Assertions {
		if err := validateAssertion(assertion, len(task.Commands)); err != nil {
			return fmt.Errorf("task %q assertion %d: %w", task.ID, index, err)
		}
	}
	return nil
}

func validateAssertion(assertion Assertion, commandCount int) error {
	supported := map[string]bool{
		"command_succeeds":            true,
		"git_diff_empty":              true,
		"stdout_contains":             true,
		"stdout_not_contains":         true,
		"stdout_nonempty":             true,
		"stdout_lines_max":            true,
		"stdout_words_max":            true,
		"stdout_estimated_tokens_max": true,
	}
	if !supported[assertion.Type] {
		return fmt.Errorf("unsupported assertion type %q", assertion.Type)
	}
	if assertion.Type == "git_diff_empty" {
		return nil
	}
	if assertion.CommandIndex < 0 || assertion.CommandIndex >= commandCount {
		return fmt.Errorf("command_index %d is outside %d commands", assertion.CommandIndex, commandCount)
	}
	switch assertion.Type {
	case "stdout_contains", "stdout_not_contains":
		if assertion.Value == "" {
			return fmt.Errorf("value is required for %s", assertion.Type)
		}
	case "stdout_lines_max", "stdout_words_max", "stdout_estimated_tokens_max":
		if assertion.Max <= 0 {
			return fmt.Errorf("max must be positive for %s", assertion.Type)
		}
	}
	return nil
}

func selectTasks(suite Suite, tasks []Task) []Task {
	var selected []Task
	for _, task := range tasks {
		if taskMatchesSelector(task, suite.HeldOut) {
			continue
		}
		if len(suite.HeldIn.IncludeTags) > 0 && !taskMatchesSelector(task, suite.HeldIn) {
			continue
		}
		selected = append(selected, task)
	}
	return selected
}

func taskMatchesSelector(task Task, selector TaskSelector) bool {
	if len(selector.IncludeTags) == 0 {
		return false
	}
	include := map[string]struct{}{}
	for _, tag := range selector.IncludeTags {
		include[tag] = struct{}{}
	}
	for _, tag := range task.RegressionTags {
		if _, ok := include[tag]; ok {
			return true
		}
	}
	return false
}

func artifactRoot(projectRoot string) string {
	return filepath.Join(projectRoot, ArtifactDir)
}
