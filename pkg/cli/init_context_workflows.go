package cli

import (
	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/templates"
)

func planRefreshContextWorkflowFiles(
	projectRoot string,
	opts initRefreshOptions,
	targets map[string]bool,
) ([]initRefreshFileChange, error) {
	artifacts, err := templates.ContextWorkflowArtifacts()
	if err != nil {
		return nil, err
	}
	var changes []initRefreshFileChange
	for _, artifact := range artifacts {
		if !initRefreshTargetMatches(targets, artifact.Path) {
			continue
		}
		change, err := planRefreshInitScaffoldFile(
			projectRoot,
			opts,
			targets,
			artifact.Path,
			artifact.Content,
			false,
			true,
		)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	return changes, nil
}

func contextWorkflowPaths() ([]string, error) {
	artifacts, err := templates.ContextWorkflowArtifacts()
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		paths = append(paths, artifact.Path)
	}
	return paths, nil
}

func addContextWorkflowTargets(known map[string]bool) error {
	paths, err := contextWorkflowPaths()
	if err != nil {
		return err
	}
	for _, path := range paths {
		known[path] = true
	}
	return nil
}

func appendContextWorkflowDeliveryPaths(paths []string, _ *config.Config) []string {
	workflowPaths, err := contextWorkflowPaths()
	if err != nil {
		return paths
	}
	return append(paths, workflowPaths...)
}
