package promptlib

import "fmt"

func Resolve(prompts []EffectivePrompt, noun, verb string) (EffectivePrompt, error) {
	identity, err := NormalizeIdentity(noun, verb)
	if err != nil {
		return EffectivePrompt{}, err
	}

	for _, prompt := range prompts {
		if prompt.Prompt.Identity == identity {
			return prompt, nil
		}
	}

	return EffectivePrompt{}, NoMatchError{
		Identity: identity,
		Nouns:    Nouns(prompts),
		Verbs:    VerbsForNoun(prompts, identity.Noun),
	}
}

func Nouns(prompts []EffectivePrompt) []string {
	seen := make(map[string]struct{})
	var nouns []string
	for _, prompt := range prompts {
		noun := prompt.Prompt.Identity.Noun
		if _, ok := seen[noun]; ok {
			continue
		}
		seen[noun] = struct{}{}
		nouns = append(nouns, noun)
	}
	return sortedStrings(nouns)
}

func VerbsForNoun(prompts []EffectivePrompt, noun string) []string {
	seen := make(map[string]struct{})
	var verbs []string
	for _, prompt := range prompts {
		if prompt.Prompt.Identity.Noun != noun {
			continue
		}
		verb := prompt.Prompt.Identity.Verb
		if _, ok := seen[verb]; ok {
			continue
		}
		seen[verb] = struct{}{}
		verbs = append(verbs, verb)
	}
	return sortedStrings(verbs)
}

type NoMatchError struct {
	Identity Identity
	Nouns    []string
	Verbs    []string
}

func (e NoMatchError) Error() string {
	if len(e.Verbs) > 0 {
		return fmt.Sprintf(
			"prompt %q not found; nearest verbs for %q: %s",
			e.Identity.CommandName(),
			e.Identity.Noun,
			joinKinds(Suggestions(e.Verbs, e.Identity.Verb, 3)),
		)
	}

	if len(e.Nouns) > 0 {
		return fmt.Sprintf(
			"prompt %q not found; nearest nouns: %s",
			e.Identity.CommandName(),
			joinKinds(Suggestions(e.Nouns, e.Identity.Noun, 3)),
		)
	}

	return fmt.Sprintf("prompt %q not found; no prompts are available", e.Identity.CommandName())
}
