package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jamesonstone/kit/internal/feature"
)

func selectMapGlyphs() mapGlyphs {
	if prefersASCIIMap() {
		return mapGlyphs{
			TreeMid:     "|-",
			TreeLast:    "`-",
			BoxTopLeft:  "+",
			BoxTopRight: "+",
			BoxBotLeft:  "+",
			BoxBotRight: "+",
			Horizontal:  "-",
			Vertical:    "|",
			Arrow:       "->",
			Present:     "*",
			Missing:     ".",
			Last:        "`-",
		}
	}

	return mapGlyphs{
		TreeMid:     "├─",
		TreeLast:    "└─",
		BoxTopLeft:  "┌",
		BoxTopRight: "┐",
		BoxBotLeft:  "└",
		BoxBotRight: "┘",
		Horizontal:  "─",
		Vertical:    "│",
		Arrow:       "▶",
		Present:     "●",
		Missing:     "○",
		Last:        "└─",
	}
}

func prefersASCIIMap() bool {
	locale := strings.ToUpper(strings.TrimSpace(os.Getenv("LC_ALL")))
	if locale == "" {
		locale = strings.ToUpper(strings.TrimSpace(os.Getenv("LANG")))
	}

	return locale == "C" || locale == "POSIX"
}

func outputGlobalDocs(out io.Writer, style humanOutputStyle, glyphs mapGlyphs, docs []feature.MapDocument) {
	_, _ = fmt.Fprintln(out, style.label("Global Docs"))
	if len(docs) == 0 {
		_, _ = fmt.Fprintf(out, "%s %s\n", mapTreePrefix(style, glyphs.Last), mapMutedIfEnabled(style, "none"))
		return
	}

	for i, doc := range docs {
		prefix := glyphs.TreeMid
		if i == len(docs)-1 {
			prefix = glyphs.TreeLast
		}
		_, _ = fmt.Fprintf(out, "%s %s\n", mapTreePrefix(style, prefix), formatMapDocument(style, doc))
	}
}

func outputFeatureDocKey(out io.Writer, style humanOutputStyle, glyphs mapGlyphs) {
	_, _ = fmt.Fprintln(out, style.label("Feature Doc Key"))
	rows := []string{
		fmt.Sprintf("%s = BRAINSTORM.md [%s] via %s", mapDocKey(style, "B"), mapRequiredOptional(style, false), mapManagedBy(style, "kit legacy brainstorm")),
		fmt.Sprintf("%s = SPEC.md [%s] via %s", mapDocKey(style, "S"), mapRequiredOptional(style, true), mapManagedBy(style, "kit spec")),
		fmt.Sprintf("%s = PLAN.md [%s] via %s", mapDocKey(style, "P"), mapRequiredOptional(style, false), mapManagedBy(style, "kit legacy plan")),
		fmt.Sprintf("%s = TASKS.md [%s] via %s", mapDocKey(style, "T"), mapRequiredOptional(style, false), mapManagedBy(style, "kit legacy tasks")),
		fmt.Sprintf("%s = ANALYSIS.md [%s] via %s", mapDocKey(style, "A"), mapRequiredOptional(style, false), mapManagedBy(style, "manual / agent-authored")),
		fmt.Sprintf("%s present  %s missing", mapPresenceMarker(style, glyphs, true), mapPresenceMarker(style, glyphs, false)),
	}
	for i, row := range rows {
		prefix := glyphs.TreeMid
		if i == len(rows)-1 {
			prefix = glyphs.TreeLast
		}
		_, _ = fmt.Fprintf(out, "%s %s\n", mapTreePrefix(style, prefix), row)
	}
}

func renderFeatureCard(out io.Writer, style humanOutputStyle, glyphs mapGlyphs, featureMap feature.FeatureMap) {
	rawLines := []string{
		featureMap.Feature.DirName,
		fmt.Sprintf("phase: %s | paused: %s", featureMap.Feature.Phase, mapYesNo(featureMap.Feature.Paused)),
		fmt.Sprintf("docs: %s", formatFeatureDocStatus(glyphs, featureMap.Documents)),
	}
	styledLines := []string{
		mapFeatureName(style, featureMap.Feature.DirName),
		fmt.Sprintf("phase: %s | paused: %s", formatPhaseValue(style, featureMap.Feature.Phase), formatPausedValue(style, featureMap.Feature.Paused)),
		fmt.Sprintf("docs: %s", formatFeatureDocStatusStyled(style, glyphs, featureMap.Documents)),
	}

	width := 0
	for _, line := range rawLines {
		lineWidth := len([]rune(line))
		if lineWidth > width {
			width = lineWidth
		}
	}

	_, _ = fmt.Fprintf(out, "%s%s%s\n", mapBoxGlyph(style, glyphs.BoxTopLeft), mapBoxGlyph(style, strings.Repeat(glyphs.Horizontal, width+2)), mapBoxGlyph(style, glyphs.BoxTopRight))
	for i, line := range styledLines {
		padding := spaces(width - len([]rune(rawLines[i])))
		_, _ = fmt.Fprintf(out, "%s %s%s %s\n", mapBoxGlyph(style, glyphs.Vertical), line, padding, mapBoxGlyph(style, glyphs.Vertical))
	}
	_, _ = fmt.Fprintf(out, "%s%s%s\n", mapBoxGlyph(style, glyphs.BoxBotLeft), mapBoxGlyph(style, strings.Repeat(glyphs.Horizontal, width+2)), mapBoxGlyph(style, glyphs.BoxBotRight))
}

func renderProjectEdges(out io.Writer, style humanOutputStyle, glyphs mapGlyphs, edges []feature.RelationshipEdge) {
	if len(edges) == 0 {
		_, _ = fmt.Fprintf(out, "  %s %s\n", mapTreePrefix(style, glyphs.Last), mapMutedIfEnabled(style, "no outgoing relationships"))
		return
	}

	for i, edge := range edges {
		prefix := glyphs.TreeMid
		if i == len(edges)-1 {
			prefix = glyphs.TreeLast
		}
		_, _ = fmt.Fprintf(
			out,
			"  %s %s %s %s %s\n",
			mapTreePrefix(style, prefix),
			mapEdgeSourceDoc(style, edge.SourceDoc),
			mapRelationshipType(style, edge.Type),
			mapArrow(style, glyphs.Arrow),
			formatEdgeTarget(style, edge),
		)
	}
}

func renderIncomingEdges(out io.Writer, style humanOutputStyle, glyphs mapGlyphs, featureMap feature.FeatureMap) {
	if len(featureMap.Incoming) == 0 {
		_, _ = fmt.Fprintf(out, "%s %s\n", mapTreePrefix(style, glyphs.Last), mapMutedIfEnabled(style, "none"))
		return
	}

	for i, edge := range featureMap.Incoming {
		prefix := glyphs.TreeMid
		if i == len(featureMap.Incoming)-1 {
			prefix = glyphs.TreeLast
		}
		_, _ = fmt.Fprintf(
			out,
			"%s %s %s %s %s %s\n",
			mapTreePrefix(style, prefix),
			mapFeatureName(style, edge.SourceFeatureID),
			mapEdgeSourceDoc(style, edge.SourceDoc),
			mapRelationshipType(style, edge.Type),
			mapArrow(style, glyphs.Arrow),
			formatEdgeTarget(style, edge),
		)
	}
}

func renderReferenceLinks(out io.Writer, style humanOutputStyle, glyphs mapGlyphs, links []feature.ReferenceLink, emptyText string) {
	if len(links) == 0 {
		_, _ = fmt.Fprintf(out, "  %s %s\n", mapTreePrefix(style, glyphs.Last), mapMutedIfEnabled(style, emptyText))
		return
	}

	for i, link := range links {
		prefix := glyphs.TreeMid
		if i == len(links)-1 {
			prefix = glyphs.TreeLast
		}
		_, _ = fmt.Fprintf(
			out,
			"  %s %s reference %s %s [%s, %s, %s] for %s (%s)\n",
			mapTreePrefix(style, prefix),
			mapEdgeSourceDoc(style, link.SourceDoc),
			mapReferenceName(style, link.Reference),
			mapMutedIfEnabled(style, formatReferenceTarget(link)),
			mapReferenceRelation(style, link.Relation),
			mapReferenceReadPolicy(style, link.ReadPolicy),
			mapReferenceStatus(style, link.Status),
			mapMutedIfEnabled(style, nonEmptyMapValue(link.UsedFor, "unspecified use")),
			mapMutedIfEnabled(style, referenceResolutionLabel(link)),
		)
	}
}

func formatReferenceTarget(link feature.ReferenceLink) string {
	parts := []string{}
	if strings.TrimSpace(link.Type) != "" {
		parts = append(parts, strings.TrimSpace(link.Type))
	}
	if strings.TrimSpace(link.Target) != "" && !strings.EqualFold(strings.TrimSpace(link.Target), "n/a") {
		parts = append(parts, formatReferenceReadTarget(link))
	}
	if len(parts) == 0 {
		return "(no location)"
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func outputFeatureContextPlan(out io.Writer, featureMap feature.FeatureMap) error {
	style := styleForWriter(out)
	_, _ = fmt.Fprintf(out, "%s: %s\n\n", mapTitle(style, "Kit Context Plan"), mapFeatureName(style, featureMap.Feature.DirName))
	if len(featureMap.References) == 0 {
		_, _ = fmt.Fprintln(out, "No front matter references are recorded for this feature.")
		return nil
	}

	groups := groupedContextEntries(featureMap.References)

	for _, policy := range referenceReadPolicyOrder() {
		entries := groups[policy]
		if len(entries) == 0 {
			continue
		}
		_, _ = fmt.Fprintln(out, style.label(strings.ToUpper(policy)))
		for _, entry := range entries {
			_, _ = fmt.Fprintf(
				out,
				"- %s: read `%s` from %s because %s (%s; %s)\n",
				strings.Join(entry.SourceDocs, ", "),
				entry.ReadTarget,
				strings.Join(entry.References, ", "),
				strings.Join(entry.UsedFor, "; "),
				strings.Join(entry.Relations, ", "),
				contextResolutionLabel(entry),
			)
		}
		_, _ = fmt.Fprintln(out)
	}
	return nil
}

func outputMapJSON(out io.Writer, value interface{}) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func outputContextPlanJSON(out io.Writer, featureMap feature.FeatureMap) error {
	payload := struct {
		Feature string                  `json:"feature"`
		Groups  []contextReferenceGroup `json:"groups"`
	}{
		Feature: featureMap.Feature.DirName,
	}
	groups := groupedContextEntries(featureMap.References)
	for _, policy := range referenceReadPolicyOrder() {
		if len(groups[policy]) == 0 {
			continue
		}
		payload.Groups = append(payload.Groups, contextReferenceGroup{
			ReadPolicy: policy,
			Entries:    groups[policy],
		})
	}
	return outputMapJSON(out, payload)
}
