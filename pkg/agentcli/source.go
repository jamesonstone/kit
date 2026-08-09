package agentcli

import (
	"path/filepath"

	"github.com/jamesonstone/kit/internal/registry"
)

func sourceFor(root string, config registry.SourceConfig) registry.Source {
	if config.Path != "" {
		path := config.Path
		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		return registry.LocalSource{Root: path}
	}
	return registry.NewGitHubSource()
}
