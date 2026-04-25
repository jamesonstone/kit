package feature

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

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

type DependencyLink struct {
	SourceFeatureID string
	SourceDoc       string
	Dependency      string
	Type            string
	Location        string
	UsedFor         string
	Status          string
}

type MapWarning struct {
	FeatureID string
	Document  string
	Line      string
	Message   string
}

type FeatureMap struct {
	Feature      Feature
	Documents    []MapDocument
	Outgoing     []RelationshipEdge
	Incoming     []RelationshipEdge
	Dependencies []DependencyLink
}

type ProjectMap struct {
	GlobalDocuments []MapDocument
	Features        []FeatureMap
	Warnings        []MapWarning
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
		dependencies, err := loadDependencyLinks(feat)
		if err != nil {
			return nil, err
		}
		projectMap.Features = append(projectMap.Features, FeatureMap{
			Feature:      feat,
			Documents:    featureDocuments(projectRoot, feat),
			Outgoing:     outgoing,
			Dependencies: dependencies,
		})
		projectMap.Warnings = append(projectMap.Warnings, warnings...)
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

func loadDependencyLinks(feat Feature) ([]DependencyLink, error) {
	sources := []struct {
		name    string
		path    string
		docType document.DocumentType
	}{
		{name: "BRAINSTORM.md", path: filepath.Join(feat.Path, "BRAINSTORM.md"), docType: document.TypeBrainstorm},
		{name: "SPEC.md", path: filepath.Join(feat.Path, "SPEC.md"), docType: document.TypeSpec},
		{name: "PLAN.md", path: filepath.Join(feat.Path, "PLAN.md"), docType: document.TypePlan},
	}

	var links []DependencyLink
	for _, source := range sources {
		if !document.Exists(source.path) {
			continue
		}

		doc, err := document.ParseFile(source.path, source.docType)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s dependencies for %s: %w", source.name, feat.DirName, err)
		}

		for _, link := range dependencyLinksFromSection(doc.GetSection("DEPENDENCIES")) {
			link.SourceFeatureID = feat.DirName
			link.SourceDoc = source.name
			links = append(links, link)
		}
	}

	return sortedDependencyLinks(links), nil
}

func dependencyLinksFromSection(section *document.Section) []DependencyLink {
	if section == nil {
		return nil
	}

	rows := dependencyTableRows(section.Content)
	if len(rows) < 3 {
		return nil
	}

	header := dependencyHeaderIndex(rows[0])
	required := []string{"dependency", "type", "location", "used for", "status"}
	for _, key := range required {
		if _, ok := header[key]; !ok {
			return nil
		}
	}

	var links []DependencyLink
	for _, row := range rows[2:] {
		dependency := dependencyCell(row, header["dependency"])
		if dependency == "" || strings.EqualFold(dependency, "none") {
			continue
		}
		links = append(links, DependencyLink{
			Dependency: dependency,
			Type:       dependencyCell(row, header["type"]),
			Location:   dependencyCell(row, header["location"]),
			UsedFor:    dependencyCell(row, header["used for"]),
			Status:     dependencyCell(row, header["status"]),
		})
	}

	return links
}

func dependencyTableRows(content string) [][]string {
	var rows [][]string
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if !strings.HasPrefix(line, "|") || !strings.Contains(strings.Trim(line, "|"), "|") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		rows = append(rows, cells)
	}

	return rows
}

func dependencyHeaderIndex(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for i, cell := range header {
		index[strings.ToLower(strings.TrimSpace(cell))] = i
	}
	return index
}

func dependencyCell(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
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

		relationships, parseWarnings := document.ParseRelationshipsSectionRelaxed(doc.GetSection("RELATIONSHIPS"))
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

func sortedDependencyLinks(links []DependencyLink) []DependencyLink {
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
		if links[i].Dependency != links[j].Dependency {
			return links[i].Dependency < links[j].Dependency
		}
		return links[i].Location < links[j].Location
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
