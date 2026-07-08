package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

func readPolicyRank(policy string) int {
	switch policy {
	case document.ReferenceReadPolicyMust:
		return 0
	case document.ReferenceReadPolicyConditional:
		return 1
	case document.ReferenceReadPolicyEvidence:
		return 2
	case document.ReferenceReadPolicySkip:
		return 3
	default:
		return 4
	}
}

func appendUniqueSorted(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	values = append(values, value)
	sort.Strings(values)
	return values
}

func contextResolutionLabel(entry contextReferenceEntry) string {
	if entry.Resolved {
		return "resolved"
	}
	if strings.TrimSpace(entry.ResolutionError) == "" {
		return "unresolved"
	}
	return "unresolved: " + entry.ResolutionError
}

func nonEmptyMapValue(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func outputMapWarnings(out io.Writer, style humanOutputStyle, glyphs mapGlyphs, warnings []feature.MapWarning) {
	if len(warnings) == 0 {
		return
	}

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, mapWarningTitle(style))
	for i, warning := range warnings {
		prefix := glyphs.TreeMid
		if i == len(warnings)-1 {
			prefix = glyphs.TreeLast
		}
		if strings.TrimSpace(warning.Line) != "" {
			_, _ = fmt.Fprintf(
				out,
				"%s %s/%s: %s %q (%s)\n",
				mapTreePrefix(style, prefix),
				mapFeatureName(style, warning.FeatureID),
				mapEdgeSourceDoc(style, warning.Document),
				mapWarningLead(style, "skipped invalid RELATIONSHIPS line"),
				warning.Line,
				mapMutedIfEnabled(style, warning.Message),
			)
			continue
		}
		_, _ = fmt.Fprintf(
			out,
			"%s %s/%s: %s\n",
			mapTreePrefix(style, prefix),
			mapFeatureName(style, warning.FeatureID),
			mapEdgeSourceDoc(style, warning.Document),
			mapMutedIfEnabled(style, warning.Message),
		)
	}
}

func formatFeatureDocStatus(glyphs mapGlyphs, docs []feature.MapDocument) string {
	parts := make([]string, 0, len(docs))
	for _, doc := range docs {
		parts = append(parts, fmt.Sprintf("%s%s", featureDocKey(doc.Name), documentPresenceMarker(glyphs, doc.Exists)))
	}
	return strings.Join(parts, " ")
}

func formatFeatureDocStatusStyled(style humanOutputStyle, glyphs mapGlyphs, docs []feature.MapDocument) string {
	parts := make([]string, 0, len(docs))
	for _, doc := range docs {
		parts = append(parts, fmt.Sprintf("%s%s", mapDocKey(style, featureDocKey(doc.Name)), mapPresenceMarker(style, glyphs, doc.Exists)))
	}
	return strings.Join(parts, " ")
}

func featureDocKey(name string) string {
	switch name {
	case "BRAINSTORM.md":
		return "B"
	case "SPEC.md":
		return "S"
	case "PLAN.md":
		return "P"
	case "TASKS.md":
		return "T"
	case "ANALYSIS.md":
		return "A"
	default:
		return "?"
	}
}

func documentPresenceMarker(glyphs mapGlyphs, exists bool) string {
	if exists {
		return glyphs.Present
	}
	return glyphs.Missing
}

func formatEdgeTarget(style humanOutputStyle, edge feature.RelationshipEdge) string {
	if edge.Resolved {
		return mapFeatureName(style, edge.TargetFeatureID)
	}
	return fmt.Sprintf("%s [%s]", mapFeatureName(style, edge.TargetFeatureID), mapUnresolvedLabel(style, resolvedLabel(edge.Resolved)))
}

func filterMapWarnings(warnings []feature.MapWarning, featureID string) []feature.MapWarning {
	if len(warnings) == 0 {
		return nil
	}

	var filtered []feature.MapWarning
	for _, warning := range warnings {
		if warning.FeatureID == featureID {
			filtered = append(filtered, warning)
		}
	}

	return filtered
}

func mapTitle(style humanOutputStyle, text string) string {
	if !style.enabled {
		return "🗺️ " + text
	}
	return whiteBold + "🗺️ " + text + reset
}

func mapTreePrefix(style humanOutputStyle, prefix string) string {
	if !style.enabled {
		return prefix
	}
	return gray + prefix + reset
}

func mapBoxGlyph(style humanOutputStyle, glyph string) string {
	if !style.enabled {
		return glyph
	}
	return gray + glyph + reset
}

func mapArrow(style humanOutputStyle, arrow string) string {
	if !style.enabled {
		return arrow
	}
	return dim + arrow + reset
}

func mapFeatureName(style humanOutputStyle, name string) string {
	if !style.enabled {
		return name
	}
	return whiteBold + name + reset
}

func mapDocumentPath(style humanOutputStyle, doc feature.MapDocument) string {
	if !style.enabled {
		return doc.Path
	}

	color := whiteBold
	switch doc.Name {
	case "CONSTITUTION.md":
		color = constitution
	case "BRAINSTORM.md":
		color = brainstorm
	case "SPEC.md":
		color = spec
	case "PLAN.md":
		color = plan
	case "TASKS.md":
		color = tasks
	case "ANALYSIS.md":
		color = reflect
	}
	return color + doc.Path + reset
}
