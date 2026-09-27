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
	start := strings.Index(existing, templates.UniversalContractBeginMarker)
	end := strings.Index(existing, templates.UniversalContractEndMarker)
	if start < 0 || end < start {
		return contractBlockMerge{handled: true, legacy: true, content: existing}
	}
	end += len(templates.UniversalContractEndMarker)
	return contractBlockMerge{handled: true, content: existing[:start] + block + existing[end:]}
}

func managedContractBlock(content string) (string, bool) {
	start := strings.Index(content, templates.UniversalContractBeginMarker)
	end := strings.Index(content, templates.UniversalContractEndMarker)
	if start < 0 || end < start {
		return "", false
	}
	return content[start : end+len(templates.UniversalContractEndMarker)], true
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
			if !strings.Contains(content, templates.UniversalContractBeginMarker) {
				version = config.InstructionScaffoldVersionTOC
				break
			}
		}
	}
	return templates.ConstitutionBaselineSectionFor(version)
}
