package cli

import (
	"fmt"
	"strings"
)

type promptProfile string

const (
	promptProfileNone     promptProfile = ""
	promptProfileFrontend promptProfile = "frontend"
)

var selectedPromptProfile promptProfile
var selectedPromptProfileExplicit bool

func init() {
	rootCmd.PersistentFlags().Var(
		&selectedPromptProfile,
		"profile",
		"prompt profile for generated agent instructions (supported: frontend)",
	)
}

func (p *promptProfile) Set(value string) error {
	normalized := strings.TrimSpace(value)
	switch promptProfile(normalized) {
	case promptProfileNone, promptProfileFrontend:
		*p = promptProfile(normalized)
		if p == &selectedPromptProfile {
			selectedPromptProfileExplicit = true
		}
		return nil
	default:
		return fmt.Errorf("unsupported profile %q; supported values: frontend", value)
	}
}

func (p *promptProfile) String() string {
	if p == nil {
		return ""
	}
	return string(*p)
}

func (p *promptProfile) Type() string {
	return "profile"
}

func currentPromptProfile() promptProfile {
	return selectedPromptProfile
}

func effectivePromptProfile(featurePath string) promptProfile {
	if selectedPromptProfileExplicit || selectedPromptProfile != promptProfileNone {
		return selectedPromptProfile
	}
	if strings.TrimSpace(featurePath) == "" {
		return promptProfileNone
	}
	if featureHasActiveFrontendProfileDependency(featurePath) {
		return promptProfileFrontend
	}
	return promptProfileNone
}

func appendPromptProfileSuffix(prompt string, profile promptProfile) string {
	if profile != promptProfileFrontend || hasMarkdownHeading(prompt, "## Frontend Profile") {
		return prompt
	}

	trimmed := strings.TrimRight(prompt, "\n")
	suffix := frontendPromptProfileSuffix()
	if trimmed == "" {
		return suffix
	}
	return trimmed + "\n\n" + suffix
}

func frontendPromptProfileSuffix() string {
	return strings.Join([]string{
		"## Frontend Profile",
		"- Load only the relevant feature docs, design materials, and frontend code; inspect existing components, tokens, layout, and interaction patterns before adding UI.",
		"- Match the product domain and audience, and build the requested usable flow rather than a generic landing page or placeholder experience.",
		"- Use familiar controls and relevant assets; cover loading, empty, error, validation, disabled, focus, interaction, and responsive states that matter to the flow.",
		"- Avoid generated-UI defaults such as needless hero sections, nested cards, decorative blobs, one-note palettes, and instructional filler unless requested.",
		"- Keep fixed-format UI dimensions stable. Run the app when rendering matters and inspect browser or screenshot evidence at relevant desktop and mobile sizes for overflow, clipping, overlap, spacing, palette, and broken interactions.",
		"- Stay tool-agnostic unless the repository or user requires a specific browser, design, or screenshot workflow.",
	}, "\n")
}

func hasMarkdownHeading(content, heading string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == heading {
			return true
		}
	}
	return false
}
