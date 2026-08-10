package prfix

import "testing"

func TestParseTargetForms(t *testing.T) {
	current := Repository{Owner: "jamesonstone", Name: "kit"}
	tests := []struct {
		value string
		want  Target
	}{
		{"https://github.com/acme/app/pull/42", Target{Repository{Owner: "acme", Name: "app"}, 42}},
		{"[acme/app#42](https://github.com/acme/app/pull/42)", Target{Repository{Owner: "acme", Name: "app"}, 42}},
		{"acme/app#42", Target{Repository{Owner: "acme", Name: "app"}, 42}},
		{"42", Target{current, 42}},
	}
	for _, test := range tests {
		got, err := ParseTarget(test.value, current)
		if err != nil {
			t.Fatalf("ParseTarget(%q): %v", test.value, err)
		}
		if got != test.want {
			t.Fatalf("ParseTarget(%q) = %#v, want %#v", test.value, got, test.want)
		}
	}
}

func TestParseTargetRejectsAmbiguousOrUnsafeValues(t *testing.T) {
	for _, value := range []string{"", "not-a-pr", "acme/../app#2", "0", "github.com/acme/app/issues/2", "https://evilgithub.com/acme/app/pull/2"} {
		if _, err := ParseTarget(value, Repository{}); err == nil {
			t.Errorf("ParseTarget(%q) succeeded", value)
		}
	}
}

func TestParseGitHubRemote(t *testing.T) {
	for _, value := range []string{
		"git@github.com:jamesonstone/kit.git",
		"https://github.com/jamesonstone/kit.git",
		"ssh://git@github.com/jamesonstone/kit.git",
	} {
		got, err := parseRemote(value)
		if err != nil || got.Slug() != "jamesonstone/kit" {
			t.Fatalf("parseRemote(%q) = %#v, %v", value, got, err)
		}
	}
	if _, err := parseRemote("git@evilgithub.com:jamesonstone/kit.git"); err == nil {
		t.Fatal("lookalike GitHub host was accepted")
	}
}
