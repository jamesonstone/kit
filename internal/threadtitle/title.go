// Package threadtitle formats agent-resolved conversation ownership without
// accessing or changing host session state.
package threadtitle

import (
	"fmt"
	"strings"
	"unicode"
)

// Input contains semantic fields resolved by the active agent. Accurate is the
// agent's judgment that Current still describes the work this conversation owns.
type Input struct {
	Scope     string
	Domain    string
	Objective string
	Current   string
	Accurate  bool
}

// Decision describes the title a host adapter should retain or apply. Title is
// the existing title when Changed is false. Reason is a stable machine value.
type Decision struct {
	Title   string `json:"title"`
	Changed bool   `json:"changed"`
	Reason  string `json:"reason"`
}

// Resolve validates and formats a candidate, suppressing changes when ownership
// remains accurate or only capitalization, whitespace, or edge punctuation differ.
// It never truncates semantic fields or applies the result to a conversation.
func Resolve(in Input) (Decision, error) {
	fields := []*string{&in.Scope, &in.Domain, &in.Objective}
	names := []string{"scope", "domain", "objective"}
	for i, field := range fields {
		if err := validate(names[i], *field); err != nil {
			return Decision{}, err
		}
		*field = strings.Join(strings.Fields(*field), " ")
	}
	if strings.ContainsAny(in.Scope, "+,") {
		return Decision{}, fmt.Errorf("scope must name one repository, project, or program, not a repository list")
	}
	title := fmt.Sprintf("[%s] %s / %s", in.Scope, in.Domain, in.Objective)
	if strings.TrimSpace(in.Current) == "" {
		return Decision{Title: title, Changed: true, Reason: "initial_title"}, nil
	}
	if in.Accurate {
		return Decision{Title: in.Current, Reason: "ownership_unchanged"}, nil
	}
	if equivalent(in.Current, title) {
		return Decision{Title: in.Current, Reason: "cosmetic_only"}, nil
	}
	return Decision{Title: title, Changed: true, Reason: "ownership_changed"}, nil
}

func validate(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must not be blank", name)
	}
	for _, r := range value {
		if (unicode.IsControl(r) && r != '\t') || r == '\u2028' || r == '\u2029' {
			return fmt.Errorf("%s must be a single line without control characters", name)
		}
	}
	if strings.ContainsAny(value, "[]/") {
		return fmt.Errorf("%s must not contain title delimiters [, ], or /", name)
	}
	return nil
}

func equivalent(current, candidate string) bool {
	// Compare fields separately: punctuation normalization must never erase the
	// structural boundary between scope, domain and objective.
	split := func(title string) []string {
		title = strings.TrimSpace(title)
		if !strings.HasPrefix(title, "[") {
			return nil
		}
		end := strings.Index(title, "]")
		if end < 0 {
			return nil
		}
		rest := strings.Split(title[end+1:], "/")
		if len(rest) != 2 {
			return nil
		}
		return []string{title[1:end], rest[0], rest[1]}
	}
	a, b := split(current), split(candidate)
	if len(a) != 3 || len(b) != 3 {
		return false
	}
	for i := range a {
		if normalize(a[i]) != normalize(b[i]) {
			return false
		}
	}
	return true
}

func normalize(value string) string {
	words := strings.Fields(strings.ToLower(value))
	for i := range words {
		words[i] = strings.Trim(words[i], ".,:;!?\"'“”‘’")
	}
	return strings.Join(words, " ")
}
