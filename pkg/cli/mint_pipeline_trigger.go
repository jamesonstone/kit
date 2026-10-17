package cli

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Publication is guarded to the actual default branch. Accept all ordinary push
// spellings, but never infer that an unknown or excluding filter can reach it.
func validateMintPushTrigger(root string, workflow *yaml.Node) error {
	on := mintField(workflow, "on")
	if on != nil && on.Kind == yaml.ScalarNode && on.Value == "push" {
		return nil
	}
	if on != nil && on.Kind == yaml.SequenceNode {
		for _, event := range on.Content {
			if event.Kind == yaml.ScalarNode && event.Value == "push" {
				return nil
			}
		}
	}
	push := mintField(on, "push")
	if push == nil || (push.Kind != yaml.MappingNode && push.Tag != "!!null") {
		return fmt.Errorf("mint publication requires an existing branch-push trigger; tag-only/reusable triggers need a project adapter")
	}
	if mintField(push, "tags") != nil || mintField(push, "tags-ignore") != nil || mintField(push, "branches-ignore") != nil {
		return fmt.Errorf("mint publication incomplete: tag/ignore filters require a reviewed default-branch adapter")
	}
	branches := mintField(push, "branches")
	if branches == nil {
		return nil
	}
	branch, known := mintKnownDefaultBranch(root)
	if !known {
		return fmt.Errorf("mint publication incomplete: verify origin/HEAD default-branch metadata before adopting a filtered push trigger")
	}
	included := false
	if branches.Kind == yaml.SequenceNode {
		for _, filter := range branches.Content {
			if filter.Kind != yaml.ScalarNode || strings.HasPrefix(filter.Value, "!") {
				return fmt.Errorf("mint publication incomplete: negative/dynamic branch filters require a reviewed adapter")
			}
			if filter.Value == branch {
				included = true
			}
		}
	}
	if !included {
		return fmt.Errorf("mint publication incomplete: push branches must explicitly include verified default branch %s", branch)
	}
	return nil
}
