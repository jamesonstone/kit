package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/spf13/cobra"
)

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	if config.Exists(cwd) {
		return fmt.Errorf("%s already exists, so this is a Kit project; run `kit reconcile` to bring it to the current structure", config.ConfigFileName)
	}
	parent, err := ancestorKitProject(cwd)
	if err != nil {
		return err
	}
	if parent != "" {
		return fmt.Errorf("this directory is inside the Kit project at %s; run Kit commands there (use `kit reconcile` to update it) instead of creating a nested project", parent)
	}

	deliveryCfg := defaultInitConfig()
	if err := selectInitIssueProvider(cmd, deliveryCfg); err != nil {
		return err
	}
	if initIssueTemplate {
		enabled := true
		deliveryCfg.Issues.IssueTemplate = &enabled
	}

	mintRelease, err := lookupMintActionRelease()
	if err != nil {
		return fmt.Errorf("verify latest Mint Action release (no files changed): %w", err)
	}
	plan, err := planInitProject(cwd, deliveryCfg, mintRelease)
	if err != nil {
		return err
	}
	out := os.Stdout
	if initOutputOnly && !initPlanDryRun {
		out = os.Stderr
	}
	surfaces := issueTemplateSurfaces(cwd)
	org := orgIssueTemplateWarning(cwd, plan.changes)
	interactive := streamsHaveInteractiveTerminal(os.Stdin, os.Stdout)
	askFirst := interactive && !initPlanDryRun
	var notes map[string]string
	var globalPath, globalAction, globalNote string
	if askFirst {
		if err := printInitProposal(out, cwd, deliveryCfg, plan.changes, surfaces, org); err != nil {
			return err
		}
	}
	plan.changes, notes, err = gatePrewrite(cwd, plan.changes, prewriteInit, interactive, initAllowOverwrite, initPlanDryRun, os.Stdin, out)
	if err != nil {
		return err
	}
	globalPath, globalAction, globalNote, err = decideGlobalConfigWrite(interactive, initAllowOverwrite, initPlanDryRun, os.Stdin, out)
	if err != nil {
		return err
	}
	plan.stats = recountInitRefreshStats(plan.changes)
	if !askFirst {
		if err := printPrewritePlan(out, plan.changes, notes); err != nil {
			return err
		}
		if err := printInitSidePlan(out, cwd, deliveryCfg, globalPath, globalAction, globalNote); err != nil {
			return err
		}
		if err := printSurfaceWarnings(out, surfaces); err != nil {
			return err
		}
		if err := printPrewriteWarnings(out, prewriteWarnings(plan.changes, nil), org); err != nil {
			return err
		}
	}
	if initPlanDryRun {
		printInitRefreshNotes(plan.notes, initRefreshOptions{})
		_, err := fmt.Fprintln(out, "Dry run complete. No files written.")
		return err
	}

	if !initOutputOnly {
		fmt.Println("🎒 Initializing Kit project...")
	}
	if err := applyInitRefreshFileChangesAtomically(plan.changes); err != nil {
		return err
	}
	if globalAction != "skip" {
		if err := populateGlobalConfig(initOutputOnly); err != nil {
			return err
		}
	}
	specsDir := deliveryCfg.SpecsPath(cwd)
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		return fmt.Errorf("failed to create specs directory: %w", err)
	}

	cfg := deliveryCfg
	if !initOutputOnly && interactive {
		inspectionCfg, inspection, err := config.LoadWithInspection(cwd)
		if err != nil {
			return err
		}
		changed, err := remediateProjectConfig(cwd, inspectionCfg, inspection, configRemediationOptions{
			Interactive: true,
			Input:       os.Stdin,
			Output:      os.Stdout,
		})
		if err != nil {
			return err
		}
		if changed {
			cfg, err = config.Load(cwd)
			if err != nil {
				return err
			}
		}
	}

	if !initOutputOnly {
		printInitRefreshNotes(plan.notes, initRefreshOptions{})
		fmt.Println("\n✅ Kit project initialized!")
		fmt.Printf(
			"   Created: %d, Updated: %d, Merged: %d, Removed: %d, Skipped: %d\n",
			plan.stats.created,
			plan.stats.updated,
			plan.stats.merged,
			plan.stats.removed,
			plan.stats.skipped,
		)
	}

	constitutionFullPath := filepath.Join(cwd, cfg.ConstitutionPath)
	prompt := buildProjectInitPrompt(cwd, constitutionFullPath, managedFileDeliverySnapshotFromInitRefresh(cwd, plan.changes))
	prompt = appendMintIncompletePrompt(prompt, plan.notes)
	if err := outputPromptWithClipboardDefault(prompt, initOutputOnly, initCopy); err != nil {
		return err
	}
	if !initOutputOnly {
		printNumberedNextSteps([]string{
			initPromptNextStep(),
			"Keep the starter Constitution unchanged until implemented evidence supports durable project-wide rules",
			"Run `kit spec <feature-name>` to create your first feature",
		})
	}
	return nil
}

func printInitProposal(out io.Writer, root string, cfg *config.Config, changes []initRefreshFileChange, surfaces []issueSurface, org string) error {
	if err := printPrewritePlan(out, changes, nil); err != nil {
		return err
	}
	path, create, update, err := config.PlanGlobalConfig(defaultInitConfig())
	if err != nil {
		return err
	}
	action := "skip"
	if create {
		action = "create"
	} else if update {
		action = "overwrite"
	}
	if err := printInitSidePlan(out, root, cfg, path, action, ""); err != nil {
		return err
	}
	if err := printSurfaceWarnings(out, surfaces); err != nil {
		return err
	}
	return printPrewriteWarnings(out, prewriteWarnings(changes, nil), org)
}

func decideGlobalConfigWrite(interactive, force, dryRun bool, in io.Reader, out io.Writer) (string, string, string, error) {
	path, create, update, err := config.PlanGlobalConfig(defaultInitConfig())
	if err != nil {
		return "", "", "", err
	}
	if create {
		return path, "create", "", nil
	}
	if !update {
		return path, "skip", "exists", nil
	}
	if force {
		return path, "overwrite", "force", nil
	}
	if dryRun && interactive {
		return path, "overwrite", "would ask before writing", nil
	}
	if interactive && !dryRun {
		ok, err := confirmPrompt(in, out, "Write "+path+"?")
		if err != nil || !ok {
			return path, "skip", "existing file; pass --force to write it", err
		}
		return path, "overwrite", "", nil
	}
	return path, "skip", "existing file; pass --force to write it", nil
}

func printInitSidePlan(out io.Writer, root string, cfg *config.Config, globalPath, globalAction, globalNote string) error {
	line := fmt.Sprintf("  %s %s", globalAction, globalPath)
	if globalNote != "" {
		line += " (" + globalNote + ")"
	}
	if _, err := fmt.Fprintln(out, line); err != nil {
		return err
	}
	specs := cfg.SpecsPath(root)
	rel, err := filepath.Rel(root, specs)
	if err != nil {
		rel = cfg.SpecsDir
	}
	rel = filepath.ToSlash(rel)
	action := "create"
	if info, statErr := os.Stat(specs); statErr == nil && info.IsDir() {
		action = "skip"
	}
	_, err = fmt.Fprintf(out, "  %s %s/\n", action, rel)
	return err
}

// ancestorKitProject returns the nearest ancestor directory holding .kit.yaml.
// An ancestor that cannot be inspected is an error, never "no project".
func ancestorKitProject(dir string) (string, error) {
	for parent := filepath.Dir(dir); parent != dir; dir, parent = parent, filepath.Dir(parent) {
		path := filepath.Join(parent, config.ConfigFileName)
		if _, err := os.Stat(path); err == nil {
			return parent, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("cannot check %s for an existing Kit project: %w", path, err)
		}
	}
	return "", nil
}

func initPromptNextStep() string {
	if promptCopiedByDefault(initOutputOnly, initCopy) {
		return "Paste the copied prompt into your agent to review repository evidence and populate only verified Makefile targets"
	}
	return "Give the prompt above to your agent to review repository evidence and populate only verified Makefile targets"
}
