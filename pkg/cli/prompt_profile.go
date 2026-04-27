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
		"- Use RLM-style context loading first: identify the immediate frontend decision, load only the relevant feature docs, notes, design materials, and code, then stop once the decision is supported.",
		"- Inspect existing frontend architecture, component libraries, styling systems, design tokens, layout conventions, and interaction patterns before inventing new UI.",
		"- Tailor visual density, hierarchy, content, and interaction patterns to the product domain and audience; operational tools should favor scanability, restrained styling, and efficient repeated use, while games or expressive sites may be more immersive.",
		"- Build or evaluate the actual usable screen, flow, app, tool, game, or site requested; do not substitute marketing placeholders for the primary experience.",
		"- Use familiar UI affordances: icons for tool actions, swatches for color, segmented controls for modes, toggles or checkboxes for binary settings, sliders or inputs for numeric values, menus for option sets, tabs for alternate views, and tooltips for unfamiliar icon-only controls.",
		"- Cover expected UI states and controls for the domain, including loading, empty, error, validation, disabled, hover/focus, interaction, and responsive states when relevant.",
		"- Use relevant visual assets, screenshots, design-file nodes, or design notes when provided, but avoid broad asset loading when they are not needed for the immediate decision.",
		"- For websites, games, product, place, person, or object-focused screens, use relevant real, provided, searched, or generated visual assets when they materially improve inspection or fidelity; avoid purely atmospheric placeholders when users need to inspect the subject.",
		"- Avoid common generated-UI defaults unless explicitly requested: unnecessary landing pages, generic hero sections, nested cards, decorative gradients or blobs, one-note palettes, visible instructional copy, and placeholder-first layouts.",
		"- Define stable responsive dimensions for fixed-format UI such as grids, boards, toolbars, counters, tiles, and media so hover states, labels, icons, loading text, and dynamic content do not shift or resize the layout unexpectedly.",
		"- When a renderer is needed to validate the change, run the app and use browser or screenshot evidence across relevant desktop and mobile viewports before claiming completion.",
		"- Before completion, render or inspect the UI for text overflow, clipping, overlapping elements, responsive layout failures, palette problems, spacing issues, clipped controls, and broken interaction states; revise until the rendered result matches the requirements.",
		"- Stay tool-agnostic: use whatever browser, screenshot, or design-material workflow is available in the current environment unless the user or project dependencies require a specific tool.",
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
