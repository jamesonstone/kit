# Design

## Purpose

Apply these defaults to UI, UX, and product-design work. Project constraints in the Constitution and project-specific guidance refine them. Design requirements should describe desired outcomes and constraints rather than prescribe implementation details.

## Hierarchy and meaning

- Prioritize information by user intent and operational importance. Distinguish primary information, secondary metadata, statuses, and actions through typography, spacing, grouping, and contrast. Unrelated information should not compete for equal prominence.
- Make interactive elements look interactive; informational elements must not imply interaction. Status indicators must express their actual meaning. Distinguish destructive, irreversible, and consequential actions. Never invent functionality to justify a visual treatment.
- Favor recognition over recall. Reduce competing elements and unnecessary decisions. Use progressive disclosure when useful, while keeping information needed to complete the task visible.

## Visual and interaction quality

- Use deliberate typography, spacing, alignment, proportions, and rhythm. Prefer coherent composition and readability over density, decoration, excessive borders, and competing accent colors. Keep consistency, but challenge existing decisions when evidence supports improvement.
- Give clear feedback for actions and understandable loading, empty, error, and success states. Use predictable interactions, appropriate keyboard navigation and focus behavior, and protection against accidental destructive actions.
- Follow [WCAG 2.2 AA](https://www.w3.org/TR/WCAG22/) where applicable. Preserve semantic HTML and accessible names, keyboard access and visible focus, sufficient contrast and target sizes, noncolor meaning, and reduced-motion preferences.
- Preserve hierarchy and readability across supported viewports. Avoid horizontal overflow unless the interaction requires it. Compose responsive layouts deliberately rather than stacking everything indiscriminately or shrinking text to fit.

## Design autonomy and validation

1. Understand user intent and operational context.
2. Diagnose underlying problems; challenge layouts and component boundaries when justified.
3. Choose the simplest effective solution. Preserve behavior and domain semantics, not necessarily presentation.
4. Implement and inspect the rendered result when browser tooling is available.
5. Critique hierarchy, readability, semantic correctness, consistency, accessibility, responsive behavior, interaction feedback, and operational usability.
6. Complete at least one deliberate visual refinement pass for substantive UI changes before completion.

Report rendered-validation limitations honestly. Never claim visual validation without rendering the interface. Avoid arbitrary design scores or subjective preferences presented as objective standards.

## Project guidance

Add project-specific design constraints here. Kit updates unchanged defaults and preserves customized documents during reconciliation.
