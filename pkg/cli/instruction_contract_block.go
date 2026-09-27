package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"

	"github.com/jamesonstone/kit/v3/internal/templates"
)

// contractBlockMerge describes how an existing entry file relates to the
// template's Kit-managed contract block.
type contractBlockMerge struct {
	handled bool   // template carries a managed block
	legacy  bool   // existing file predates the managed block
	content string // existing content with the block replaced
}

// mergeManagedContractBlock replaces only the Kit-managed contract block in an
// existing entry file. Files without the block are left untouched so refresh
// never appends a second contract below pre-contract guidance.
func mergeManagedContractBlock(existing, template string) contractBlockMerge {
	block, ok := managedContractBlock(template)
	if !ok {
		return contractBlockMerge{}
	}
	start, end, ok := managedContractBlockBounds(existing)
	if !ok {
		return contractBlockMerge{handled: true, legacy: true, content: existing}
	}
	return contractBlockMerge{handled: true, content: existing[:start] + block + existing[end:]}
}

func managedContractBlock(content string) (string, bool) {
	start, end, ok := managedContractBlockBounds(content)
	if !ok {
		return "", false
	}
	return content[start:end], true
}

// managedContractBlockBounds locates a complete block: the first begin marker
// and the first end marker after it.
func managedContractBlockBounds(content string) (int, int, bool) {
	start := strings.Index(content, templates.UniversalContractBeginMarker)
	if start < 0 {
		return 0, 0, false
	}
	rel := strings.Index(content[start:], templates.UniversalContractEndMarker)
	if rel < 0 {
		return 0, 0, false
	}
	return start, start + rel + len(templates.UniversalContractEndMarker), true
}

// constitutionBaselineForProject points the Constitution at the universal
// contract only when every entry file carries the managed block after this
// refresh; pre-contract projects keep the legacy baseline until they migrate.
func constitutionBaselineForProject(projectRoot string, cfg *config.Config, planned []initRefreshFileChange) string {
	version := cfg.EffectiveInstructionScaffoldVersion()
	if version == config.InstructionScaffoldVersionMemory {
		after := make(map[string]string, len(planned))
		for _, change := range planned {
			if change.after != "" {
				after[change.relativePath] = change.after
			}
		}
		for _, relativePath := range instructionFiles(cfg) {
			content, ok := after[filepath.ToSlash(relativePath)]
			if !ok {
				data, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(relativePath)))
				if err != nil {
					continue
				}
				content = string(data)
			}
			if _, ok := managedContractBlock(content); !ok {
				version = config.InstructionScaffoldVersionTOC
				break
			}
		}
	}
	return templates.ConstitutionBaselineSectionFor(version)
}
