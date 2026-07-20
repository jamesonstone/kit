package instructions

import (
	"embed"
	"fmt"
	"strings"
)

const CurrentAgentVersion = "v2"

type agentInstructionVersion struct {
	name string
	path string
}

var agentInstructionVersions = []agentInstructionVersion{
	{name: "v1", path: "versions/v1.md"},
	{name: "v2", path: "versions/v2.md"},
}

//go:embed versions/*.md
var agentInstructionFiles embed.FS

func AgentInstructions(version string) (string, error) {
	if version == "" {
		version = CurrentAgentVersion
	}

	for _, candidate := range agentInstructionVersions {
		if candidate.name != version {
			continue
		}
		content, err := agentInstructionFiles.ReadFile(candidate.path)
		if err != nil {
			return "", fmt.Errorf("read instructions version %q: %w", version, err)
		}
		return string(content), nil
	}

	return "", fmt.Errorf(
		"unsupported instructions version %q; available versions: %s",
		version,
		strings.Join(AgentInstructionVersions(), ", "),
	)
}

func AgentInstructionVersions() []string {
	versions := make([]string, 0, len(agentInstructionVersions))
	for _, version := range agentInstructionVersions {
		versions = append(versions, version.name)
	}
	return versions
}
