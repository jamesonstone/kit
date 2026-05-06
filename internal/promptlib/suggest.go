package promptlib

import "sort"

func Suggestions(candidates []string, target string, limit int) []string {
	if limit <= 0 {
		return nil
	}

	type scored struct {
		value string
		score int
	}
	scoredCandidates := make([]scored, 0, len(candidates))
	for _, candidate := range candidates {
		scoredCandidates = append(scoredCandidates, scored{
			value: candidate,
			score: levenshtein(candidate, target),
		})
	}

	sort.SliceStable(scoredCandidates, func(i, j int) bool {
		if scoredCandidates[i].score == scoredCandidates[j].score {
			return scoredCandidates[i].value < scoredCandidates[j].value
		}
		return scoredCandidates[i].score < scoredCandidates[j].score
	})

	if len(scoredCandidates) > limit {
		scoredCandidates = scoredCandidates[:limit]
	}

	suggestions := make([]string, 0, len(scoredCandidates))
	for _, candidate := range scoredCandidates {
		suggestions = append(suggestions, candidate.value)
	}
	return suggestions
}

func sortedStrings(values []string) []string {
	sort.Strings(values)
	return values
}

func levenshtein(left, right string) int {
	if left == right {
		return 0
	}
	leftRunes := []rune(left)
	rightRunes := []rune(right)
	if len(leftRunes) == 0 {
		return len(rightRunes)
	}
	if len(rightRunes) == 0 {
		return len(leftRunes)
	}

	previous := make([]int, len(rightRunes)+1)
	current := make([]int, len(rightRunes)+1)
	for j := range previous {
		previous[j] = j
	}
	for i := 1; i <= len(leftRunes); i++ {
		current[0] = i
		for j := 1; j <= len(rightRunes); j++ {
			cost := 0
			if leftRunes[i-1] != rightRunes[j-1] {
				cost = 1
			}
			current[j] = minInt(
				current[j-1]+1,
				previous[j]+1,
				previous[j-1]+cost,
			)
		}
		copy(previous, current)
	}
	return previous[len(rightRunes)]
}

func minInt(values ...int) int {
	minimum := values[0]
	for _, value := range values[1:] {
		if value < minimum {
			minimum = value
		}
	}
	return minimum
}
