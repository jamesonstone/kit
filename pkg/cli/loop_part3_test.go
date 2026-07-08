package cli

import (
	"strconv"
	"strings"
)

func loopAgentScript(confidence int, emitResult bool, mutate bool) string {
	result := ""
	if emitResult {
		result = `printf 'KIT_LOOP_RESULT: {"stage":"%s","status":"done","confidence":` + fmtInt(confidence) + `,"blockers":[]}\n' "$KIT_LOOP_STAGE"`
	} else {
		result = `echo "done without loop result"`
	}
	mutations := ""
	if mutate {
		mutations = `case "$KIT_LOOP_STAGE" in
  clarify)
    cat > docs/specs/0001-alpha/SPEC.md <<'DOC'
` + validV2SpecWithPhase("0001-alpha", "ready") + `DOC
    ;;
  ready)
    cat > docs/specs/0001-alpha/SPEC.md <<'DOC'
` + validV2SpecWithPhase("0001-alpha", "implement") + `DOC
    ;;
  implement)
    cat > docs/specs/0001-alpha/SPEC.md <<'DOC'
` + validV2SpecWithPhase("0001-alpha", "validate") + `DOC
    ;;
  validate)
    cat > docs/specs/0001-alpha/SPEC.md <<'DOC'
` + validV2SpecWithPhase("0001-alpha", "reflect") + `DOC
    ;;
  reflect)
    cat > docs/specs/0001-alpha/SPEC.md <<'DOC'
` + validV2SpecWithPhase("0001-alpha", "deliver") + `DOC
    ;;
  deliver)
    cat > docs/specs/0001-alpha/SPEC.md <<'DOC'
` + validV2SpecWithPhase("0001-alpha", "complete") + `DOC
    ;;
esac
`
	}
	return `#!/bin/sh
set -eu
cat >/dev/null
` + mutations + result + `
`
}

func validV2SpecWithPhase(dirName string, phase string) string {
	status := "open"
	confidence := 0
	unresolved := 1
	switch phase {
	case "ready", "implement", "validate", "reflect", "deliver", "complete":
		status = "ready"
		confidence = 95
		unresolved = 0
	case "blocked":
		status = "blocked"
	}
	return v2SpecWithClarification(dirName, phase, status, confidence, unresolved)
}

func v2SpecWithClarification(dirName string, phase string, status string, confidence int, unresolved int) string {
	id, slug, ok := strings.Cut(dirName, "-")
	if !ok {
		id = ""
		slug = dirName
	}
	return `---
kit_metadata_version: 1
artifact: spec
workflow_version: 2
phase: ` + phase + `
clarification:
  status: ` + status + `
  confidence: ` + fmtInt(confidence) + `
  unresolved_questions: ` + fmtInt(unresolved) + `
feature:
  id: "` + id + `"
  slug: ` + slug + `
  dir: ` + dirName + `
---
# SPEC

## THESIS

Thesis for ` + slug + `.

## CONTEXT

Repo-grounded context.

## CLARIFICATIONS

No unresolved clarification questions.

## REQUIREMENTS

- Requirement one.

## ASSUMPTIONS

No blocking assumptions.

## ACCEPTANCE CRITERIA

- AC-001: Binary-verifiable criterion.

## IMPLEMENTATION PLAN

Implement the planned change.

## TASK CHECKLIST

- [x] T001: Maintain v2 workflow state. Expected files: ` + "`docs/specs/" + dirName + "/SPEC.md`" + `.

## VALIDATION MAP

- AC-001 -> go test ./...

## REFLECTION NOTES

No remaining risks.

## DOCUMENTATION UPDATES

README and command docs are current.

## DELIVERY DECISION

No delivery mutation requested.

## EVIDENCE

Validation evidence recorded.
`
}

func fmtInt(value int) string {
	return strconv.Itoa(value)
}

func loopReflectRunnerForTest() fakeReflectEvidenceRunner {
	return fakeReflectEvidenceRunner{
		"make test": {ExitCode: 0},
		"make lint": {ExitCode: 0},
		"git merge-base HEAD origin/main": {
			ExitCode: 0,
			Stdout:   "base\n",
		},
		"git diff --name-only base...HEAD": {
			ExitCode: 0,
			Stdout:   "docs/specs/0001-alpha/SPEC.md\n",
		},
		"git diff --name-only":                     {ExitCode: 0},
		"git diff --name-only --cached":            {ExitCode: 0},
		"git ls-files --others --exclude-standard": {ExitCode: 0},
		"git log --format=%H%x00%ct -- docs/specs/0001-alpha/SPEC.md": {
			ExitCode: 0,
			Stdout:   "readyhash\x001700000000\n",
		},
		"git show readyhash:docs/specs/0001-alpha/SPEC.md": {
			ExitCode: 0,
			Stdout:   validV2SpecWithPhase("0001-alpha", "ready"),
		},
		"git log --format=%H readyhash..HEAD -- docs/specs/0001-alpha/SPEC.md": {
			ExitCode: 0,
		},
	}
}
