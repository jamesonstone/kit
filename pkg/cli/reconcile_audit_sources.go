package cli

func newFinding(severity reconcileSeverity, path, issue, update string) reconcileFinding {
	return reconcileFinding{
		Severity:          severity,
		FilePath:          path,
		Issue:             issue,
		UpdateInstruction: update,
	}
}
