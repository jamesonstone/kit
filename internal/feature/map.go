package feature

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/instructions"
)

type MapDocument struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Required  bool   `json:"required"`
	Exists    bool   `json:"exists"`
	ManagedBy string `json:"managed_by"`
}

type RelationshipEdge struct {
	SourceFeatureID string `json:"source_feature_id"`
	SourceDoc       string `json:"source_doc"`
	Type            string `json:"type"`
	TargetFeatureID string `json:"target_feature_id"`
	Resolved        bool   `json:"resolved"`
}

type ReferenceLink struct {
	ID              string `json:"id,omitempty"`
	SourceFeatureID string `json:"source_feature_id"`
	SourceDoc       string `json:"source_doc"`
	Reference       string `json:"reference"`
	Type            string `json:"type"`
	Target          string `json:"target"`
	SelectorType    string `json:"selector_type,omitempty"`
	Selector        string `json:"selector,omitempty"`
	Relation        string `json:"relation"`
	ReadPolicy      string `json:"read_policy"`
	UsedFor         string `json:"used_for"`
	Status          string `json:"status"`
	NodeID          string `json:"node_id"`
	Resolved        bool   `json:"resolved"`
	Resolution      string `json:"resolution,omitempty"`
	ResolutionError string `json:"resolution_error,omitempty"`
}

type MapWarning struct {
	FeatureID string `json:"feature_id"`
	Document  string `json:"document"`
	Line      string `json:"line,omitempty"`
	Message   string `json:"message"`
}

type FeatureMap struct {
	Feature    Feature            `json:"feature"`
	Documents  []MapDocument      `json:"documents"`
	Outgoing   []RelationshipEdge `json:"outgoing"`
	Incoming   []RelationshipEdge `json:"incoming"`
	References []ReferenceLink    `json:"references"`
}

type ProjectMap struct {
	GlobalDocuments []MapDocument `json:"global_documents"`
	Features        []FeatureMap  `json:"features"`
	Warnings        []MapWarning  `json:"warnings"`
}

func BuildProjectMap(projectRoot string, cfg *config.Config) (*ProjectMap, error) {
	specsDir := cfg.SpecsPath(projectRoot)
	features, err := ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to list features: %w", err)
	}

	knownFeatures := make(map[string]struct{}, len(features))
	for _, feat := range features {
		knownFeatures[feat.DirName] = struct{}{}
	}

	projectMap := &ProjectMap{
		GlobalDocuments: projectGlobalDocuments(projectRoot, cfg),
	}

	for _, feat := range features {
		outgoing, warnings, err := loadRelationshipEdges(feat, knownFeatures)
		if err != nil {
			return nil, err
		}
		references, referenceWarnings, err := loadReferenceLinks(projectRoot, cfg, feat)
		if err != nil {
			return nil, err
		}
		projectMap.Features = append(projectMap.Features, FeatureMap{
			Feature:    feat,
			Documents:  featureDocuments(projectRoot, feat),
			Outgoing:   outgoing,
			References: references,
		})
		projectMap.Warnings = append(projectMap.Warnings, warnings...)
		projectMap.Warnings = append(projectMap.Warnings, referenceWarnings...)
	}

	incomingByTarget := make(map[string][]RelationshipEdge)
	for _, featureMap := range projectMap.Features {
		for _, edge := range featureMap.Outgoing {
			incomingByTarget[edge.TargetFeatureID] = append(incomingByTarget[edge.TargetFeatureID], edge)
		}
	}

	for i := range projectMap.Features {
		projectMap.Features[i].Incoming = sortedEdges(incomingByTarget[projectMap.Features[i].Feature.DirName])
		projectMap.Features[i].Outgoing = sortedEdges(projectMap.Features[i].Outgoing)
	}
	projectMap.Features = logicallyOrderedFeatureMaps(projectMap.Features)
	projectMap.Warnings = sortedWarnings(projectMap.Warnings)

	return projectMap, nil
}

func loadReferenceLinks(projectRoot string, cfg *config.Config, feat Feature) ([]ReferenceLink, []MapWarning, error) {
	sources := []struct {
		name    string
		path    string
		docType document.DocumentType
	}{
		{name: "BRAINSTORM.md", path: filepath.Join(feat.Path, "BRAINSTORM.md"), docType: document.TypeBrainstorm},
		{name: "SPEC.md", path: filepath.Join(feat.Path, "SPEC.md"), docType: document.TypeSpec},
		{name: "PLAN.md", path: filepath.Join(feat.Path, "PLAN.md"), docType: document.TypePlan},
		{name: "TASKS.md", path: filepath.Join(feat.Path, "TASKS.md"), docType: document.TypeTasks},
	}

	var links []ReferenceLink
	var warnings []MapWarning
	for _, source := range sources {
		if !document.Exists(source.path) {
			continue
		}

		doc, err := document.ParseFile(source.path, source.docType)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse %s references for %s: %w", source.name, feat.DirName, err)
		}

		for _, reference := range doc.References() {
			resolution := resolveReference(projectRoot, cfg, reference)
			links = append(links, ReferenceLink{
				ID:              reference.ID,
				SourceFeatureID: feat.DirName,
				SourceDoc:       source.name,
				Reference:       reference.Name,
				Type:            reference.Type,
				Target:          reference.Target,
				SelectorType:    reference.SelectorType,
				Selector:        reference.Selector,
				Relation:        reference.Relation,
				ReadPolicy:      reference.ReadPolicy,
				UsedFor:         reference.UsedFor,
				Status:          reference.Status,
				NodeID:          resolution.NodeID,
				Resolved:        resolution.Resolved,
				Resolution:      resolution.Kind,
				ResolutionError: resolution.Error,
			})
			if !resolution.Resolved &&
				reference.ReadPolicy != document.ReferenceReadPolicySkip &&
				reference.Status == document.ReferenceStatusActive {
				warnings = append(warnings, MapWarning{
					FeatureID: feat.DirName,
					Document:  source.name,
					Message:   fmt.Sprintf("unresolved reference %q: %s", reference.Name, resolution.Error),
				})
			}
		}
	}

	return sortedReferenceLinks(links), sortedWarnings(warnings), nil
}

func loadRelationshipEdges(feat Feature, knownFeatures map[string]struct{}) ([]RelationshipEdge, []MapWarning, error) {
	sources := []struct {
		name    string
		path    string
		docType document.DocumentType
	}{
		{name: "BRAINSTORM.md", path: filepath.Join(feat.Path, "BRAINSTORM.md"), docType: document.TypeBrainstorm},
		{name: "SPEC.md", path: filepath.Join(feat.Path, "SPEC.md"), docType: document.TypeSpec},
	}

	var edges []RelationshipEdge
	var warnings []MapWarning
	for _, source := range sources {
		if !document.Exists(source.path) {
			continue
		}

		doc, err := document.ParseFile(source.path, source.docType)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse %s for %s: %w", source.name, feat.DirName, err)
		}

		relationships, parseWarnings := doc.Relationships()
		for _, parseWarning := range parseWarnings {
			warnings = append(warnings, MapWarning{
				FeatureID: feat.DirName,
				Document:  source.name,
				Line:      parseWarning.Line,
				Message:   parseWarning.Message,
			})
		}

		for _, relationship := range relationships {
			_, resolved := knownFeatures[relationship.Target]
			edges = append(edges, RelationshipEdge{
				SourceFeatureID: feat.DirName,
				SourceDoc:       source.name,
				Type:            relationship.Type,
				TargetFeatureID: relationship.Target,
				Resolved:        resolved,
			})
		}
	}

	return sortedEdges(edges), warnings, nil
}

func projectGlobalDocuments(projectRoot string, cfg *config.Config) []MapDocument {
	docs := []MapDocument{
		{
			Name:      "CONSTITUTION.md",
			Path:      cfg.ConstitutionPath,
			Required:  true,
			Exists:    document.Exists(cfg.ConstitutionAbsPath(projectRoot)),
			ManagedBy: "kit init",
		},
		{
			Name:      "PROJECT_PROGRESS_SUMMARY.md",
			Path:      relativePath(projectRoot, cfg.ProgressSummaryPath(projectRoot)),
			Required:  true,
			Exists:    document.Exists(cfg.ProgressSummaryPath(projectRoot)),
			ManagedBy: "Kit lifecycle commands",
		},
	}

	version := detectInstructionScaffoldVersion(projectRoot, cfg)
	if version == instructionScaffoldVersionUnknown {
		return docs
	}

	docs = appendMapDocuments(projectRoot, docs, instructions.InstructionDocs(cfg, version))

	if version != config.InstructionScaffoldVersionTOC {
		return docs
	}

	docs = appendMapDocuments(projectRoot, docs, instructions.SupportDocs(version))

	return docs
}

const instructionScaffoldVersionUnknown = instructions.UnknownVersion

func detectInstructionScaffoldVersion(projectRoot string, cfg *config.Config) int {
	return instructions.DetectVersion(projectRoot, cfg)
}

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
