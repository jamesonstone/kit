package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/term"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

func buildRegistrySelectorEntries(projectRoot string, registry []registryRuleset) ([]registrySelectorEntry, error) {
	sort.SliceStable(registry, func(i, j int) bool {
		return registry[i].Slug < registry[j].Slug
	})
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	entries := make([]registrySelectorEntry, 0, len(registry))
	seen := map[string]bool{}
	for _, item := range registry {
		if item.Slug == "" {
			return nil, fmt.Errorf("registry ruleset has empty slug")
		}
		if seen[item.Slug] {
			return nil, fmt.Errorf("registry ruleset %q is duplicated", item.Slug)
		}
		seen[item.Slug] = true

		entry := registrySelectorEntry{Registry: item}
		localPath := rulesetPath(projectRoot, item.Slug)
		if document.Exists(localPath) {
			localContent, err := os.ReadFile(localPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read %s: %w", rulesetTarget(item.Slug), err)
			}
			local := parseRuleset(string(localContent), localPath)
			if issues := validateRulesetDocument(local, item.Slug); len(issues) > 0 {
				return nil, fmt.Errorf("local ruleset %s is invalid: %s", rulesetTarget(item.Slug), strings.Join(issues, "; "))
			}
			entry.Local = &local
			entry.LocalContent = string(localContent)
			entry.Installed = true
			entry.CurrentActive = local.Metadata.Status == document.ReferenceStatusActive
			entry.DesiredActive = entry.CurrentActive
			entry.Modified = localRulesetModified(entry.LocalContent, item.Content, item.Metadata.Status)
			entry.RegistryState = selectorRegistryState(cfg, item, entry.LocalContent, entry.Modified)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func selectorRegistryState(cfg *config.Config, item registryRuleset, localContent string, modified bool) string {
	if strings.TrimSpace(localContent) == "" {
		return "registry"
	}
	artifact, tracked := rulesetRegistryState(cfg, item.Slug)
	localHash, err := normalizedRulesetContentHash(localContent, item.Metadata.Status)
	if err == nil && localHash == item.NormalizedHash {
		return registryArtifactStateManaged
	}
	if tracked {
		switch artifact.State {
		case registryArtifactStateConflict:
			return registryArtifactStateConflict
		case registryArtifactStateLocalCustom:
			return registryArtifactStateLocalCustom
		case registryArtifactStateManaged:
			if artifact.InstalledHash != "" && artifact.InstalledHash != item.NormalizedHash {
				return "update-available"
			}
			return registryArtifactStateManaged
		}
	}
	if modified {
		return registryArtifactStateLocalCustom
	}
	return "local"
}

func localRulesetModified(localContent, registryContent, registryStatus string) bool {
	normalizedLocal, err := setRulesetStatus(localContent, registryStatus)
	if err != nil {
		normalizedLocal = localContent
	}
	return strings.TrimSpace(normalizedLocal) != strings.TrimSpace(registryContent)
}

func selectRegistryRulesets(in io.Reader, out io.Writer, entries []registrySelectorEntry) error {
	if len(entries) == 0 {
		return nil
	}
	if inFile, ok := in.(*os.File); ok && terminalWriterCheck(out) && term.IsTerminal(int(inFile.Fd())) {
		return selectRegistryRulesetsRaw(inFile, out, entries)
	}
	return selectRegistryRulesetsLine(in, out, entries)
}

func selectRegistryRulesetsLine(in io.Reader, out io.Writer, entries []registrySelectorEntry) error {
	renderRegistryRulesetSelector(out, entries, -1)
	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}
	if _, err := fmt.Fprint(out, "Enter numbers separated by spaces to toggle, or press Enter to apply: "); err != nil {
		return err
	}
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read ruleset selection: %w", err)
	}
	for _, token := range strings.Fields(line) {
		index, err := strconv.Atoi(token)
		if err != nil || index < 1 || index > len(entries) {
			return fmt.Errorf("invalid ruleset selection: %s", token)
		}
		toggleRegistrySelectorEntry(&entries[index-1])
	}
	return nil
}

func selectRegistryRulesetsRaw(in *os.File, out io.Writer, entries []registrySelectorEntry) error {
	state, err := term.MakeRaw(int(in.Fd()))
	if err != nil {
		return fmt.Errorf("failed to enter raw terminal mode: %w", err)
	}
	defer term.Restore(int(in.Fd()), state)

	fd := in.Fd()
	if fdProvider, ok := out.(interface{ Fd() uintptr }); ok {
		fd = fdProvider.Fd()
	}
	rawOut := &rawTerminalLineWriter{
		writer: out,
		fd:     fd,
	}
	reader := bufio.NewReader(in)
	cursor := 0
	for {
		renderRegistryRulesetSelector(rawOut, entries, cursor)
		key, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("failed to read ruleset selector input: %w", err)
		}
		switch key {
		case 3:
			return fmt.Errorf("ruleset selection cancelled")
		case 'q', 'Q':
			return fmt.Errorf("ruleset selection cancelled")
		case '\t':
			cursor = moveRegistrySelectorCursor(cursor, len(entries), 1, true)
		case ' ', 'x', 'X':
			toggleRegistrySelectorEntry(&entries[cursor])
		case 'v', 'V', '?':
			renderRegistryRulesetPreview(rawOut, entries[cursor])
			if _, err := reader.ReadByte(); err != nil {
				return fmt.Errorf("failed to read ruleset preview input: %w", err)
			}
		case '\r', '\n':
			_, _ = fmt.Fprint(rawOut, "\n")
			return nil
		case 'j', 'J':
			cursor = moveRegistrySelectorCursor(cursor, len(entries), 1, false)
		case 'k', 'K':
			cursor = moveRegistrySelectorCursor(cursor, len(entries), -1, false)
		case 27:
			second, err := reader.ReadByte()
			if err != nil {
				return err
			}
			third, err := reader.ReadByte()
			if err != nil {
				return err
			}
			if second != '[' {
				continue
			}
			switch third {
			case 'A':
				cursor = moveRegistrySelectorCursor(cursor, len(entries), -1, false)
			case 'B':
				cursor = moveRegistrySelectorCursor(cursor, len(entries), 1, false)
			case 'Z':
				cursor = moveRegistrySelectorCursor(cursor, len(entries), -1, true)
			}
		}
	}
}

func moveRegistrySelectorCursor(cursor, count, delta int, wrap bool) int {
	if count <= 0 {
		return 0
	}
	next := cursor + delta
	if next < 0 {
		if wrap {
			return count - 1
		}
		return 0
	}
	if next >= count {
		if wrap {
			return 0
		}
		return count - 1
	}
	return next
}

type rawTerminalLineWriter struct {
	writer     io.Writer
	fd         uintptr
	previousCR bool
}

func (w *rawTerminalLineWriter) Fd() uintptr {
	return w.fd
}

func (w *rawTerminalLineWriter) Write(p []byte) (int, error) {
	translated := make([]byte, 0, len(p)+8)
	for _, b := range p {
		if b == '\n' && !w.previousCR {
			translated = append(translated, '\r')
		}
		translated = append(translated, b)
		w.previousCR = b == '\r'
	}

	n, err := w.writer.Write(translated)
	if err != nil {
		return 0, err
	}
	if n != len(translated) {
		return 0, io.ErrShortWrite
	}
	return len(p), nil
}

func toggleRegistrySelectorEntry(entry *registrySelectorEntry) {
	entry.DesiredActive = !entry.DesiredActive
	entry.Touched = true
}
