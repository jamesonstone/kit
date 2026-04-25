package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

var mapCmd = &cobra.Command{
	Use:   "map [feature]",
	Short: "Show a read-only map of canonical Kit documents",
	Long: `Render the current document hierarchy and explicit feature relationships.

Without a feature argument, shows the full project map.
With a feature argument, shows a focused map for that feature.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMap,
}

func init() {
	rootCmd.AddCommand(mapCmd)
}

func runMap(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	projectMap, err := feature.BuildProjectMap(projectRoot, cfg)
	if err != nil {
		return fmt.Errorf("failed to build document map: %w", err)
	}

	if len(args) == 0 {
		return outputProjectMap(cmd.OutOrStdout(), projectMap)
	}

	feat, err := loadFeatureWithState(cfg.SpecsPath(projectRoot), cfg, args[0])
	if err != nil {
		return err
	}

	for _, featureMap := range projectMap.Features {
		if featureMap.Feature.DirName == feat.DirName {
			return outputFeatureMap(cmd.OutOrStdout(), projectMap.GlobalDocuments, featureMap, filterMapWarnings(projectMap.Warnings, feat.DirName))
		}
	}

	return fmt.Errorf("feature '%s' not found in project map", feat.DirName)
}

func outputProjectMap(out io.Writer, projectMap *feature.ProjectMap) error {
	glyphs := selectMapGlyphs()
	style := styleForWriter(out)

	_, _ = fmt.Fprintln(out, mapTitle(style, "Kit Map"))
	_, _ = fmt.Fprintln(out)
	outputGlobalDocs(out, style, glyphs, projectMap.GlobalDocuments)
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.label("Lifecycle"))
	_, _ = fmt.Fprintf(out, "%s %s\n", mapTreePrefix(style, glyphs.Last), mapLifecycleLine(style))
	_, _ = fmt.Fprintln(out)
	outputFeatureDocKey(out, style, glyphs)
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.label("Feature Graph"))
	if len(projectMap.Features) == 0 {
		_, _ = fmt.Fprintf(out, "%s %s\n", mapTreePrefix(style, glyphs.Last), mapMutedIfEnabled(style, "none"))
	} else {
		for i, featureMap := range projectMap.Features {
			if i > 0 {
				_, _ = fmt.Fprintln(out)
			}
			renderFeatureCard(out, style, glyphs, featureMap)
			renderProjectEdges(out, style, glyphs, featureMap.Outgoing)
			renderDependencyLinks(out, style, glyphs, featureMap.Dependencies, "no dependency links")
		}
	}

	outputMapWarnings(out, style, glyphs, projectMap.Warnings)

	return nil
}

func outputFeatureMap(out io.Writer, globalDocs []feature.MapDocument, featureMap feature.FeatureMap, warnings []feature.MapWarning) error {
	glyphs := selectMapGlyphs()
	style := styleForWriter(out)

	_, _ = fmt.Fprintf(out, "%s: %s\n\n", mapTitle(style, "Kit Map"), mapFeatureName(style, featureMap.Feature.DirName))
	outputGlobalDocs(out, style, glyphs, globalDocs)
	_, _ = fmt.Fprintln(out)
	outputFeatureDocKey(out, style, glyphs)
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.label("Feature Focus"))
	renderFeatureCard(out, style, glyphs, featureMap)

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.label("Incoming Relationships"))
	renderIncomingEdges(out, style, glyphs, featureMap)

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.label("Outgoing Relationships"))
	renderProjectEdges(out, style, glyphs, featureMap.Outgoing)

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.label("Dependency Links"))
	renderDependencyLinks(out, style, glyphs, featureMap.Dependencies, "none")

	outputMapWarnings(out, style, glyphs, warnings)

	return nil
}

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
		fmt.Sprintf("%s = BRAINSTORM.md [%s] via %s", mapDocKey(style, "B"), mapRequiredOptional(style, false), mapManagedBy(style, "kit brainstorm")),
		fmt.Sprintf("%s = SPEC.md [%s] via %s", mapDocKey(style, "S"), mapRequiredOptional(style, true), mapManagedBy(style, "kit spec")),
		fmt.Sprintf("%s = PLAN.md [%s] via %s", mapDocKey(style, "P"), mapRequiredOptional(style, true), mapManagedBy(style, "kit plan")),
		fmt.Sprintf("%s = TASKS.md [%s] via %s", mapDocKey(style, "T"), mapRequiredOptional(style, true), mapManagedBy(style, "kit tasks")),
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

func renderDependencyLinks(out io.Writer, style humanOutputStyle, glyphs mapGlyphs, links []feature.DependencyLink, emptyText string) {
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
			"  %s %s dependency %s %s [%s] for %s\n",
			mapTreePrefix(style, prefix),
			mapEdgeSourceDoc(style, link.SourceDoc),
			mapDependencyName(style, link.Dependency),
			mapMutedIfEnabled(style, formatDependencyLocation(link)),
			mapDependencyStatus(style, link.Status),
			mapMutedIfEnabled(style, nonEmptyMapValue(link.UsedFor, "unspecified use")),
		)
	}
}

func formatDependencyLocation(link feature.DependencyLink) string {
	parts := []string{}
	if strings.TrimSpace(link.Type) != "" {
		parts = append(parts, strings.TrimSpace(link.Type))
	}
	if strings.TrimSpace(link.Location) != "" && !strings.EqualFold(strings.TrimSpace(link.Location), "n/a") {
		parts = append(parts, strings.TrimSpace(link.Location))
	}
	if len(parts) == 0 {
		return "(no location)"
	}
	return "(" + strings.Join(parts, ", ") + ")"
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

func mapRequiredOptional(style humanOutputStyle, required bool) string {
	label := requiredOptional(required)
	if !style.enabled {
		return label
	}
	if required {
		return whiteBold + label + reset
	}
	return dim + label + reset
}

func mapPresentMissing(style humanOutputStyle, exists bool) string {
	label := presentMissing(exists)
	if !style.enabled {
		return label
	}
	if exists {
		return plan + label + reset
	}
	return dim + label + reset
}

func mapManagedBy(style humanOutputStyle, text string) string {
	if !style.enabled {
		return text
	}
	return dim + text + reset
}

func mapDocKey(style humanOutputStyle, key string) string {
	if !style.enabled {
		return key
	}
	switch key {
	case "B":
		return brainstorm + key + reset
	case "S":
		return spec + key + reset
	case "P":
		return plan + key + reset
	case "T":
		return tasks + key + reset
	case "A":
		return reflect + key + reset
	default:
		return whiteBold + key + reset
	}
}

func mapPresenceMarker(style humanOutputStyle, glyphs mapGlyphs, exists bool) string {
	marker := documentPresenceMarker(glyphs, exists)
	if !style.enabled {
		return marker
	}
	if exists {
		return plan + marker + reset
	}
	return dim + marker + reset
}

func mapEdgeSourceDoc(style humanOutputStyle, name string) string {
	if !style.enabled {
		return name
	}
	switch name {
	case "BRAINSTORM.md":
		return brainstorm + name + reset
	case "SPEC.md":
		return spec + name + reset
	case "PLAN.md":
		return plan + name + reset
	case "TASKS.md":
		return tasks + name + reset
	default:
		return dim + name + reset
	}
}

func mapRelationshipType(style humanOutputStyle, rel string) string {
	if !style.enabled {
		return rel
	}
	return dim + rel + reset
}

func mapDependencyName(style humanOutputStyle, text string) string {
	if !style.enabled {
		return text
	}
	return whiteBold + text + reset
}

func mapDependencyStatus(style humanOutputStyle, text string) string {
	if strings.EqualFold(strings.TrimSpace(text), "stale") {
		if !style.enabled {
			return text
		}
		return implement + text + reset
	}
	if !style.enabled {
		return text
	}
	return dim + text + reset
}

func mapUnresolvedLabel(style humanOutputStyle, text string) string {
	if !style.enabled {
		return text
	}
	return implement + text + reset
}

func mapWarningTitle(style humanOutputStyle) string {
	if !style.enabled {
		return "Warnings"
	}
	return implement + "Warnings" + reset
}

func mapWarningLead(style humanOutputStyle, text string) string {
	if !style.enabled {
		return text
	}
	return implement + text + reset
}

func mapMutedIfEnabled(style humanOutputStyle, text string) string {
	if !style.enabled {
		return text
	}
	return dim + text + reset
}

func mapLifecycleLine(style humanOutputStyle) string {
	parts := []string{
		mapDocumentPath(style, feature.MapDocument{Name: "CONSTITUTION.md", Path: "CONSTITUTION.md"}),
		mapDocumentPath(style, feature.MapDocument{Name: "BRAINSTORM.md", Path: "BRAINSTORM.md"}) + mapMutedIfEnabled(style, " (optional)"),
		mapDocumentPath(style, feature.MapDocument{Name: "SPEC.md", Path: "SPEC.md"}),
		mapDocumentPath(style, feature.MapDocument{Name: "PLAN.md", Path: "PLAN.md"}),
		mapDocumentPath(style, feature.MapDocument{Name: "TASKS.md", Path: "TASKS.md"}),
		mapUnresolvedLabel(style, "implement"),
		reflect + "reflect" + reset,
	}
	if !style.enabled {
		parts = []string{"CONSTITUTION.md", "BRAINSTORM.md (optional)", "SPEC.md", "PLAN.md", "TASKS.md", "implement", "reflect"}
	}
	return strings.Join(parts, " -> ")
}
