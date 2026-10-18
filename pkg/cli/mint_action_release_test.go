package cli

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type mintActionTransport func(*http.Request) (*http.Response, error)

func (f mintActionTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestMintActionLatestPublishedAnnotatedRelease(t *testing.T) {
	for _, tc := range []struct {
		name, release, action string
		fail                  bool
	}{
		{"published", `{"tag_name":"v1.0.0","published_at":"2026-10-18T00:00:00Z"}`, "", false},
		{"unpublished", `{"tag_name":"v1.0.0","draft":true}`, "", true},
		{"missing interface", `{"tag_name":"v1.0.0","published_at":"2026-10-18T00:00:00Z"}`, "inputs: {command: {}}\n", true},
		{"transport failure", "", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			previous := mintActionHTTPClient
			t.Cleanup(func() { mintActionHTTPClient = previous })
			commit, tag := strings.Repeat("a", 40), strings.Repeat("b", 40)
			calls := 0
			mintActionHTTPClient = &http.Client{Transport: mintActionTransport(func(req *http.Request) (*http.Response, error) {
				calls++
				if req.URL.Scheme != "https" || req.URL.Host != "api.github.com" || req.Header.Get("Authorization") != "" || req.Method != http.MethodGet {
					t.Fatalf("unexpected target/authority: %s", req.URL)
				}
				if tc.release == "" {
					return nil, fmt.Errorf("private transport diagnostic")
				}
				var body string
				switch req.URL.Path {
				case "/repos/jamesonstone/mint/releases/latest":
					body = tc.release
				case "/repos/jamesonstone/mint/git/ref/tags/v1.0.0":
					body = `{"ref":"refs/tags/v1.0.0","object":{"type":"tag","sha":"` + tag + `"}}`
				case "/repos/jamesonstone/mint/git/tags/" + tag:
					body = `{"sha":"` + tag + `","object":{"type":"commit","sha":"` + commit + `"}}`
				case "/repos/jamesonstone/mint/contents/action.yml":
					if req.URL.Query().Get("ref") != commit || req.Header.Get("Accept") != "application/vnd.github.raw+json" {
						t.Fatal("Action source was not bound to immutable release")
					}
					body = tc.action
					if body == "" {
						body = "inputs:\n"
						for _, input := range []string{"command", "prev-tag", "current-tag", "current-ref", "owner", "repo", "output", "commitish", "github-token"} {
							body += "  " + input + ": {}\n"
						}
					}
				default:
					t.Fatalf("unexpected request %s", req.URL)
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			})}
			got, err := latestMintActionRelease()
			if (err != nil) != tc.fail {
				t.Fatalf("release=%+v error=%v", got, err)
			}
			if !tc.fail && (got.Version != "v1.0.0" || got.SHA != commit || calls != 4) {
				t.Fatalf("not latest immutable source: %+v calls=%d", got, calls)
			}
			if err != nil && strings.Contains(err.Error(), "private transport") {
				t.Fatal("transport diagnostics leaked")
			}
		})
	}
}
