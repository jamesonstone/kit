package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var reconcileCopy bool
var reconcileOutputOnly bool
var reconcileAll bool
var reconcileMigrateReferences bool
var reconcileMigrateVerification bool
var reconcileIncludeFiles bool
var reconcileForce bool
var reconcileDryRun bool
var reconcileDiff bool
var reconcileRefreshFiles []string

var promptReconcileMenu = readReconcileMenu

type reconcileMenuChoice struct {
	IncludeFiles bool
	Force        bool
	OutputPrompt bool
}

var reconcileCmd = &cobra.Command{
	Use:   "reconcile [feature]",
	Short: "Reconcile Kit-managed project files, rules, and docs",
	Long: `Audit Kit-managed project documents and scaffold artifacts against the
current Kit contract, and optionally apply Kit-managed project-file and
ruleset refreshes.

Without a feature argument, reconciles the whole project by default.
Use --all as an explicit alias for whole-project mode.
With a feature argument, audits only that feature's docs plus related project-summary drift.

In an interactive terminal, Kit asks whether to include files, whether to force
the file refresh, and whether to output a coding-agent prompt too.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runReconcile,
}

func init() {
	reconcileCmd.Flags().BoolVar(&reconcileCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	reconcileCmd.Flags().BoolVar(&reconcileOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	reconcileCmd.Flags().BoolVar(&reconcileAll, "all", false, "audit the whole project explicitly")
	reconcileCmd.Flags().BoolVar(&reconcileMigrateReferences, "migrate-references", false, "include instructions for migrating deprecated front matter dependencies to references")
	reconcileCmd.Flags().BoolVar(&reconcileMigrateVerification, "migrate-verification", false, "include advisory instructions for adding executable verification fields to active tasks")
	reconcileCmd.Flags().BoolVar(&reconcileIncludeFiles, "include-files", false, "include Kit-managed file and ruleset refreshes before auditing docs")
	reconcileCmd.Flags().BoolVarP(&reconcileForce, "force", "f", false, "force included Kit-managed file and ruleset refreshes")
	reconcileCmd.Flags().BoolVar(&reconcileDryRun, "dry-run", false, "preview included file refreshes without writing files")
	reconcileCmd.Flags().BoolVar(&reconcileDiff, "diff", false, "print planned included file refreshes as a unified diff with --dry-run")
	reconcileCmd.Flags().StringArrayVar(&reconcileRefreshFiles, "file", nil, "limit included refresh to one Kit-managed file; repeat for multiple files")
	addPromptOnlyFlag(reconcileCmd)
	rootCmd.AddCommand(reconcileCmd)
}

func runReconcile(cmd *cobra.Command, args []string) error {
	if reconcileAll && len(args) > 0 {
		return fmt.Errorf("--all cannot be used with a feature argument")
	}
	if reconcileDiff && !reconcileDryRun {
		return fmt.Errorf("--diff requires --dry-run")
	}

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var feat *feature.Feature
	if len(args) == 1 {
		feat, err = loadFeatureWithState(cfg.SpecsPath(projectRoot), cfg, args[0])
		if err != nil {
			return fmt.Errorf("failed to resolve feature: %w", err)
		}
	}

	promptOnly := promptOnlyEnabled(cmd)
	includeFiles := reconcileIncludeFiles || reconcileForce || reconcileDryRun || reconcileDiff || len(reconcileRefreshFiles) > 0
	outputPrompt := !(reconcileDryRun || reconcileDiff)
	if shouldPromptReconcileMenu(cmd, len(args) > 0, promptOnly) {
		choice, err := promptReconcileMenu(cmd.InOrStdin(), cmd.OutOrStdout())
		if err != nil {
			return err
		}
		includeFiles = choice.IncludeFiles
		reconcileForce = choice.Force
		outputPrompt = choice.OutputPrompt
	}
	if promptOnly {
		includeFiles = false
		outputPrompt = true
	}
	if includeFiles {
		if err := runInitRefresh(projectRoot, initRefreshOptions{
			force:                       reconcileForce,
			dryRun:                      reconcileDryRun,
			diff:                        reconcileDiff,
			files:                       reconcileRefreshFiles,
			outputOnly:                  reconcileOutputOnly,
			suppressDocumentationPrompt: true,
		}); err != nil {
			return err
		}
		if !reconcileDryRun {
			cfg, err = config.Load(projectRoot)
			if err != nil {
				return fmt.Errorf("failed to reload config after included file refresh: %w", err)
			}
		}
	}

	report, err := buildReconcileReport(projectRoot, cfg, feat)
	if err != nil {
		return err
	}
	report.ReferenceMigration = reconcileMigrateReferences
	report.VerificationMigration = reconcileMigrateVerification
	if active, err := feature.FindActiveFeatureWithState(cfg.SpecsPath(projectRoot), cfg); err != nil {
		return fmt.Errorf("failed to resolve active feature: %w", err)
	} else if feat == nil || (active != nil && active.DirName == feat.DirName) {
		report.Findings = append(report.Findings, auditActiveFrontendRulesetAdvisory(projectRoot, active)...)
	}

	if len(report.Findings) == 0 && !report.ReferenceMigration && !report.VerificationMigration {
		if outputPrompt && includeFiles && !reconcileDryRun {
			if !reconcileOutputOnly {
				fmt.Fprintln(cmd.OutOrStdout(), "\nCoding-agent prompt:")
			}
			return outputPromptWithClipboardDefault(buildInitRefreshDocumentationPrompt(projectRoot, cfg), reconcileOutputOnly, reconcileCopy)
		}
		_, err := fmt.Fprintln(cmd.OutOrStdout(), report.cleanResult())
		return err
	}

	outputOnly, _ := cmd.Flags().GetBool("output-only")
	if !outputPrompt {
		if !outputOnly {
			printReconcileSummary(report)
		}
		return nil
	}
	if !outputOnly {
		printReconcileSummary(report)
		printWorkflowInstructions("reconcile (supporting step)", []string{
			"run the generated prompt in the current coding agent session",
			"keep changes limited to documentation reconciliation",
		})
	}

	return outputPromptWithClipboardDefault(buildReconcilePrompt(report), outputOnly, reconcileCopy)
}

func shouldPromptReconcileMenu(cmd *cobra.Command, featureScoped bool, promptOnly bool) bool {
	if featureScoped || promptOnly || reconcileOutputOnly {
		return false
	}
	for _, flag := range []string{"include-files", "force", "dry-run", "diff", "file", "migrate-references", "migrate-verification"} {
		if cmd.Flags().Changed(flag) {
			return false
		}
	}
	inFile, ok := cmd.InOrStdin().(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(inFile.Fd()))
}

func readReconcileMenu(in io.Reader, out io.Writer) (reconcileMenuChoice, error) {
	style := styleForWriter(out)
	if _, err := fmt.Fprintln(out); err != nil {
		return reconcileMenuChoice{}, err
	}
	if _, err := fmt.Fprintln(out, style.title("🧩", "Reconcile Options")); err != nil {
		return reconcileMenuChoice{}, err
	}
	reader := bufio.NewReader(in)
	includeFiles, err := promptReconcileBool(reader, out, "include files?", true)
	if err != nil {
		return reconcileMenuChoice{}, err
	}
	force := false
	if includeFiles {
		force, err = promptReconcileBool(reader, out, "force these changes?", false)
		if err != nil {
			return reconcileMenuChoice{}, err
		}
	}
	outputPrompt, err := promptReconcileBool(reader, out, "output coding-agent prompt too?", true)
	if err != nil {
		return reconcileMenuChoice{}, err
	}
	return reconcileMenuChoice{
		IncludeFiles: includeFiles,
		Force:        force,
		OutputPrompt: outputPrompt,
	}, nil
}

func promptReconcileBool(reader *bufio.Reader, out io.Writer, question string, defaultValue bool) (bool, error) {
	suffix := "[Y/n]"
	if !defaultValue {
		suffix = "[y/N]"
	}
	if _, err := fmt.Fprintf(out, "  %s %s ", question, suffix); err != nil {
		return false, err
	}
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("failed to read reconcile option %q: %w", question, err)
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	switch answer {
	case "":
		return defaultValue, nil
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("%s must be yes or no", question)
	}
}
