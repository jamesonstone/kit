package templates

import (
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
)

const (
	workLedgerNoteBegin = "<!-- BEGIN KIT-MANAGED WORK LEDGER -->"
	workLedgerNoteEnd   = "<!-- END KIT-MANAGED WORK LEDGER -->"
)

// WorkLedgerNote renders the agent-contract ledger note from repository
// config. An omitted provider returns an empty string. The text uses only
// configured values and never a person, workspace, project, URL, or bot label
// of Kit's choosing.
func WorkLedgerNote(cfg config.WorkLedgerConfig) string {
	if !cfg.Enabled() {
		return ""
	}
	states := cfg.Plane.States.Resolved()
	project := cfg.Plane.Project
	if strings.TrimSpace(cfg.Plane.Name) != "" {
		project = cfg.Plane.Name + " (" + cfg.Plane.Project + ")"
	}
	label := "unset"
	if strings.TrimSpace(cfg.OwnerLabel) != "" {
		label = cfg.OwnerLabel
	}
	body := fmt.Sprintf("Provider `%s`. Workspace `%s`. Project `%s`. States: active `%s`, in review `%s`, done `%s`. Owner label: `%s`. The repository owner supplies this target. Follow `delivery` for the lifecycle. Kit stores no credentials and does not call the ledger API.",
		strings.TrimSpace(cfg.Provider), cfg.Plane.Workspace, project, states.Active, states.InReview, states.Done, label)
	return workLedgerNoteBegin + "\n" + body + "\n" + workLedgerNoteEnd + "\n"
}

// ApplyWorkLedgerNote inserts or removes the ledger note on an agent entry
// file. The note is present only when the provider is set.
func ApplyWorkLedgerNote(content string, cfg config.WorkLedgerConfig) string {
	stripped := stripWorkLedgerNote(content)
	note := WorkLedgerNote(cfg)
	if note == "" {
		return stripped
	}
	if i := strings.Index(stripped, UniversalContractEndMarker); i >= 0 {
		at := i + len(UniversalContractEndMarker)
		return stripped[:at] + "\n\n" + note + stripped[at:]
	}
	if !strings.HasSuffix(stripped, "\n") {
		stripped += "\n"
	}
	return stripped + "\n" + note
}

func stripWorkLedgerNote(content string) string {
	for {
		start := strings.Index(content, workLedgerNoteBegin)
		if start < 0 {
			return content
		}
		end := strings.Index(content[start:], workLedgerNoteEnd)
		if end < 0 {
			return content
		}
		stop := start + end + len(workLedgerNoteEnd)
		rest := content[stop:]
		head := strings.TrimRight(content[:start], "\n")
		if strings.TrimSpace(rest) == "" {
			content = head + "\n"
			continue
		}
		content = head + "\n\n" + strings.TrimLeft(rest, "\n")
	}
}
