package feature

import (
	"path/filepath"
	"sort"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/instructions"
)

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
