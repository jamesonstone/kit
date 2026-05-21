package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

const (
	promptContextActiveFeature        = "active feature"
	promptContextOptionalFeature      = "optional feature"
	promptContextProject              = "Kit project"
	promptContextTaskList             = "task list"
	promptContextReconciliationReport = "reconciliation findings"
)

type promptFeatureContext struct {
	ProjectRoot string
	Config      *config.Config
	Feature     *feature.Feature
}

func promptDefaultEditorConfig() freeTextInputConfig {
	return newFreeTextInputConfig(false, "", false, true)
}

func promptProjectContext() (string, *config.Config, error) {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return "", nil, err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return "", nil, fmt.Errorf("failed to load config: %w", err)
	}

	return projectRoot, cfg, nil
}

func activePromptFeatureContext(command string, requiredDocs ...string) (*promptFeatureContext, error) {
	projectRoot, cfg, err := promptProjectContext()
	if err != nil {
		return nil, err
	}

	specsDir := cfg.SpecsPath(projectRoot)
	feat, err := feature.FindActiveFeatureWithState(specsDir, cfg)
	if err != nil {
		return nil, err
	}
	if feat == nil {
		feat, err = promptForFeatureContext(command, specsDir, cfg)
		if err != nil {
			return nil, err
		}
	}

	if err := requirePromptFeatureDocs(command, feat, requiredDocs...); err != nil {
		return nil, err
	}

	return &promptFeatureContext{ProjectRoot: projectRoot, Config: cfg, Feature: feat}, nil
}

func optionalPromptFeatureContext() (*promptFeatureContext, error) {
	projectRoot, cfg, err := promptProjectContext()
	if err != nil {
		return nil, err
	}

	feat, err := feature.FindActiveFeatureWithState(cfg.SpecsPath(projectRoot), cfg)
	if err != nil {
		return nil, err
	}

	return &promptFeatureContext{ProjectRoot: projectRoot, Config: cfg, Feature: feat}, nil
}

func promptForFeatureContext(command, specsDir string, cfg *config.Config) (*feature.Feature, error) {
	context, err := collectMissingPromptContext(
		command,
		"an existing feature slug or directory name",
		"prompt feature context",
		promptDefaultEditorConfig(),
	)
	if err != nil {
		return nil, err
	}

	ref := featureRefFromPromptContext(context)
	if ref == "" {
		return nil, fmt.Errorf("feature context for %q must include a feature slug or directory name", command)
	}

	feat, err := loadFeatureWithState(specsDir, cfg, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve feature context %q for %q: %w", ref, command, err)
	}
	return feat, nil
}

func collectMissingPromptContext(
	command string,
	requirement string,
	fieldName string,
	inputCfg freeTextInputConfig,
) (string, error) {
	fmt.Printf("Built-in prompt %q requires %s. Do you have this context? [y/N]: ", command, requirement)

	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to read prompt context confirmation: %w", err)
	}

	answer := strings.ToLower(strings.TrimSpace(input))
	if answer != "y" && answer != "yes" {
		return "", fmt.Errorf("built-in prompt %q requires %s", command, requirement)
	}

	context, err := readEditorText(inputCfg, fieldName, false)
	if err != nil {
		return "", err
	}
	context = strings.TrimSpace(context)
	if context == "" {
		return "", fmt.Errorf("prompt context for %q cannot be empty", command)
	}

	return context, nil
}

func featureRefFromPromptContext(context string) string {
	for _, line := range strings.Split(context, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		for _, prefix := range []string{"feature:", "feature=", "feature "} {
			if strings.HasPrefix(lower, prefix) {
				return strings.TrimSpace(line[len(prefix):])
			}
		}
		return line
	}
	return ""
}

func requirePromptFeatureDocs(command string, feat *feature.Feature, docs ...string) error {
	for _, name := range docs {
		path := filepath.Join(feat.Path, name)
		if document.Exists(path) {
			continue
		}
		return fmt.Errorf("built-in prompt %q requires %s for feature %s; expected %s", command, name, feat.Slug, path)
	}
	return nil
}
