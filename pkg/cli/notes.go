package cli

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

const notesSchemaVersion = 1

var notesNow = time.Now

type notesOptions struct {
	add         bool
	copyPath    bool
	json        bool
	private     bool
	section     string
	title       string
	source      string
	status      string
	sensitivity string
}

type notesResult struct {
	SchemaVersion  int    `json:"schema_version"`
	Kind           string `json:"kind"`
	Feature        string `json:"feature"`
	NotesPath      string `json:"notes_path"`
	NotePath       string `json:"note_path,omitempty"`
	Private        bool   `json:"private,omitempty"`
	CreatedFeature bool   `json:"created_feature,omitempty"`
}

var notesCmd = newNotesCommand()

func init() {
	rootCmd.AddCommand(notesCmd)
}

func newNotesCommand() *cobra.Command {
	options := &notesOptions{
		section:     "inbox",
		source:      "manual",
		status:      "active",
		sensitivity: "internal",
	}
	cmd := &cobra.Command{
		Use:   "notes [feature]",
		Short: "Manage feature notes and local source-material scaffolds",
		Long: strings.TrimSpace(`
Manage feature notes directories under docs/notes/<feature>.

With a feature argument, Kit creates or refreshes the notes scaffold for that
feature. Without an argument, Kit opens a small selector for existing features
or a new feature notes directory.

Notes are source material, not canonical truth. Promote durable decisions,
requirements, and constraints into SPEC.md or other canonical project docs.
Private notes are stored under private/ and ignored by git while keeping the
directory contract visible.
`),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNotes(cmd, args, *options)
		},
	}
	cmd.Flags().BoolVar(&options.add, "add", false, "create a timestamped note template")
	cmd.Flags().BoolVar(&options.copyPath, "copy-path", false, "copy the feature notes directory path to the clipboard")
	cmd.Flags().BoolVar(&options.json, "json", false, "emit machine-readable JSON")
	cmd.Flags().BoolVar(&options.private, "private", false, "create the note under private/ and mark sensitivity as private")
	cmd.Flags().StringVar(&options.section, "section", options.section, "note section: inbox, references, responses, or private")
	cmd.Flags().StringVar(&options.title, "title", "", "note title used for the H1 and filename")
	cmd.Flags().StringVar(&options.source, "source", options.source, "front matter source value, such as slack or manual")
	cmd.Flags().StringVar(&options.status, "status", options.status, "front matter status value")
	cmd.Flags().StringVar(&options.sensitivity, "sensitivity", options.sensitivity, "front matter sensitivity value")
	return cmd
}

func runNotes(cmd *cobra.Command, args []string, options notesOptions) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)
	if err := ensureDir(specsDir); err != nil {
		return err
	}

	reader := bufio.NewReader(cmd.InOrStdin())
	interactive := len(args) == 0
	feat, createdFeature, err := resolveNotesFeature(projectRoot, cfg, specsDir, args, reader, cmd.OutOrStdout())
	if err != nil {
		return err
	}
	feature.ApplyLifecycleState(feat, cfg)

	if interactive && !options.add && !options.copyPath && !options.json {
		if err := promptNotesAction(reader, cmd.OutOrStdout(), &options); err != nil {
			return err
		}
	}
	applyNotesShortcuts(&options)
	if err := validateNotesOptions(options); err != nil {
		return err
	}

	notesPath, notesRelPath, err := ensureFeatureNotesDir(projectRoot, feat.DirName)
	if err != nil {
		return err
	}

	var notePath string
	if options.add {
		if interactive {
			if err := promptNoteMetadata(cmd, reader, cmd.OutOrStdout(), &options); err != nil {
				return err
			}
			applyNotesShortcuts(&options)
			if err := validateNotesOptions(options); err != nil {
				return err
			}
		}
		notePath, err = createFeatureNoteFile(projectRoot, feat.DirName, options)
		if err != nil {
			return err
		}
	}

	if options.copyPath {
		if err := clipboardCopyFunc(notesPath); err != nil {
			return fmt.Errorf("failed to copy notes path to clipboard: %w", err)
		}
	}

	result := notesResult{
		SchemaVersion:  notesSchemaVersion,
		Kind:           "notes_result",
		Feature:        feat.DirName,
		NotesPath:      notesRelPath,
		NotePath:       relativeProjectPath(projectRoot, notePath),
		Private:        options.add && effectiveNoteSection(options) == "private",
		CreatedFeature: createdFeature,
	}
	if options.json {
		return outputJSON(cmd.OutOrStdout(), result)
	}
	return printNotesResult(cmd.OutOrStdout(), result, options.copyPath)
}
