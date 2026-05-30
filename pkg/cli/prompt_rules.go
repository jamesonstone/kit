package cli

import "fmt"

func docsOnlyWorkflowRule(target string) string {
	return fmt.Sprintf(
		"Only update %s; do not modify product code, tests, runtime config, generated artifacts, or implementation files.",
		target,
	)
}
