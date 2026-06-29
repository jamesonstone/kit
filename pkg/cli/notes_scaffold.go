package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

const featureNotesPrivateGitignore = `*
!.gitignore
!README.md
`

func ensureFeatureNotesScaffold(projectRoot, featureDirName string) error {
	notesPath := featureNotesPath(projectRoot, featureDirName)
	for _, dir := range []string{
		notesPath,
		filepath.Join(notesPath, "inbox"),
		filepath.Join(notesPath, "references"),
		filepath.Join(notesPath, "responses"),
		filepath.Join(notesPath, "private"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	for _, dir := range []string{
		notesPath,
		filepath.Join(notesPath, "inbox"),
		filepath.Join(notesPath, "references"),
		filepath.Join(notesPath, "responses"),
	} {
		if err := ensurePlaceholderFile(dir); err != nil {
			return err
		}
	}

	if err := writeFileIfMissing(filepath.Join(notesPath, "README.md"), featureNotesReadme(featureDirName)); err != nil {
		return err
	}
	if err := writeFileIfMissing(filepath.Join(notesPath, "private", ".gitignore"), featureNotesPrivateGitignore); err != nil {
		return err
	}
	if err := writeFileIfMissing(filepath.Join(notesPath, "private", "README.md"), featurePrivateNotesReadme(featureDirName)); err != nil {
		return err
	}

	return nil
}

func writeFileIfMissing(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func featureNotesReadme(featureDirName string) string {
	return fmt.Sprintf(`# Feature Notes: %s

This directory stores optional source material for the feature. Notes can
include Slack excerpts, customer context, screenshots, research links, draft
responses, and other supporting inputs.

Notes are source material, not canonical truth. Promote durable decisions,
requirements, and implementation constraints into `+"`SPEC.md`"+` or another
canonical project document before relying on them for implementation.

## Directories

- `+"`inbox/`"+` - unsorted captured notes and conversation excerpts.
- `+"`references/`"+` - source material, links, research, examples, and assets.
- `+"`responses/`"+` - draft or sent responses related to the feature.
- `+"`private/`"+` - local-only sensitive conversations or tertiary context.

Agents should ignore `+"`.gitkeep`"+` files and read only the notes relevant to
the current task.
`, featureDirName)
}

func featurePrivateNotesReadme(featureDirName string) string {
	return fmt.Sprintf(`# Private Notes: %s

Files in this directory are intentionally ignored by git. Use it for sensitive
conversation history, customer-specific context, raw Slack excerpts, and other
local-only material that can inform future work without entering the repository.

Keep this README and `+"`.gitignore`"+` tracked so future agents know the private
notes location exists. Promote durable, non-sensitive conclusions into
`+"`SPEC.md`"+`, `+"`docs/CONSTITUTION.md`"+`, or another canonical project document.
`, featureDirName)
}
