package cli

import (
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/instructions"
)

type repoContextDoc struct {
	Label string
	Path  string
	Use   string
}

func loadRepoInstructionContext(projectRoot string) (*config.Config, int) {
	cfg := config.LoadOrDefault(projectRoot)
	return cfg, instructions.DetectVersion(projectRoot, cfg)
}

func existingRepoInstructionDocs(projectRoot string, cfg *config.Config) []repoContextDoc {
	var docs []repoContextDoc
	for _, doc := range instructions.ExistingInstructionDocs(projectRoot, cfg) {
		docs = append(docs, repoContextDoc{
			Label: doc.Label,
			Path:  filepath.Join(projectRoot, filepath.FromSlash(doc.RelativePath)),
			Use:   doc.Use,
		})
	}

	return docs
}

func existingRepoKnowledgeDocs(projectRoot string, cfg *config.Config) []repoContextDoc {
	var docs []repoContextDoc
	for _, doc := range instructions.ExistingSupportDocs(projectRoot, cfg) {
		docs = append(docs, repoContextDoc{
			Label: doc.Label,
			Path:  filepath.Join(projectRoot, filepath.FromSlash(doc.RelativePath)),
			Use:   doc.Use,
		})
	}

	return docs
}

func repoKnowledgeEntrypointPath(projectRoot string, cfg *config.Config) string {
	return instructions.KnowledgeEntrypointPath(projectRoot, cfg)
}

func repoReferencesEntrypointPath(projectRoot string, cfg *config.Config) string {
	return instructions.ReferencesEntrypointPath(projectRoot, cfg)
}
