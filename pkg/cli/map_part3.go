package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

func formatReferenceReadTarget(link feature.ReferenceLink) string {
	target := nonEmptyMapValue(link.Target, "unknown target")
	if strings.TrimSpace(link.Selector) == "" {
		return target
	}
	if strings.TrimSpace(link.SelectorType) == "" {
		return target + " :: " + strings.TrimSpace(link.Selector)
	}
	return target + " :: " + strings.TrimSpace(link.SelectorType) + "=" + strings.TrimSpace(link.Selector)
}

func referenceResolutionLabel(link feature.ReferenceLink) string {
	if link.Resolved {
		return "resolved"
	}
	if strings.TrimSpace(link.ResolutionError) == "" {
		return "unresolved"
	}
	return "unresolved: " + link.ResolutionError
}

type contextReferenceGroup struct {
	ReadPolicy string                  `json:"read_policy"`
	Entries    []contextReferenceEntry `json:"entries"`
}

type contextReferenceEntry struct {
	ReadPolicy      string   `json:"read_policy"`
	ReadTarget      string   `json:"read_target"`
	NodeID          string   `json:"node_id"`
	References      []string `json:"references"`
	SourceDocs      []string `json:"source_docs"`
	Relations       []string `json:"relations"`
	UsedFor         []string `json:"used_for"`
	Resolved        bool     `json:"resolved"`
	ResolutionError string   `json:"resolution_error,omitempty"`
}

func groupedContextEntries(links []feature.ReferenceLink) map[string][]contextReferenceEntry {
	entriesByKey := map[string]contextReferenceEntry{}
	for _, link := range links {
		key := contextReferenceKey(link)
		entry := entriesByKey[key]
		if entry.ReadTarget == "" {
			entry.ReadTarget = formatReferenceReadTarget(link)
			entry.NodeID = link.NodeID
			entry.ReadPolicy = normalizedReadPolicy(link.ReadPolicy)
			entry.Resolved = link.Resolved
			entry.ResolutionError = link.ResolutionError
		}
		if readPolicyRank(normalizedReadPolicy(link.ReadPolicy)) < readPolicyRank(entry.ReadPolicy) {
			entry.ReadPolicy = normalizedReadPolicy(link.ReadPolicy)
		}
		entry.References = appendUniqueSorted(entry.References, nonEmptyMapValue(link.Reference, "unnamed reference"))
		entry.SourceDocs = appendUniqueSorted(entry.SourceDocs, nonEmptyMapValue(link.SourceDoc, "unknown source"))
		entry.Relations = appendUniqueSorted(entry.Relations, nonEmptyMapValue(link.Relation, "unspecified relation"))
		entry.UsedFor = appendUniqueSorted(entry.UsedFor, nonEmptyMapValue(link.UsedFor, "no purpose recorded"))
		if !link.Resolved {
			entry.Resolved = false
			if strings.TrimSpace(entry.ResolutionError) == "" {
				entry.ResolutionError = link.ResolutionError
			}
		}
		entriesByKey[key] = entry
	}

	groups := map[string][]contextReferenceEntry{}
	for _, entry := range entriesByKey {
		groups[entry.ReadPolicy] = append(groups[entry.ReadPolicy], entry)
	}
	for policy := range groups {
		sort.SliceStable(groups[policy], func(i, j int) bool {
			if groups[policy][i].ReadTarget != groups[policy][j].ReadTarget {
				return groups[policy][i].ReadTarget < groups[policy][j].ReadTarget
			}
			return groups[policy][i].NodeID < groups[policy][j].NodeID
		})
	}
	return groups
}

func contextReferenceKey(link feature.ReferenceLink) string {
	return strings.Join([]string{
		strings.TrimSpace(link.Target),
		strings.TrimSpace(link.SelectorType),
		strings.TrimSpace(link.Selector),
	}, "\x00")
}

func referenceReadPolicyOrder() []string {
	return []string{
		document.ReferenceReadPolicyMust,
		document.ReferenceReadPolicyConditional,
		document.ReferenceReadPolicyEvidence,
		document.ReferenceReadPolicySkip,
		"unspecified",
	}
}

func normalizedReadPolicy(policy string) string {
	policy = strings.TrimSpace(policy)
	if policy == "" {
		return "unspecified"
	}
	return policy
}

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
