package cli

import (
	"encoding/json"
	"os"
	stdreflect "reflect"
	"regexp"
	"strings"
	"testing"
)

type ste100Fixture struct {
	ID, Category, Request, Before, After, Review string
	Anchors, Expected                            []string
}

// These tests inspect curated fixtures. They do not run a writing engine, prove
// semantic equivalence, or validate the full ASD dictionary or live host behavior.
func TestSTE100ResponseFixtures(t *testing.T) {
	data, err := os.ReadFile("testdata/ste100/responses.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []ste100Fixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	required := map[string]bool{"plan": false, "progress": false, "review": false, "blocked": false, "completion": false, "protected": false, "spanish": false, "tone": false}
	// Match fenced content first, then inline technical tokens and source quotations.
	protected := regexp.MustCompile("(?s)```.*?```|`[^`]+`|\"[^\"]*\"")
	for _, fixture := range fixtures {
		t.Run(fixture.ID, func(t *testing.T) {
			if _, ok := required[fixture.ID]; !ok || required[fixture.ID] {
				t.Fatalf("unknown or duplicate fixture %q", fixture.ID)
			}
			required[fixture.ID] = true
			if fixture.Category == "" || fixture.Review == "" || fixture.Before == fixture.After {
				t.Fatal("fixture needs a distinct rewrite and semantic-review note")
			}
			before := protected.FindAllString(fixture.Before, -1)
			after := protected.FindAllString(fixture.After, -1)
			// Condition-first prose can change token order; content and counts must remain.
			if !stdreflect.DeepEqual(tokenCounts(before), tokenCounts(after)) {
				t.Fatalf("protected content changed: %q -> %q", before, after)
			}
			for _, anchor := range fixture.Anchors {
				if !strings.Contains(fixture.Before, anchor) || !strings.Contains(fixture.After, anchor) {
					t.Errorf("lost evidence/uncertainty anchor %q", anchor)
				}
			}
			for _, expected := range fixture.Expected {
				if !strings.Contains(fixture.After, expected) {
					t.Errorf("override fixture lacks %q", expected)
				}
			}
			if fixture.ID == "spanish" || fixture.ID == "tone" {
				if fixture.Request == "" {
					t.Fatal("override fixture needs an explicit user request")
				}
			}
			if fixture.ID == "plan" {
				steps := strings.Split(fixture.After, "\n")
				if len(steps) != 4 || !strings.HasPrefix(steps[0], "1. When the worktree is ready,") || !strings.HasPrefix(steps[2], "3. If the test fails,") {
					t.Fatal("plan lost ordered single-action steps or controlling conditions")
				}
			}
			if fixture.ID == "review" && !strings.HasPrefix(fixture.After, "If `req.User` is nil,") {
				t.Fatal("finding condition must precede possible result")
			}
		})
	}
	for id, found := range required {
		if !found {
			t.Errorf("missing fixture %s", id)
		}
	}
}

func tokenCounts(tokens []string) map[string]int {
	counts := make(map[string]int)
	for _, token := range tokens {
		counts[token]++
	}
	return counts
}
