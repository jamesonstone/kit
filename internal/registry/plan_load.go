package registry

import (
	"context"
	"fmt"
)

type loadedArtifact struct {
	Catalog  CatalogArtifact
	Content  string
	Sections []SectionHash
}

func loadVisibleArtifacts(ctx context.Context, source Source, cfg SourceConfig) ([]loadedArtifact, string, error) {
	catalog, revision, err := source.LoadCatalog(ctx, cfg)
	if err != nil {
		return nil, "", err
	}
	artifacts := VisibleArtifacts(catalog)
	loaded := make([]loadedArtifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		content, err := source.LoadArtifact(ctx, cfg, artifact, revision)
		if err != nil {
			return nil, "", err
		}
		doc, err := ParseMarkdown(content)
		if err != nil {
			return nil, "", fmt.Errorf("validate %s: %w", ArtifactKey(artifact.Kind, artifact.Slug), err)
		}
		if err := ValidateDocument(doc, artifact); err != nil {
			return nil, "", fmt.Errorf("validate %s: %w", ArtifactKey(artifact.Kind, artifact.Slug), err)
		}
		if len(artifact.AppliesTo) == 0 {
			artifact.AppliesTo = append([]string(nil), doc.Metadata.AppliesTo...)
		}
		if len(artifact.Paths) == 0 {
			artifact.Paths = append([]string(nil), doc.Metadata.Paths...)
		}
		if len(artifact.Dependencies) == 0 {
			artifact.Dependencies = append([]string(nil), doc.Metadata.Dependencies...)
		}
		sections, err := HashSections(content)
		if err != nil {
			return nil, "", err
		}
		loaded = append(loaded, loadedArtifact{Catalog: artifact, Content: content, Sections: sections})
	}
	enriched := Catalog{SchemaVersion: CatalogSchemaVersion}
	for _, item := range loaded {
		enriched.Artifacts = append(enriched.Artifacts, item.Catalog)
	}
	if err := ValidateCatalog(enriched); err != nil {
		return nil, "", fmt.Errorf("validate project-visible registry graph: %w", err)
	}
	return loaded, revision, nil
}

func recordFromRemote(source SourceConfig, revision string, loaded loadedArtifact, content, state string) ArtifactRecord {
	artifact := loaded.Catalog
	return ArtifactRecord{
		Kind:          artifact.Kind,
		Slug:          artifact.Slug,
		Description:   artifact.Description,
		Path:          artifact.TargetPath,
		Version:       artifact.Version,
		ReadPolicy:    artifact.ReadPolicy,
		AppliesTo:     append([]string(nil), artifact.AppliesTo...),
		Paths:         append([]string(nil), artifact.Paths...),
		Dependencies:  append([]string(nil), artifact.Dependencies...),
		SourceRepo:    source.Repo,
		SourceBranch:  source.Branch,
		SourceCommit:  revision,
		SourcePath:    artifact.SourcePath,
		InstalledHash: artifact.Digest,
		ContentHash:   HashContent(content),
		State:         state,
		Sections:      append([]SectionHash(nil), loaded.Sections...),
	}
}
