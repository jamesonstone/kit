package promptlib

import (
	"fmt"
	"sort"
)

func Merge(sources ...Source) ([]EffectivePrompt, error) {
	records := make(map[Identity][]SourcePrompt)
	for _, source := range sources {
		seen := make(map[Identity]struct{})
		for _, prompt := range source.Prompts {
			normalized, err := NormalizePrompt(prompt)
			if err != nil {
				return nil, fmt.Errorf("invalid %s prompt in %s: %w", source.Kind, source.Location, err)
			}
			if _, ok := seen[normalized.Identity]; ok {
				return nil, fmt.Errorf(
					"duplicate %s prompt %q in %s after normalization",
					source.Kind,
					normalized.Identity.CommandName(),
					source.Location,
				)
			}
			seen[normalized.Identity] = struct{}{}
			records[normalized.Identity] = append(records[normalized.Identity], SourcePrompt{
				Prompt:   normalized,
				Kind:     source.Kind,
				Location: source.Location,
			})
		}
	}

	effective := make([]EffectivePrompt, 0, len(records))
	for _, prompts := range records {
		sort.SliceStable(prompts, func(i, j int) bool {
			return sourceRank(prompts[i].Kind) > sourceRank(prompts[j].Kind)
		})
		winner := prompts[0]
		effective = append(effective, EffectivePrompt{
			Prompt:   winner.Prompt,
			Kind:     winner.Kind,
			Location: winner.Location,
			Shadowed: prompts[1:],
		})
	}

	SortEffective(effective)
	return effective, nil
}

func SortEffective(prompts []EffectivePrompt) {
	sort.SliceStable(prompts, func(i, j int) bool {
		left := prompts[i].Prompt.Identity
		right := prompts[j].Prompt.Identity
		if left.Noun == right.Noun {
			return left.Verb < right.Verb
		}
		return left.Noun < right.Noun
	})
}

func sourceRank(kind SourceKind) int {
	switch kind {
	case SourceLocal:
		return 3
	case SourceGlobal:
		return 2
	case SourceBuiltin:
		return 1
	default:
		return 0
	}
}

func joinKinds(kinds []string) string {
	if len(kinds) == 0 {
		return ""
	}
	if len(kinds) == 1 {
		return kinds[0]
	}

	result := kinds[0]
	for _, kind := range kinds[1:] {
		result += ", " + kind
	}
	return result
}
