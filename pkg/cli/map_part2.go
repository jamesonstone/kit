package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jamesonstone/kit/internal/feature"
)

func formatMapDocument(style humanOutputStyle, doc feature.MapDocument) string {
	return fmt.Sprintf(
		"%s [%s, %s] via %s",
		mapDocumentPath(style, doc),
		mapRequiredOptional(style, doc.Required),
		mapPresentMissing(style, doc.Exists),
		mapManagedBy(style, doc.ManagedBy),
	)
}

func requiredOptional(required bool) string {
	if required {
		return "required"
	}
	return "optional"
}

func presentMissing(exists bool) string {
	if exists {
		return "present"
	}
	return "missing"
}

func resolvedLabel(resolved bool) string {
	if resolved {
		return "resolved"
	}
	return "unresolved"
}

func mapYesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

type mapGlyphs struct {
	TreeMid     string
	TreeLast    string
	BoxTopLeft  string
	BoxTopRight string
	BoxBotLeft  string
	BoxBotRight string
	Horizontal  string
	Vertical    string
	Arrow       string
	Present     string
	Missing     string
	Last        string
}

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
		if len(line) > width {
			width = len(line)
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
