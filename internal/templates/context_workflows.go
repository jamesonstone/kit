package templates

import (
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed context_workflows/*.md
var contextWorkflowFiles embed.FS

type ContextWorkflowArtifact struct {
	Slug    string
	Path    string
	Content string
}

func ContextWorkflowArtifacts() ([]ContextWorkflowArtifact, error) {
	entries, err := contextWorkflowFiles.ReadDir("context_workflows")
	if err != nil {
		return nil, fmt.Errorf("read embedded context workflows: %w", err)
	}
	artifacts := make([]ContextWorkflowArtifact, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		content, err := contextWorkflowFiles.ReadFile("context_workflows/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read embedded workflow %s: %w", entry.Name(), err)
		}
		slug := strings.TrimSuffix(entry.Name(), ".md")
		artifacts = append(artifacts, ContextWorkflowArtifact{
			Slug: slug, Path: "docs/references/workflows/" + entry.Name(), Content: string(content),
		})
	}
	sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].Slug < artifacts[j].Slug })
	return artifacts, nil
}
