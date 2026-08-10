package prfix

import (
	"strings"
	"testing"
)

func TestValidateResolutionRequiresExactHeadAndActiveNamedThreads(t *testing.T) {
	pullRequest := PullRequest{HeadRefOID: "abc"}
	feedback := []Feedback{{ThreadID: "T1"}, {ThreadID: "T2"}}
	got, err := ValidateResolution("abc", pullRequest, []string{"T2", "T1", "T2"}, feedback)
	if err != nil || strings.Join(got, ",") != "T1,T2" {
		t.Fatalf("resolution = %#v, %v", got, err)
	}
	for _, test := range []struct {
		head    string
		threads []string
	}{
		{"old", []string{"T1"}},
		{"abc", nil},
		{"abc", []string{"resolved-or-outdated"}},
	} {
		if _, err := ValidateResolution(test.head, pullRequest, test.threads, feedback); err == nil {
			t.Errorf("unsafe resolution accepted: %#v", test)
		}
	}
}
