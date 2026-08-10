package contract

const (
	WorkTypeFeature     = "feature"
	WorkTypeMaintenance = "maintenance"
)

func applyWorkTypeGate(resolved *Resolved) {
	workType := resolved.Hints.WorkType
	implementationDelivery := contains(resolved.Hints.Workflows, "implementation-delivery")

	switch {
	case workType == "" && (implementationDelivery || resolved.Hints.Feature != ""):
		blockForWorkType(resolved,
			"implementation-delivery or feature hints require explicit work type feature or maintenance",
			"re-run contract resolution with `--work-type feature` or `--work-type maintenance`")
	case workType != "" && workType != WorkTypeFeature && workType != WorkTypeMaintenance:
		blockForWorkType(resolved,
			"work type must be feature or maintenance",
			"choose `--work-type feature` or `--work-type maintenance`")
	case workType == WorkTypeFeature && resolved.Hints.Feature == "":
		blockForWorkType(resolved,
			"feature work requires a canonical feature hint",
			"re-run contract resolution with `--work-type feature --feature <feature>`")
	case workType == WorkTypeMaintenance && resolved.Hints.Feature != "":
		blockForWorkType(resolved,
			"maintenance work cannot include a feature hint",
			"remove `--feature` or classify the work with `--work-type feature`")
	}
}

func blockForWorkType(resolved *Resolved, diagnostic, nextAction string) {
	resolved.State = "blocked"
	resolved.Diagnostics = appendUnique(resolved.Diagnostics, diagnostic)
	resolved.NextActions = appendUnique(resolved.NextActions, nextAction)
}
