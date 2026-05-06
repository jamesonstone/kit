package promptlib

import (
	"fmt"
	"strings"
	"unicode"
)

func NormalizeIdentity(noun, verb string) (Identity, error) {
	normalizedNoun, err := NormalizePart(noun, "noun")
	if err != nil {
		return Identity{}, err
	}
	normalizedVerb, err := NormalizePart(verb, "verb")
	if err != nil {
		return Identity{}, err
	}

	return Identity{Noun: normalizedNoun, Verb: normalizedVerb}, nil
}

func NormalizePart(value, fieldName string) (string, error) {
	var sb strings.Builder
	lastHyphen := false

	for _, r := range strings.TrimSpace(value) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(unicode.ToLower(r))
			lastHyphen = false
			continue
		}

		if sb.Len() == 0 || lastHyphen {
			continue
		}
		sb.WriteByte('-')
		lastHyphen = true
	}

	normalized := strings.Trim(sb.String(), "-")
	if normalized == "" {
		return "", fmt.Errorf("prompt %s %q normalizes to empty; use letters or numbers", fieldName, value)
	}

	return normalized, nil
}

func NormalizePrompt(prompt Prompt) (Prompt, error) {
	identity, err := NormalizeIdentity(prompt.Identity.Noun, prompt.Identity.Verb)
	if err != nil {
		return Prompt{}, err
	}
	prompt.Identity = identity
	if strings.TrimSpace(prompt.Content) == "" && prompt.Render == nil {
		return Prompt{}, fmt.Errorf("prompt %q content cannot be empty", prompt.Identity.CommandName())
	}
	return prompt, nil
}
