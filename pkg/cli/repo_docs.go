package cli

import (
	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/instructions"
)

func repoKnowledgeEntrypointPath(projectRoot string, cfg *config.Config) string {
	return instructions.KnowledgeEntrypointPath(projectRoot, cfg)
}
