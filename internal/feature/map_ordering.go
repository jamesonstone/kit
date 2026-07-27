package feature

import (
	"path/filepath"
	"sort"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/instructions"
)

func featureDocuments(projectRoot string, feat Feature) []MapDocument {
	return []MapDocument{
		{
			Name:      "BRAINSTORM.md",
			Path:      relativePath(projectRoot, filepath.Join(feat.Path, "BRAINSTORM.md")),
			Required:  false,
			Exists:    document.Exists(filepath.Join(feat.Path, "BRAINSTORM.md")),
			ManagedBy: "kit legacy brainstorm",
		},
		{
			Name:      "SPEC.md",
			Path:      relativePath(projectRoot, filepath.Join(feat.Path, "SPEC.md")),
			Required:  true,
			Exists:    document.Exists(filepath.Join(feat.Path, "SPEC.md")),
			ManagedBy: "kit spec",
		},
		{
			Name:      "PLAN.md",
			Path:      relativePath(projectRoot, filepath.Join(feat.Path, "PLAN.md")),
			Required:  false,
			Exists:    document.Exists(filepath.Join(feat.Path, "PLAN.md")),
			ManagedBy: "kit legacy plan",
		},
		{
			Name:      "TASKS.md",
			Path:      relativePath(projectRoot, filepath.Join(feat.Path, "TASKS.md")),
			Required:  false,
			Exists:    document.Exists(filepath.Join(feat.Path, "TASKS.md")),
			ManagedBy: "kit legacy tasks",
		},
		{
			Name:      "ANALYSIS.md",
			Path:      relativePath(projectRoot, filepath.Join(feat.Path, "ANALYSIS.md")),
			Required:  false,
			Exists:    document.Exists(filepath.Join(feat.Path, "ANALYSIS.md")),
			ManagedBy: "manual / agent-authored",
		},
	}
}

func relativePath(projectRoot, absPath string) string {
	rel, err := filepath.Rel(projectRoot, absPath)
	if err != nil {
		return absPath
	}
	return filepath.ToSlash(rel)
}

func sortedEdges(edges []RelationshipEdge) []RelationshipEdge {
	if len(edges) == 0 {
		return nil
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].SourceFeatureID != edges[j].SourceFeatureID {
			return edges[i].SourceFeatureID < edges[j].SourceFeatureID
		}
		if edges[i].SourceDoc != edges[j].SourceDoc {
			return edges[i].SourceDoc < edges[j].SourceDoc
		}
		if edges[i].Type != edges[j].Type {
			return edges[i].Type < edges[j].Type
		}
		return edges[i].TargetFeatureID < edges[j].TargetFeatureID
	})

	return edges
}

func sortedReferenceLinks(links []ReferenceLink) []ReferenceLink {
	if len(links) == 0 {
		return nil
	}

	sort.Slice(links, func(i, j int) bool {
		if links[i].SourceFeatureID != links[j].SourceFeatureID {
			return links[i].SourceFeatureID < links[j].SourceFeatureID
		}
		if links[i].SourceDoc != links[j].SourceDoc {
			return links[i].SourceDoc < links[j].SourceDoc
		}
		if links[i].Reference != links[j].Reference {
			return links[i].Reference < links[j].Reference
		}
		if links[i].Target != links[j].Target {
			return links[i].Target < links[j].Target
		}
		return links[i].NodeID < links[j].NodeID
	})

	return links
}

func sortedWarnings(warnings []MapWarning) []MapWarning {
	if len(warnings) == 0 {
		return nil
	}

	sort.Slice(warnings, func(i, j int) bool {
		if warnings[i].FeatureID != warnings[j].FeatureID {
			return warnings[i].FeatureID < warnings[j].FeatureID
		}
		if warnings[i].Document != warnings[j].Document {
			return warnings[i].Document < warnings[j].Document
		}
		if warnings[i].Line != warnings[j].Line {
			return warnings[i].Line < warnings[j].Line
		}
		return warnings[i].Message < warnings[j].Message
	})

	return warnings
}

func logicallyOrderedFeatureMaps(features []FeatureMap) []FeatureMap {
	if len(features) < 2 {
		return features
	}

	featureByID := make(map[string]FeatureMap, len(features))
	orderIndex := make(map[string]int, len(features))
	indegree := make(map[string]int, len(features))
	dependents := make(map[string][]string, len(features))

	for i, featureMap := range features {
		id := featureMap.Feature.DirName
		featureByID[id] = featureMap
		orderIndex[id] = i
		indegree[id] = 0
	}

	for _, featureMap := range features {
		for _, edge := range featureMap.Outgoing {
			if !edge.Resolved || !relationshipOrdersFeatures(edge.Type) {
				continue
			}
			if _, ok := featureByID[edge.TargetFeatureID]; !ok {
				continue
			}

			indegree[featureMap.Feature.DirName]++
			dependents[edge.TargetFeatureID] = append(dependents[edge.TargetFeatureID], featureMap.Feature.DirName)
		}
	}

	queue := make([]string, 0, len(features))
	for _, featureMap := range features {
		id := featureMap.Feature.DirName
		if indegree[id] == 0 {
			queue = append(queue, id)
		}
	}

	var ordered []FeatureMap
	seen := make(map[string]bool, len(features))
	for len(queue) > 0 {
		sort.Slice(queue, func(i, j int) bool {
			return orderIndex[queue[i]] < orderIndex[queue[j]]
		})

		id := queue[0]
		queue = queue[1:]
		if seen[id] {
			continue
		}

		seen[id] = true
		ordered = append(ordered, featureByID[id])

		for _, dependent := range dependents[id] {
			indegree[dependent]--
			if indegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	if len(ordered) == len(features) {
		return ordered
	}

	for _, featureMap := range features {
		if !seen[featureMap.Feature.DirName] {
			ordered = append(ordered, featureMap)
		}
	}

	return ordered
}

func relationshipOrdersFeatures(relationshipType string) bool {
	return relationshipType == "builds on" || relationshipType == "depends on"
}

func appendMapDocuments(projectRoot string, docs []MapDocument, registryDocs []instructions.Doc) []MapDocument {
	for _, doc := range registryDocs {
		docs = append(docs, MapDocument{
			Name:      filepath.Base(doc.RelativePath),
			Path:      doc.RelativePath,
			Required:  doc.Required,
			Exists:    document.Exists(filepath.Join(projectRoot, filepath.FromSlash(doc.RelativePath))),
			ManagedBy: doc.ManagedBy,
		})
	}

	return docs
}
