package instructions

import (
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

const UnknownVersion = 0

const (
	AgentsMDPath            = "AGENTS.md"
	ClaudeMDPath            = "CLAUDE.md"
	CopilotInstructionsPath = ".github/copilot-instructions.md"
)

type Doc struct {
	Label        string
	RelativePath string
	Use          string
	Required     bool
	ManagedBy    string
}

func InstructionRelativePaths(cfg *config.Config) []string {
	if cfg == nil {
		cfg = config.Default()
	}

	files := make([]string, 0, len(cfg.Agents)+1)
	for _, file := range cfg.Agents {
		files = appendUnique(files, file)
	}
	files = appendUnique(files, CopilotInstructionsPath)
	return files
}

func DetectVersion(projectRoot string, cfg *config.Config) int {
	if cfg != nil && config.IsInstructionScaffoldVersionSupported(cfg.InstructionScaffoldVersion) {
		return cfg.InstructionScaffoldVersion
	}

	for _, doc := range SupportDocs(config.InstructionScaffoldVersionTOC) {
		if document.Exists(filepath.Join(projectRoot, filepath.FromSlash(doc.RelativePath))) {
			return config.InstructionScaffoldVersionTOC
		}
	}

	for _, relativePath := range InstructionRelativePaths(cfg) {
		if document.Exists(filepath.Join(projectRoot, filepath.FromSlash(relativePath))) {
			return config.InstructionScaffoldVersionVerbose
		}
	}

	return UnknownVersion
}

func InstructionDocs(cfg *config.Config, version int) []Doc {
	use := "repository instruction contract; keep aligned with canonical docs"
	if version == config.InstructionScaffoldVersionTOC {
		use = "thin instruction entrypoint; keep aligned with the repo-local docs tree"
	}

	docs := make([]Doc, 0, len(InstructionRelativePaths(cfg)))
	for _, relativePath := range InstructionRelativePaths(cfg) {
		docs = append(docs, Doc{
			Label:        LabelForPath(relativePath),
			RelativePath: relativePath,
			Use:          use,
			Required:     true,
			ManagedBy:    "kit scaffold-agents",
		})
	}

	return docs
}

func SupportDocs(version int) []Doc {
	if version != config.InstructionScaffoldVersionTOC {
		return nil
	}

	return []Doc{
		{
			Label:        "AGENTS DOCS",
			RelativePath: "docs/agents/README.md",
			Use:          "repo-local entrypoint and read-order guide",
			Required:     true,
			ManagedBy:    "kit scaffold-agents",
		},
		{
			Label:        "WORKFLOWS",
			RelativePath: "docs/agents/WORKFLOWS.md",
			Use:          "spec-driven versus ad hoc routing",
			Required:     true,
			ManagedBy:    "kit scaffold-agents",
		},
		{
			Label:        "RLM",
			RelativePath: "docs/agents/RLM.md",
			Use:          "repository-scale discovery and progressive disclosure",
			Required:     true,
			ManagedBy:    "kit scaffold-agents",
		},
		{
			Label:        "TOOLING",
			RelativePath: "docs/agents/TOOLING.md",
			Use:          "skills, dispatch, worktrees, and secondary globals",
			Required:     true,
			ManagedBy:    "kit scaffold-agents",
		},
		{
			Label:        "GUARDRAILS",
			RelativePath: "docs/agents/GUARDRAILS.md",
			Use:          "hard constraints and completion bar",
			Required:     true,
			ManagedBy:    "kit scaffold-agents",
		},
		{
			Label:        "REFERENCES",
			RelativePath: "docs/references/README.md",
			Use:          "repo-wide references index",
			Required:     true,
			ManagedBy:    "kit scaffold-agents",
		},
		{
			Label:        "TESTING REFERENCE",
			RelativePath: "docs/references/testing.md",
			Use:          "durable repo-wide testing guidance",
			Required:     true,
			ManagedBy:    "kit scaffold-agents",
		},
		{
			Label:        "TOOLING REFERENCE",
			RelativePath: "docs/references/tooling.md",
			Use:          "durable repo-wide tooling guidance",
			Required:     true,
			ManagedBy:    "kit scaffold-agents",
		},
		{
			Label:        "EXTERNAL SYSTEMS",
			RelativePath: "docs/references/external-systems.md",
			Use:          "durable external-system notes",
			Required:     true,
			ManagedBy:    "kit scaffold-agents",
		},
	}
}

func ExistingInstructionDocs(projectRoot string, cfg *config.Config) []Doc {
	version := DetectVersion(projectRoot, cfg)
	if version == UnknownVersion {
		return nil
	}

	return existingDocs(projectRoot, InstructionDocs(cfg, version))
}

func ExistingSupportDocs(projectRoot string, cfg *config.Config) []Doc {
	if DetectVersion(projectRoot, cfg) != config.InstructionScaffoldVersionTOC {
		return nil
	}

	return existingDocs(projectRoot, SupportDocs(config.InstructionScaffoldVersionTOC))
}

func KnowledgeEntrypointPath(projectRoot string, cfg *config.Config) string {
	return existingDocPathByLabel(projectRoot, cfg, "AGENTS DOCS")
}

func ReferencesEntrypointPath(projectRoot string, cfg *config.Config) string {
	return existingDocPathByLabel(projectRoot, cfg, "REFERENCES")
}

func LabelForPath(path string) string {
	switch filepath.Base(path) {
	case "AGENTS.md":
		return "AGENTS"
	case "CLAUDE.md":
		return "CLAUDE"
	case "copilot-instructions.md":
		return "COPILOT"
	default:
		return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
}

func appendUnique(items []string, value string) []string {
	for _, existing := range items {
		if existing == value {
			return items
		}
	}

	return append(items, value)
}

func existingDocs(projectRoot string, docs []Doc) []Doc {
	var existing []Doc
	for _, doc := range docs {
		if document.Exists(filepath.Join(projectRoot, filepath.FromSlash(doc.RelativePath))) {
			existing = append(existing, doc)
		}
	}

	return existing
}

func existingDocPathByLabel(projectRoot string, cfg *config.Config, label string) string {
	for _, doc := range ExistingSupportDocs(projectRoot, cfg) {
		if doc.Label == label {
			return filepath.Join(projectRoot, filepath.FromSlash(doc.RelativePath))
		}
	}

	return ""
}
