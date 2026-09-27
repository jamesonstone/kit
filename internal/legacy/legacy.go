// Package legacy recognizes content that earlier Kit releases generated so
// migration can tell Kit-owned files from project edits. It is migration-only:
// nothing here generates legacy output.
//
// fingerprints.json is produced by scripts/harvest-legacy-fingerprints.sh,
// which builds every released Kit version, initializes scratch projects under
// each supported instruction scaffold version, and records section-level
// fingerprints of the generated Markdown plus every rule version in Kit's
// history. Regenerate it only to add releases that predate this package.
package legacy

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"
	"sync"
)

//go:embed fingerprints.json
var fingerprintData []byte

// Fingerprints is the decoded form of fingerprints.json.
type Fingerprints struct {
	// Sections maps a generated path to fingerprints of its `##` sections
	// (the preamble before the first heading counts as one section).
	Sections map[string][]string `json:"sections"`
	// Headings maps a generated path to every section key Kit ever wrote.
	Headings map[string][]string `json:"headings"`
	// Rules maps a rule slug to fingerprints of every shipped version.
	Rules map[string][]string `json:"rules"`
	// Paths lists every Markdown path a released Kit generated at init.
	Paths []string `json:"paths"`
}

type index struct {
	sections map[string]bool
	headings map[string]map[string]bool
	rules    map[string]map[string]bool
	paths    []string
}

var (
	loadOnce sync.Once
	loaded   index
)

func data() index {
	loadOnce.Do(func() {
		var fp Fingerprints
		if err := json.Unmarshal(fingerprintData, &fp); err != nil {
			panic("legacy: invalid embedded fingerprints: " + err.Error())
		}
		loaded = index{
			sections: map[string]bool{},
			headings: map[string]map[string]bool{},
			rules:    map[string]map[string]bool{},
			paths:    fp.Paths,
		}
		for _, hashes := range fp.Sections {
			for _, hash := range hashes {
				loaded.sections[hash] = true
			}
		}
		for path, keys := range fp.Headings {
			loaded.headings[path] = setOf(keys)
		}
		for slug, hashes := range fp.Rules {
			loaded.rules[slug] = setOf(hashes)
		}
	})
	return loaded
}

func setOf(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

// GeneratedPaths lists every Markdown path a released Kit generated at init.
func GeneratedPaths() []string {
	return append([]string(nil), data().paths...)
}

// RuleKnown reports whether content is byte-for-byte (after normalization) a
// version of the rule that Kit shipped under slug.
func RuleKnown(slug, content string) bool {
	return data().rules[slug][RuleFingerprint(content)]
}

// SectionKnown reports whether a section at path matches a Kit-generated one.
func SectionKnown(path string, section Section) bool {
	return data().sections[SectionFingerprint(path, section.Text)]
}

// HeadingKnown reports whether Kit ever wrote a section with key at path.
func HeadingKnown(path, key string) bool {
	return data().headings[path][key]
}

var sectionPattern = regexp.MustCompile(`(?m)^##\s+(.+)$`)

// Section is one `##` section, or the preamble when Key is empty.
type Section struct {
	Key  string // upper-cased heading text; empty for the preamble
	Raw  string // exact source text
	Text string // normalized text used for fingerprints
}

// Sections splits Markdown at `##` headings. The preamble is always first.
func Sections(content string) []Section {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	matches := sectionPattern.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return []Section{{Raw: content, Text: Normalize(content)}}
	}
	sections := []Section{{Raw: content[:matches[0][0]], Text: Normalize(content[:matches[0][0]])}}
	for i, match := range matches {
		end := len(content)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		raw := content[match[0]:end]
		sections = append(sections, Section{
			Key:  strings.ToUpper(strings.TrimSpace(content[match[2]:match[3]])),
			Raw:  raw,
			Text: Normalize(raw),
		})
	}
	return sections
}

// Normalize removes line-ending, trailing-space, and outer blank-line noise.
func Normalize(content string) string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// SectionFingerprint identifies normalized section text at a path.
func SectionFingerprint(path, text string) string {
	return fingerprint(path, text)
}

// RuleFingerprint identifies a rule's content independent of its `status:`
// front-matter value, which projects set per rule.
func RuleFingerprint(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	frontMatter, body := "", content
	if strings.HasPrefix(content, "---\n") {
		if end := strings.Index(content[4:], "\n---"); end >= 0 {
			frontMatter = content[4 : 4+end]
			body = content[4+end+4:]
		}
	}
	var kept []string
	for _, line := range strings.Split(frontMatter, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "status:") {
			kept = append(kept, line)
		}
	}
	return fingerprint(Normalize(strings.Join(kept, "\n")), Normalize(body))
}

func fingerprint(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:8])
}
