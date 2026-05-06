package promptlib

import "testing"

func TestNormalizeIdentityUsesKebabCase(t *testing.T) {
	got, err := NormalizeIdentity("Coding Agent", "SHORT_prompt!!")
	if err != nil {
		t.Fatalf("NormalizeIdentity() error = %v", err)
	}

	want := Identity{Noun: "coding-agent", Verb: "short-prompt"}
	if got != want {
		t.Fatalf("NormalizeIdentity() = %#v, want %#v", got, want)
	}
}

func TestNormalizePartRejectsEmptyValues(t *testing.T) {
	if _, err := NormalizePart("!!!", "noun"); err == nil {
		t.Fatal("expected empty normalized noun to fail")
	}
}

func TestNormalizePromptRejectsEmptyStaticContent(t *testing.T) {
	_, err := NormalizePrompt(Prompt{
		Identity: Identity{Noun: "custom", Verb: "review"},
		Content:  "   ",
	})
	if err == nil {
		t.Fatal("expected empty static prompt content to fail")
	}
}

func TestNormalizePromptAllowsDynamicPromptWithoutStaticContent(t *testing.T) {
	_, err := NormalizePrompt(Prompt{
		Identity: Identity{Noun: "workflow", Verb: "plan"},
		Render: func() (string, error) {
			return "rendered", nil
		},
	})
	if err != nil {
		t.Fatalf("NormalizePrompt() error = %v", err)
	}
}
