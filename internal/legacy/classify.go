package legacy

import "strings"

// Document describes how much of a Markdown file Kit generated.
type Document struct {
	// Kit holds sections that match a released Kit's generated output.
	Kit []Section
	// Modified holds sections under a heading Kit wrote whose text no
	// released Kit generated: Kit-owned content that someone edited.
	Modified []Section
	// Project holds sections Kit never wrote, including preamble text after
	// the title line.
	Project []Section
}

// Unmodified reports whether every section is exactly Kit-generated.
func (d Document) Unmodified() bool {
	return len(d.Kit) > 0 && len(d.Modified) == 0 && len(d.Project) == 0
}

// ModifiedKeys names the edited Kit sections for diagnostics.
func (d Document) ModifiedKeys() []string {
	keys := make([]string, 0, len(d.Modified))
	for _, section := range d.Modified {
		key := section.Key
		if key == "" {
			key = "(preamble)"
		}
		keys = append(keys, key)
	}
	return keys
}

// Classify splits a file Kit generated at path into Kit, edited-Kit, and
// project sections.
func Classify(path, content string) Document {
	var doc Document
	for _, section := range Sections(content) {
		switch {
		case section.Text == "":
			continue
		case SectionKnown(path, section):
			doc.Kit = append(doc.Kit, section)
		case section.Key == "":
			title, rest := splitTitle(section.Raw)
			if title != "" {
				doc.Kit = append(doc.Kit, Section{Raw: title, Text: Normalize(title)})
			}
			if Normalize(rest) != "" {
				doc.Project = append(doc.Project, Section{Raw: rest, Text: Normalize(rest)})
			}
		case HeadingKnown(path, section.Key):
			doc.Modified = append(doc.Modified, section)
		default:
			doc.Project = append(doc.Project, section)
		}
	}
	return doc
}

// splitTitle separates a leading `# ` title line from the rest of a preamble.
func splitTitle(preamble string) (string, string) {
	trimmed := strings.TrimLeft(preamble, "\n")
	line, rest, _ := strings.Cut(trimmed, "\n")
	if strings.HasPrefix(line, "# ") {
		return line, rest
	}
	return "", preamble
}

// progressSummaryHeadings are the only `##` sections, in order, that every
// released Kit's progress-summary rollup wrote.
var progressSummaryHeadings = []string{
	"FEATURE PROGRESS TABLE",
	"PROJECT INTENT",
	"GLOBAL CONSTRAINTS",
	"FEATURE SUMMARIES",
	"LAST UPDATED",
}

// ProgressSummaryPath is where every release wrote the progress rollup.
const ProgressSummaryPath = "docs/PROJECT_PROGRESS_SUMMARY.md"

// ProgressSummaryGenerated reports whether content has exactly the structure
// Kit's rollup wrote. Kit regenerated the whole file on every run, so a file
// with that structure is Kit-owned derived data whatever its values.
func ProgressSummaryGenerated(content string) bool {
	sections := Sections(content)
	if Normalize(sections[0].Raw) != "# PROJECT PROGRESS SUMMARY" || len(sections)-1 != len(progressSummaryHeadings) {
		return false
	}
	for i, heading := range progressSummaryHeadings {
		if sections[i+1].Key != heading {
			return false
		}
	}
	return true
}
