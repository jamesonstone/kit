package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func renderCIDiagnosisJSON(w io.Writer, diagnosis ciDiagnosis) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(diagnosis)
}

func renderCIDiagnosisHuman(w io.Writer, diagnosis ciDiagnosis) error {
	style := styleForWriter(w)
	if _, err := fmt.Fprintln(w, style.title("🩺", "CI Diagnosis")); err != nil {
		return err
	}
	if err := renderCISection(w, style, "Target", renderCITargetLines(diagnosis.Target)); err != nil {
		return err
	}
	if err := renderCISection(w, style, "Failing Checks", renderCICheckLines(diagnosis)); err != nil {
		return err
	}
	if err := renderCISection(w, style, "Root Cause", []string{diagnosis.RootCause}); err != nil {
		return err
	}
	if err := renderCISection(w, style, "Evidence", renderCIEvidenceLines(diagnosis)); err != nil {
		return err
	}
	if err := renderCISection(w, style, "Recommended Fix", []string{diagnosis.Recommendation}); err != nil {
		return err
	}
	return renderCISection(w, style, "Agent Prompt", []string{"```text", diagnosis.AgentPrompt, "```"})
}

func renderCISection(w io.Writer, style humanOutputStyle, title string, lines []string) error {
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, style.label(title)); err != nil {
		return err
	}
	if len(lines) == 0 {
		lines = []string{"none"}
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

func renderCITargetLines(target ciTarget) []string {
	lines := []string{
		"- repository: " + target.Repository,
		"- kind: " + target.Kind,
	}
	if target.Branch != "" {
		lines = append(lines, "- branch: "+target.Branch)
	}
	if target.PRNumber != 0 {
		lines = append(lines, "- pr: #"+strconv.Itoa(target.PRNumber))
	}
	if target.RunID != 0 {
		lines = append(lines, "- run: "+strconv.FormatInt(target.RunID, 10))
	}
	if target.Workflow != "" {
		lines = append(lines, "- workflow: "+target.Workflow)
	}
	if target.Job != "" {
		lines = append(lines, "- job: "+target.Job)
	}
	if target.HeadSHA != "" {
		lines = append(lines, "- head_sha: "+target.HeadSHA)
	}
	return lines
}

func renderCICheckLines(diagnosis ciDiagnosis) []string {
	if !diagnosis.FailureFound {
		return []string{"none"}
	}
	var lines []string
	for _, check := range diagnosis.FailingChecks {
		line := "- " + check.Name
		if check.Workflow != "" {
			line += " (" + check.Workflow + ")"
		}
		if check.Link != "" {
			line += ": " + check.Link
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		for _, run := range diagnosis.Runs {
			line := "- run " + strconv.FormatInt(run.RunID, 10)
			if run.Workflow != "" {
				line += " (" + run.Workflow + ")"
			}
			if run.URL != "" {
				line += ": " + run.URL
			}
			lines = append(lines, line)
		}
	}
	for _, check := range diagnosis.ExternalChecks {
		lines = append(lines, "- external: "+check.Name+" ("+firstNonEmpty(check.State, check.Bucket)+")")
	}
	return lines
}

func renderCIEvidenceLines(diagnosis ciDiagnosis) []string {
	if len(diagnosis.Runs) == 0 {
		if len(diagnosis.Evidence) == 0 {
			return []string{"No GitHub Actions log excerpt was available."}
		}
		return diagnosis.Evidence
	}
	var lines []string
	for _, run := range diagnosis.Runs {
		lines = append(lines, fmt.Sprintf("Run %d: %s", run.RunID, firstNonEmpty(run.Workflow, run.Name)))
		for _, job := range run.FailedJobs {
			lines = append(lines, fmt.Sprintf("Job: %s", job.Name))
			if len(job.FailedSteps) > 0 {
				lines = append(lines, "Failed steps: "+strings.Join(job.FailedSteps, ", "))
			}
			lines = append(lines, "```text")
			lines = append(lines, job.LogExcerpt...)
			lines = append(lines, "```")
		}
	}
	return lines
}

func buildCIAgentPrompt(diagnosis ciDiagnosis) string {
	return renderBuilderText(func(sb *strings.Builder) {
		sb.WriteString("Fix the GitHub Actions failure described below.\n\n")
		sb.WriteString("Verify each finding against current code. Fix only still-valid issues, skip the rest with a brief reason, keep changes minimal, and validate.\n\n")
		sb.WriteString("## Target\n")
		for _, line := range renderCITargetLines(diagnosis.Target) {
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n## Root Cause\n")
		sb.WriteString(diagnosis.RootCause + "\n\n")
		sb.WriteString("## Recommended Fix\n")
		sb.WriteString(diagnosis.Recommendation + "\n\n")
		sb.WriteString("## Evidence\n")
		if len(diagnosis.Runs) == 0 && len(diagnosis.Evidence) == 0 {
			sb.WriteString("No GitHub Actions log excerpt was available.\n")
		} else {
			for _, run := range diagnosis.Runs {
				sb.WriteString(fmt.Sprintf("### Run %d: %s\n", run.RunID, firstNonEmpty(run.Workflow, run.Name)))
				for _, job := range run.FailedJobs {
					sb.WriteString(fmt.Sprintf("#### Job: %s\n", job.Name))
					if len(job.FailedSteps) > 0 {
						sb.WriteString("Failed steps: " + strings.Join(job.FailedSteps, ", ") + "\n")
					}
					sb.WriteString("```text\n")
					for _, line := range job.LogExcerpt {
						sb.WriteString(line + "\n")
					}
					sb.WriteString("```\n")
				}
			}
		}
		if len(diagnosis.ExternalChecks) > 0 {
			sb.WriteString("\n## External Failed Checks\n")
			for _, check := range diagnosis.ExternalChecks {
				sb.WriteString("- " + check.Name)
				if check.Link != "" {
					sb.WriteString(": " + check.Link)
				}
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n## Required Output\n")
		sb.WriteString("- State whether the failure is still valid.\n")
		sb.WriteString("- Summarize the minimal fix applied or explain why no code change was needed.\n")
		sb.WriteString("- List the verification commands run and their observed results.\n")
	})
}

func openCIDispatchPrompt(opts ciOptions, diagnosis ciDiagnosis) error {
	initialContent := diagnosis.AgentPrompt
	edited, err := readEditorTextWithInitialContent(
		opts.InputConfig,
		"CI dispatch tasks",
		initialContent,
		false,
		false,
	)
	if err != nil {
		return err
	}
	tasks, err := normalizeDispatchTasks(edited)
	if err != nil {
		return err
	}
	workingDirectory := strings.TrimSpace(opts.RepoPath)
	if workingDirectory == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		workingDirectory = cwd
	} else if abs, err := filepath.Abs(workingDirectory); err == nil {
		workingDirectory = abs
	}
	prompt := buildDispatchPrompt(tasks, dispatchMaxSubagents, workingDirectory, dispatchInputSourceEditor, dispatchPromptOptions{})
	return outputPromptWithoutSubagentsWithClipboardDefault(prompt, false, false)
}
