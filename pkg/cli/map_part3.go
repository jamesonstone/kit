package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

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
