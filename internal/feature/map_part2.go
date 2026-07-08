package feature

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/instructions"
)

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
