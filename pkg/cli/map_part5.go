package cli

import (
	"strings"

	"github.com/jamesonstone/kit/internal/feature"
)

func mapRequiredOptional(style humanOutputStyle, required bool) string {
	label := requiredOptional(required)
	if !style.enabled {
		return label
	}
	if required {
		return whiteBold + label + reset
	}
	return dim + label + reset
}

func mapPresentMissing(style humanOutputStyle, exists bool) string {
	label := presentMissing(exists)
	if !style.enabled {
		return label
	}
	if exists {
		return plan + label + reset
	}
	return dim + label + reset
}

func mapManagedBy(style humanOutputStyle, text string) string {
	if !style.enabled {
		return text
	}
	return dim + text + reset
}

func mapDocKey(style humanOutputStyle, key string) string {
	if !style.enabled {
		return key
	}
	switch key {
	case "B":
		return brainstorm + key + reset
	case "S":
		return spec + key + reset
	case "P":
		return plan + key + reset
	case "T":
		return tasks + key + reset
	case "A":
		return reflect + key + reset
	default:
		return whiteBold + key + reset
	}
}

func mapPresenceMarker(style humanOutputStyle, glyphs mapGlyphs, exists bool) string {
	marker := documentPresenceMarker(glyphs, exists)
	if !style.enabled {
		return marker
	}
	if exists {
		return plan + marker + reset
	}
	return dim + marker + reset
}

func mapEdgeSourceDoc(style humanOutputStyle, name string) string {
	if !style.enabled {
		return name
	}
	switch name {
	case "BRAINSTORM.md":
		return brainstorm + name + reset
	case "SPEC.md":
		return spec + name + reset
	case "PLAN.md":
		return plan + name + reset
	case "TASKS.md":
		return tasks + name + reset
	default:
		return dim + name + reset
	}
}

func mapRelationshipType(style humanOutputStyle, rel string) string {
	if !style.enabled {
		return rel
	}
	return dim + rel + reset
}

func mapReferenceName(style humanOutputStyle, text string) string {
	if !style.enabled {
		return text
	}
	return whiteBold + text + reset
}

func mapReferenceStatus(style humanOutputStyle, text string) string {
	if strings.EqualFold(strings.TrimSpace(text), "stale") {
		if !style.enabled {
			return text
		}
		return implement + text + reset
	}
	if !style.enabled {
		return text
	}
	return dim + text + reset
}

func mapReferenceRelation(style humanOutputStyle, text string) string {
	if !style.enabled {
		return nonEmptyMapValue(text, "relation:unspecified")
	}
	return dim + nonEmptyMapValue(text, "relation:unspecified") + reset
}

func mapReferenceReadPolicy(style humanOutputStyle, text string) string {
	if !style.enabled {
		return nonEmptyMapValue(text, "read:unspecified")
	}
	return dim + nonEmptyMapValue(text, "read:unspecified") + reset
}

func mapUnresolvedLabel(style humanOutputStyle, text string) string {
	if !style.enabled {
		return text
	}
	return implement + text + reset
}

func mapWarningTitle(style humanOutputStyle) string {
	if !style.enabled {
		return "Warnings"
	}
	return implement + "Warnings" + reset
}

func mapWarningLead(style humanOutputStyle, text string) string {
	if !style.enabled {
		return text
	}
	return implement + text + reset
}

func mapMutedIfEnabled(style humanOutputStyle, text string) string {
	if !style.enabled {
		return text
	}
	return dim + text + reset
}

func mapLifecycleLine(style humanOutputStyle) string {
	parts := []string{
		mapDocumentPath(style, feature.MapDocument{Name: "CONSTITUTION.md", Path: "CONSTITUTION.md"}),
		mapDocumentPath(style, feature.MapDocument{Name: "BRAINSTORM.md", Path: "BRAINSTORM.md"}) + mapMutedIfEnabled(style, " (optional)"),
		mapDocumentPath(style, feature.MapDocument{Name: "SPEC.md", Path: "SPEC.md"}),
		mapUnresolvedLabel(style, "clarify"),
		mapUnresolvedLabel(style, "ready"),
		mapUnresolvedLabel(style, "implement"),
		mapUnresolvedLabel(style, "validate"),
		reflect + "reflect" + reset,
		mapUnresolvedLabel(style, "deliver"),
	}
	if !style.enabled {
		parts = []string{"CONSTITUTION.md", "BRAINSTORM.md (optional)", "SPEC.md", "clarify", "ready", "implement", "validate", "reflect", "deliver"}
	}
	return strings.Join(parts, " -> ")
}
