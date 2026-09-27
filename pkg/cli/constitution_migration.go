package cli

import (
	"strings"

	"github.com/jamesonstone/kit/v3/internal/legacy"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

// legacyConstitutionPath is where every release generated the Constitution;
// fingerprints are keyed by it whatever path the project configures.
const legacyConstitutionPath = "docs/CONSTITUTION.md"

// migrateConstitutionSections replaces Constitution starter sections that an
// earlier Kit generated and nobody edited with the current starter, and drops
// unedited starter sections the current template no longer has. Edited and
// project sections are kept as written. An unedited CONSTRAINTS section carries
// the Kit-managed baseline, which points at the contract, so it is replaced only
// once every entry file carries the contract block.
func migrateConstitutionSections(content string, entriesConverged bool) string {
	current := map[string]string{}
	for _, section := range legacy.Sections(templates.Constitution) {
		if section.Key != "" {
			current[section.Key] = section.Raw
		}
	}
	sections := legacy.Sections(content)
	var out strings.Builder
	for i, section := range sections {
		raw := section.Raw
		if section.Key != "" && (entriesConverged || section.Key != "CONSTRAINTS") && legacy.SectionKnown(legacyConstitutionPath, section) {
			replacement, ok := current[section.Key]
			if !ok {
				continue
			}
			raw = replacement
		}
		if i < len(sections)-1 && !strings.HasSuffix(raw, "\n\n") {
			raw = strings.TrimRight(raw, "\n") + "\n\n"
		}
		out.WriteString(raw)
	}
	return strings.TrimRight(out.String(), "\n") + "\n"
}
