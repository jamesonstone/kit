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
	Name      string
	Path      string
	Required  bool
	Exists    bool
	ManagedBy string
}

type RelationshipEdge struct {
	SourceFeatureID string
	SourceDoc       string
	Type            string
	TargetFeatureID string
	Resolved        bool
}

type FeatureMap struct {
	Feature   Feature
	Documents []MapDocument
	Outgoing  []RelationshipEdge
	Incoming  []RelationshipEdge
}

type ProjectMap struct {
	GlobalDocuments []MapDocument
	Features        []FeatureMap
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
		outgoing, err := loadRelationshipEdges(feat, knownFeatures)
		if err != nil {
			return nil, err
		}
		projectMap.Features = append(projectMap.Features, FeatureMap{
			Feature:   feat,
			Documents: featureDocuments(projectRoot, feat),
			Outgoing:  outgoing,
		})
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

	return projectMap, nil
}

func loadRelationshipEdges(feat Feature, knownFeatures map[string]struct{}) ([]RelationshipEdge, error) {
	sources := []struct {
		name    string
		path    string
		docType document.DocumentType
	}{
		{name: "BRAINSTORM.md", path: filepath.Join(feat.Path, "BRAINSTORM.md"), docType: document.TypeBrainstorm},
		{name: "SPEC.md", path: filepath.Join(feat.Path, "SPEC.md"), docType: document.TypeSpec},
	}

	var edges []RelationshipEdge
	for _, source := range sources {
		if !document.Exists(source.path) {
			continue
		}

		doc, err := document.ParseFile(source.path, source.docType)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s for %s: %w", source.name, feat.DirName, err)
		}

		relationships, err := document.ParseRelationshipsSection(doc.GetSection("RELATIONSHIPS"))
		if err != nil {
			return nil, fmt.Errorf("failed to parse RELATIONSHIPS in %s for %s: %w", source.name, feat.DirName, err)
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

	return sortedEdges(edges), nil
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
			ManagedBy: "kit rollup",
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
			ManagedBy: "kit brainstorm",
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
			Required:  true,
			Exists:    document.Exists(filepath.Join(feat.Path, "PLAN.md")),
			ManagedBy: "kit plan",
		},
		{
			Name:      "TASKS.md",
			Path:      relativePath(projectRoot, filepath.Join(feat.Path, "TASKS.md")),
			Required:  true,
			Exists:    document.Exists(filepath.Join(feat.Path, "TASKS.md")),
			ManagedBy: "kit tasks",
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
