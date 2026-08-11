package cli

import (
	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/instructions"
)

func repoKnowledgeEntrypointPath(projectRoot string, cfg *config.Config) string {
	return instructions.KnowledgeEntrypointPath(projectRoot, cfg)
}
