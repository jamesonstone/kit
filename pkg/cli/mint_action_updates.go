package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
	"gopkg.in/yaml.v3"
)

// Refresh only parsed Action uses scalars. Never reserialize a project workflow.
// Tags stay tags; full commit pins stay immutable pins. Branches, expressions and
// nonstandard encodings are left for the project to choose explicitly.
func planMintActionUpdates(root string, planned []initRefreshFileChange, release mintActionRelease) ([]initRefreshFileChange, []string, error) {
	workflows, err := mintWorkflows(root)
	if err != nil {
		return nil, nil, err
	}
	indices := map[string]int{}
	for i, change := range planned {
		indices[change.relativePath] = i
	}
	sources := map[string]string{}
	for _, workflow := range workflows {
		sources[workflow.path] = string(workflow.data)
	}
	for _, change := range planned {
		if strings.HasPrefix(change.relativePath, ".github/workflows/") {
			sources[change.relativePath] = change.after
		}
	}
	paths := make([]string, 0, len(sources))
	for path := range sources {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	var notes []string
	for _, path := range paths {
		before := sources[path]
		after, messages, err := refreshMintActionRefs(before, release)
		if err != nil {
			return nil, nil, fmt.Errorf("refresh Mint refs in %s: %w", path, err)
		}
		for _, message := range messages {
			notes = append(notes, "INCOMPLETE Mint ref refresh in "+path+": "+message)
		}
		if after == before {
			continue
		}
		if i, ok := indices[path]; ok {
			planned[i].after = after
			planned[i].strictMintWorkflow = true
			if planned[i].result == instructionFileSkipped {
				planned[i].result = instructionFileUpdated
			}
		} else {
			planned = append(planned, *newMintPipelineFileChange(root, path, before, after, instructionFileUpdated))
		}
	}
	return planned, notes, nil
}

func refreshMintActionRefs(source string, release mintActionRelease) (string, []string, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(source), &doc); err != nil {
		return "", nil, err
	}
	type edit struct {
		start, end int
		value      string
	}
	if mintComplexYAML(&doc) {
		return source, []string{"shared YAML anchors/aliases/merges preserved; inspect Mint refs manually, currentness is unknown"}, nil
	}
	var edits []edit
	var notes []string
	lines := strings.SplitAfter(source, "\n")
	for _, node := range mintActionMappings(&doc) {
		if node.Kind == yaml.MappingNode {
			for i := 0; i+1 < len(node.Content); i += 2 {
				key, value := node.Content[i], node.Content[i+1]
				if key.Value != "uses" || value.Kind != yaml.ScalarNode || !strings.HasPrefix(strings.ToLower(value.Value), "jamesonstone/mint@") {
					continue
				}
				_, ref, _ := strings.Cut(value.Value, "@")
				target := release.Version
				switch {
				case mintActionSHA.MatchString(ref):
					target = release.SHA
				case mintActionVersion.MatchString(ref):
				default:
					notes = append(notes, "custom reference "+ref+" preserved; currentness is unknown")
					continue
				}
				if value.Style != 0 && value.Style != yaml.SingleQuotedStyle && value.Style != yaml.DoubleQuotedStyle {
					notes = append(notes, "nonstandard uses scalar preserved")
					continue
				}
				old := value.Value
				replacement := strings.TrimSuffix(old, ref) + target
				if value.Style == yaml.SingleQuotedStyle {
					old = "'" + old + "'"
					replacement = "'" + replacement + "'"
				}
				if value.Style == yaml.DoubleQuotedStyle {
					old = "\"" + old + "\""
					replacement = "\"" + replacement + "\""
				}
				line := value.Line - 1
				if line < 0 || line >= len(lines) {
					continue
				}
				runes := []rune(lines[line])
				col := value.Column - 1
				if col < 0 || col > len(runes) {
					continue
				}
				start := len(string(runes[:col]))
				if !strings.HasPrefix(lines[line][start:], old) {
					notes = append(notes, "encoded uses scalar preserved")
					continue
				}
				offset := start
				for _, prior := range lines[:line] {
					offset += len(prior)
				}
				edits = append(edits, edit{offset, offset + len(old), replacement})
				// Refresh an exact conventional version annotation; preserve other comments.
				tail := lines[line][start+len(old):]
				trimmed := strings.TrimSpace(tail)
				if strings.HasPrefix(trimmed, "# ") {
					fields := strings.Fields(strings.TrimPrefix(trimmed, "# "))
					if len(fields) > 0 && mintActionVersion.MatchString(fields[0]) {
						pos := strings.Index(tail, fields[0])
						edits = append(edits, edit{offset + len(old) + pos, offset + len(old) + pos + len(fields[0]), release.Version})
					}
				}
			}
		}
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	for _, e := range edits {
		source = source[:e.start] + e.value + source[e.end:]
	}
	return source, notes, nil
}

func planMintActionRefresh(root string, release mintActionRelease) (*initRefreshPlan, error) {
	cfg, err := config.Load(root)
	if err != nil {
		return nil, err
	}
	changes, notes, err := planMintActionUpdates(root, nil, release)
	if err != nil {
		return nil, err
	}
	plan := &initRefreshPlan{cfg: cfg, changes: changes, notes: notes}
	for _, c := range changes {
		if c.result == instructionFileUpdated {
			plan.stats.updated++
		}
	}
	return plan, nil
}

func refreshPlannedMintActions(root string, release *mintActionRelease, changes []initRefreshFileChange, notes []string) ([]initRefreshFileChange, []string, error) {
	if release == nil {
		return changes, notes, nil
	}
	changes, updateNotes, err := planMintActionUpdates(root, changes, *release)
	return changes, append(notes, updateNotes...), err
}
