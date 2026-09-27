// Command fingerprintgen writes internal/legacy/fingerprints.json from the
// output tree of scripts/harvest-legacy-fingerprints.sh and Kit's rule history.
//
//	go run ./internal/legacy/fingerprintgen <harvest-dir> > internal/legacy/fingerprints.json
package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/legacy"
)

// protected developer-experience files are not migration targets.
var protected = map[string]bool{
	"README.md":                        true,
	".github/pull_request_template.md": true,
}

type set map[string]map[string]bool

func (s set) add(key, value string) {
	if s[key] == nil {
		s[key] = map[string]bool{}
	}
	s[key][value] = true
}

func (s set) sorted() map[string][]string {
	out := make(map[string][]string, len(s))
	for key, values := range s {
		for value := range values {
			out[key] = append(out[key], value)
		}
		sort.Strings(out[key])
	}
	return out
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: fingerprintgen <harvest-dir>")
		os.Exit(2)
	}
	sections, headings, rules, paths := set{}, set{}, set{}, set{}
	err := filepath.WalkDir(os.Args[1], func(file string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(file, ".md") {
			return err
		}
		rel, err := filepath.Rel(os.Args[1], file)
		if err != nil {
			return err
		}
		// <tag>/<variant>/<project path>
		parts := strings.SplitN(filepath.ToSlash(rel), "/", 3)
		if len(parts) != 3 || protected[parts[2]] {
			return nil
		}
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		recordFile(parts[2], string(content), sections, headings, rules, paths)
		return nil
	})
	if err != nil {
		fail(err)
	}
	if err := recordRuleHistory(rules); err != nil {
		fail(err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(legacy.Fingerprints{
		Sections: sections.sorted(),
		Headings: headings.sorted(),
		Rules:    rules.sorted(),
		Paths:    paths.sorted()["generated"],
	}); err != nil {
		fail(err)
	}
}

func recordFile(rel, content string, sections, headings, rules, paths set) {
	if strings.HasPrefix(rel, "docs/references/rules/") {
		rules.add(strings.TrimSuffix(path.Base(rel), ".md"), legacy.RuleFingerprint(content))
		return
	}
	paths.add("generated", rel)
	for _, section := range legacy.Sections(content) {
		sections.add(rel, legacy.SectionFingerprint(rel, section.Text))
		headings.add(rel, section.Key)
	}
}

// recordRuleHistory adds every rule version that ever existed in Kit's history;
// earlier releases installed rules fetched from the default branch.
func recordRuleHistory(rules set) error {
	out, err := exec.Command("git", "rev-list", "--all", "--objects", "--", "docs/references/rules").Output()
	if err != nil {
		return fmt.Errorf("list rule history: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || !strings.HasPrefix(fields[1], "docs/references/rules/") || !strings.HasSuffix(fields[1], ".md") {
			continue
		}
		slug := strings.TrimSuffix(path.Base(fields[1]), ".md")
		if slug == "README" {
			continue
		}
		blob, err := exec.Command("git", "cat-file", "blob", fields[0]).Output()
		if err != nil {
			return fmt.Errorf("read %s: %w", fields[1], err)
		}
		rules.add(slug, legacy.RuleFingerprint(string(blob)))
	}
	return nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
