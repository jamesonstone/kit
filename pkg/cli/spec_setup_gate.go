package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/templates"
)

type specSetupGateDecision string

const (
	specSetupGateContinue specSetupGateDecision = "continue"
	specSetupGateReinit   specSetupGateDecision = "re-init"
)

type specSetupStatus struct {
	Reasons             []string
	MissingConfig       bool
	MissingConstitution bool
}

func (s specSetupStatus) Ready() bool {
	return len(s.Reasons) == 0
}

func resolveSpecProjectContext(promptOnly bool) (string, *config.Config, specSetupStatus, error) {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		if promptOnly {
			return "", nil, specSetupStatus{}, err
		}
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return "", nil, specSetupStatus{}, fmt.Errorf("failed to find Kit project root and working directory: %w", cwdErr)
		}
		cfg := defaultInitConfig()
		return cwd, cfg, assessSpecSetup(cwd, cfg, true), nil
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return "", nil, specSetupStatus{}, err
	}
	return projectRoot, cfg, assessSpecSetup(projectRoot, cfg, false), nil
}

func assessSpecSetup(projectRoot string, cfg *config.Config, missingConfig bool) specSetupStatus {
	if cfg == nil {
		cfg = defaultInitConfig()
	}

	status := specSetupStatus{MissingConfig: missingConfig}
	if missingConfig {
		status.Reasons = append(status.Reasons, config.ConfigFileName+" is missing")
	}

	constitutionPath := cfg.ConstitutionAbsPath(projectRoot)
	constitutionRel := displayProjectRelativePath(projectRoot, constitutionPath)
	content, err := os.ReadFile(constitutionPath)
	if err != nil {
		if os.IsNotExist(err) {
			status.MissingConstitution = true
			status.Reasons = append(status.Reasons, constitutionRel+" is missing")
		} else {
			status.Reasons = append(status.Reasons, fmt.Sprintf("%s could not be read: %v", constitutionRel, err))
		}
	} else {
		text := string(content)
		doc := document.Parse(text, constitutionPath, document.TypeConstitution)
		bootstrap := isBootstrapConstitution(text)
		if !bootstrap && doc.HasUnresolvedPlaceholders() {
			status.Reasons = append(status.Reasons, constitutionRel+" still contains TODO placeholders")
		}
		if !bootstrap {
			status.Reasons = append(status.Reasons, constitutionPopulationReasons(constitutionRel, doc)...)
			status.Reasons = append(status.Reasons, constitutionValidationReasons(doc)...)
		}
	}

	if missingConfig {
		return status
	}

	for _, relativePath := range requiredSpecSetupInstructionPaths(cfg) {
		if !document.Exists(filepath.Join(projectRoot, relativePath)) {
			status.Reasons = append(status.Reasons, relativePath+" is missing")
		}
	}

	return status
}

func constitutionPopulationReasons(constitutionRel string, doc *document.Document) []string {
	var reasons []string
	for _, requiredSection := range doc.RequiredSections() {
		section := doc.GetSection(requiredSection)
		if section == nil {
			continue
		}
		if setupSectionHasVisibleContent(section.Content) {
			continue
		}
		reasons = append(reasons, fmt.Sprintf("%s section %q has no project-specific content", constitutionRel, requiredSection))
	}
	return reasons
}

func setupSectionHasVisibleContent(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		if setupVisibleLineContent(line) != "" {
			return true
		}
	}
	return false
}

func setupVisibleLineContent(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "<!--") {
		if idx := strings.Index(trimmed, "-->"); idx != -1 {
			return strings.TrimSpace(trimmed[idx+3:])
		}
		return ""
	}
	if idx := strings.Index(trimmed, "<!--"); idx != -1 {
		trimmed = strings.TrimSpace(trimmed[:idx])
	}
	return trimmed
}

func constitutionValidationReasons(doc *document.Document) []string {
	const maxValidationReasons = 3

	validationErrors := doc.Validate()
	if len(validationErrors) == 0 {
		return nil
	}

	reasons := make([]string, 0, maxValidationReasons+1)
	for i, validationErr := range validationErrors {
		if i >= maxValidationReasons {
			break
		}
		reasons = append(reasons, validationErr.Error())
	}
	if remaining := len(validationErrors) - maxValidationReasons; remaining > 0 {
		reasons = append(reasons, fmt.Sprintf("%s has %d more validation issue(s)", displayProjectRelativePath("", doc.Path), remaining))
	}
	return reasons
}

func requiredSpecSetupInstructionPaths(cfg *config.Config) []string {
	return instructionArtifactPaths(
		cfg,
		instructionFileSelection{},
		cfg.EffectiveInstructionScaffoldVersion(),
		true,
	)
}

func runSpecSetupGate(projectRoot string, cfg *config.Config, status specSetupStatus, outputOnly bool) (bool, error) {
	if status.Ready() {
		return false, nil
	}

	decision, err := promptSpecSetupGate(status.Reasons)
	if err != nil {
		return false, err
	}

	switch decision {
	case specSetupGateContinue:
		return false, ensureSpecSetupBypassBaseline(projectRoot, cfg, status)
	case specSetupGateReinit:
		constitutionPath := cfg.ConstitutionAbsPath(projectRoot)
		prompt := buildProjectInitPrompt(projectRoot, constitutionPath)
		if err := outputPromptWithClipboardDefault(prompt, outputOnly, specCopy); err != nil {
			return false, err
		}
		if !outputOnly {
			printNumberedNextSteps([]string{
				"Paste the copied prompt into your agent to populate docs/CONSTITUTION.md",
				"Run `kit reconcile --include-files` if missing support docs need to be restored",
				"Run `kit spec <feature>` again when setup is ready",
			})
		}
		return true, nil
	default:
		return false, fmt.Errorf("unsupported setup gate decision %q", decision)
	}
}

func ensureSpecSetupBypassBaseline(projectRoot string, cfg *config.Config, status specSetupStatus) error {
	if status.MissingConfig {
		if err := config.Save(projectRoot, cfg); err != nil {
			return fmt.Errorf("failed to create %s for spec bypass: %w", config.ConfigFileName, err)
		}
	}

	if status.MissingConstitution {
		if err := document.Write(cfg.ConstitutionAbsPath(projectRoot), templates.Constitution); err != nil {
			return fmt.Errorf("failed to create %s for spec bypass: %w", cfg.ConstitutionPath, err)
		}
	}

	return nil
}

func readSpecSetupGateDecision(reasons []string) (specSetupGateDecision, error) {
	style := styleForStdout()
	printSectionBanner("🧱", "Kit Setup Check")
	fmt.Println(style.muted("Kit project setup appears incomplete:"))
	for _, reason := range reasons {
		fmt.Printf("  %s\n", style.bullet(reason))
	}
	fmt.Println()
	fmt.Printf("  %s %s\n", style.label("continue"), style.muted("bypass setup and start the spec now (default)"))
	fmt.Printf("  %s  %s\n", style.label("re-init"), style.muted("copy the kit init prompt and stop before spec creation"))
	fmt.Println()

	rl, err := newMultilineReadline()
	if err != nil {
		return "", fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer closeMultilineReadline(rl)

	rl.SetPrompt(whiteBold + "   setup [continue/re-init]: " + reset)
	return normalizeSpecSetupGateDecision(readLineRL(rl))
}

func normalizeSpecSetupGateDecision(raw string) (specSetupGateDecision, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "c", "continue", "b", "bypass", "skip", "s", "y", "yes":
		return specSetupGateContinue, nil
	case "r", "re-init", "reinit", "init", "i":
		return specSetupGateReinit, nil
	default:
		return "", fmt.Errorf("setup gate decision must be continue or re-init")
	}
}

func displayProjectRelativePath(projectRoot, path string) string {
	if projectRoot == "" {
		return filepath.ToSlash(path)
	}
	relative, err := filepath.Rel(projectRoot, path)
	if err != nil || strings.HasPrefix(relative, "..") {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(relative)
}
