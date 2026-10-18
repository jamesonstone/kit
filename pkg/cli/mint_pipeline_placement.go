package cli

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func mintSelectionRequested(s mintPipelineSelection) bool {
	return s.workflow != "" || s.job != "" || s.beforeStep != ""
}

func validateMintSelection(s mintPipelineSelection) error {
	if mintSelectionRequested(s) && (s.workflow == "" || s.job == "" || s.beforeStep == "") {
		return fmt.Errorf("incomplete Mint selection: specify --mint-workflow, --mint-job and --mint-before-step")
	}
	return nil
}

func mintInsertionPoint(steps *yaml.Node, before string) (int, error) {
	index := -1
	ids := map[string]bool{}
	for i, step := range steps.Content {
		if id := mintField(step, "id"); id != nil {
			if ids[id.Value] || id.Value == "kit_mint_release" || id.Value == "kit_mint_changelog" {
				return 0, fmt.Errorf("duplicate/reserved Mint step id %s", id.Value)
			}
			ids[id.Value] = true
			if id.Value == before {
				index = i
			}
		}
	}
	if index < 0 {
		return 0, fmt.Errorf("--mint-before-step must select an existing step id")
	}
	return index, nil
}
