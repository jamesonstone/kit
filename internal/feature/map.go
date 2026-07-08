package feature

import (
	"fmt"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
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
