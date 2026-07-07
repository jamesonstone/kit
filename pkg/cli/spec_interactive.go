package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

var promptSpecDeliveryIntent = readSpecDeliveryIntent

func readLineRL(rl *readline.Instance) string {
	line, err := rl.Readline()
	if err != nil {
		if err == readline.ErrInterrupt || err == io.EOF {
			return ""
		}
		return ""
	}
	return normalizeSpecAnswer(line)
}

func runSpecInteractive(
	specPath, brainstormPath string,
	feat *feature.Feature,
	projectRoot string,
	cfg *config.Config,
	inputCfg freeTextInputConfig,
	replaceThesis bool,
	outputOnly bool,
) (*specAnswers, error) {
	thesis, err := promptSpecThesis(inputCfg)
	if err != nil {
		return nil, err
	}
	deliveryIntent, err := promptSpecDeliveryIntent()
	if err != nil {
		return nil, err
	}
	if err := updateSpecThesisAndDeliveryIntent(specPath, thesis, deliveryIntent, replaceThesis); err != nil {
		return nil, err
	}
	answers := specAnswers{
		Problem:        thesis,
		DeliveryIntent: specDeliveryIntentPromptText(deliveryIntent),
	}
	return &answers, nil
}

func promptSpecThesis(inputCfg freeTextInputConfig) (string, error) {
	style := styleForStdout()

	printSectionBanner("📝", "Spec Thesis")
	fmt.Println(style.label("What to write"))
	fmt.Println(style.muted("  One concise paragraph or a short set of paragraphs describing the feature goal."))
	fmt.Println()
	fmt.Println(style.label("What Kit handles next"))
	fmt.Println(style.muted("  The coding agent will infer, research, clarify, and fill every other SPEC.md section before implementation."))
	if inputCfg.usesEditor() {
		fmt.Println()
		fmt.Printf("%s\n", style.muted(fmt.Sprintf("A %s will open for this response.", inputCfg.editorLabel())))
		return readEditorText(inputCfg, "feature thesis", false)
	}

	rl, err := newMultilineReadline()
	if err != nil {
		return "", fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer closeMultilineReadline(rl)

	fmt.Println()
	fmt.Println(style.muted("Press Enter to submit. Use Shift+Enter or Ctrl+J to insert newlines."))
	fmt.Println(style.muted("Consecutive blank lines are preserved."))
	rl.SetPrompt(whiteBold + "   > " + reset)
	thesis := readLineRL(rl)
	if thesis == "" {
		return "", fmt.Errorf("feature thesis cannot be empty")
	}
	return thesis, nil
}

func readSpecDeliveryIntent() (string, error) {
	rl, err := newMultilineReadline()
	if err != nil {
		return "", fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer closeMultilineReadline(rl)

	style := styleForStdout()
	printSectionBanner("🚦", "Delivery Intent")
	fmt.Println(style.muted("Kit records intent only. No Git or GitHub mutation happens here."))
	fmt.Println()
	fmt.Printf("  %s  %s\n", style.label("no"), style.muted("capture the idea only; no issue/branch/PR intent yet (default)"))
	fmt.Printf("  %s %s\n", style.label("yes"), style.muted("intend to create a new issue, branch, and PR later through Kit-managed rules"))
	fmt.Printf("  %s   %s\n", style.label("continue"), style.muted("continue on the current branch/current issue/current PR if one exists"))
	fmt.Println()
	rl.SetPrompt(whiteBold + "   delivery intent [no/yes/continue]: " + reset)
	return normalizeSpecDeliveryIntent(readLineRL(rl))
}

func normalizeSpecDeliveryIntent(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "n", "no":
		return specDeliveryIntentIdeaOnly, nil
	case "y", "yes":
		return specDeliveryIntentIssueBranchPRLater, nil
	case "c", "continue":
		return specDeliveryIntentContinueCurrent, nil
	default:
		return "", fmt.Errorf("delivery intent must be yes, no, or continue")
	}
}

func updateSpecThesisAndDeliveryIntent(specPath, thesis, deliveryIntent string, replaceThesis bool) error {
	content, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", specPath, err)
	}

	updated := string(content)
	if replaceThesis {
		updated = replaceMarkdownSection(updated, "THESIS", strings.TrimSpace(thesis))
	} else {
		updated = appendMarkdownSectionNote(
			updated,
			"THESIS",
			fmt.Sprintf("Thesis Revision - %s", time.Now().Format("2006-01-02")),
			strings.TrimSpace(thesis),
		)
	}
	updated = replaceMarkdownSection(updated, "DELIVERY DECISION", specDeliveryDecisionSection(deliveryIntent))

	doc := document.Parse(updated, specPath, document.TypeSpec)
	metadataUpdate := document.MetadataUpsert{
		WorkflowVersion: 2,
		Phase:           "clarify",
		DeliveryIntent:  deliveryIntent,
		Clarification:   clarificationState(document.ClarificationStatusOpen, 0, 1),
	}
	if doc.Metadata != nil && doc.Metadata.Feature != (document.FeatureMetadata{}) {
		metadataUpdate.Feature = doc.Metadata.Feature
	}
	withMetadata, _, err := document.UpsertMetadata(updated, document.TypeSpec, metadataUpdate)
	if err != nil {
		return fmt.Errorf("failed to update delivery intent metadata in %s: %w", specPath, err)
	}
	if err := document.Write(specPath, withMetadata); err != nil {
		return fmt.Errorf("failed to write %s: %w", specPath, err)
	}
	return nil
}

func replaceMarkdownSection(content, heading, body string) string {
	lines := strings.Split(content, "\n")
	start, end := markdownSectionBounds(lines, heading)
	replacement := []string{"## " + heading, "", strings.TrimSpace(body), ""}
	if start == -1 {
		if strings.TrimSpace(content) == "" {
			return strings.Join(replacement, "\n")
		}
		return strings.TrimRight(content, "\n") + "\n\n" + strings.Join(replacement, "\n")
	}
	updated := append([]string{}, lines[:start]...)
	updated = append(updated, replacement...)
	updated = append(updated, lines[end:]...)
	return strings.Join(updated, "\n")
}

func appendMarkdownSectionNote(content, heading, noteHeading, body string) string {
	lines := strings.Split(content, "\n")
	start, end := markdownSectionBounds(lines, heading)
	note := []string{"### " + noteHeading, "", strings.TrimSpace(body), ""}
	if start == -1 {
		return replaceMarkdownSection(content, heading, strings.Join(note, "\n"))
	}

	section := strings.Join(lines[start+1:end], "\n")
	if strings.TrimSpace(section) == "" {
		return replaceMarkdownSection(content, heading, strings.Join(note, "\n"))
	}

	updated := append([]string{}, lines[:end]...)
	updated = append(updated, "")
	updated = append(updated, note...)
	updated = append(updated, lines[end:]...)
	return strings.Join(updated, "\n")
}

func markdownSectionBounds(lines []string, heading string) (int, int) {
	target := "## " + strings.TrimSpace(heading)
	start := -1
	end := len(lines)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if start == -1 {
			if trimmed == target {
				start = i
			}
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			end = i
			break
		}
	}
	return start, end
}

func specDeliveryDecisionSection(deliveryIntent string) string {
	switch deliveryIntent {
	case specDeliveryIntentIssueBranchPRLater:
		return "User intends to create a new issue, branch, and PR later using Kit-managed repository rules.\n\nNo Git or GitHub mutation was performed by `kit spec`. Actual issue, branch, commit, push, and PR work remains behind the coding-agent delivery hard gate after clarification, acceptance criteria, and validation are stable."
	case specDeliveryIntentContinueCurrent:
		return "User intends for the coding agent to continue on the current branch/current issue/current PR if one exists.\n\nNo Git or GitHub mutation was performed by `kit spec`. The coding agent must run the repository delivery hard gate before any issue, branch, commit, push, PR, or review-thread mutation."
	case specDeliveryIntentIdeaOnly, "":
		return "Idea capture only. The user has not requested issue, branch, or PR intent yet.\n\nNo Git or GitHub mutation was performed by `kit spec`. Delivery remains undecided until the user or coding agent resolves it after clarification."
	default:
		return fmt.Sprintf("Unrecognized delivery intent recorded by Kit: `%s`.\n\nNo Git or GitHub mutation was performed by `kit spec`. Clarify delivery intent before implementation or delivery.", deliveryIntent)
	}
}

func specDeliveryIntentPromptText(deliveryIntent string) string {
	switch deliveryIntent {
	case specDeliveryIntentIssueBranchPRLater:
		return "yes - user intends to create a new issue, branch, and PR later using Kit-managed repository rules; `kit spec` performed no Git or GitHub mutation"
	case specDeliveryIntentContinueCurrent:
		return "continue - coding agent should continue on the current branch/current issue/current PR if one exists; `kit spec` performed no Git or GitHub mutation"
	case specDeliveryIntentIdeaOnly, "":
		return "no - idea-only SPEC.md capture; no issue/branch/PR intent yet and `kit spec` performed no Git or GitHub mutation"
	default:
		return deliveryIntent
	}
}
