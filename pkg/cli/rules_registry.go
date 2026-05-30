package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/jamesonstone/kit/internal/document"
)

const (
	rulesetRegistryOwner  = "jamesonstone"
	rulesetRegistryRepo   = "kit"
	rulesetRegistryBranch = "main"
	rulesetRegistryAPIURL = "https://api.github.com/repos/jamesonstone/kit/contents/docs/references/rules?ref=main"
	inactiveRulesetStatus = document.ReferenceStatusOptional
)

type rulesetRegistryFetchFunc func(context.Context) ([]registryRuleset, error)

var rulesetRegistryFetcher rulesetRegistryFetchFunc = fetchGitHubRulesetRegistry

type registryRuleset struct {
	Slug     string
	Content  string
	Metadata rulesetMetadata
}

type registrySelectorEntry struct {
	Registry      registryRuleset
	Local         *rulesetDocument
	LocalContent  string
	Installed     bool
	Modified      bool
	CurrentActive bool
	DesiredActive bool
	Touched       bool
}

type registrySelectorSummary struct {
	Imported    int
	Activated   int
	Deactivated int
	Unchanged   int
}

type githubContentEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

func runRulesAddRegistrySelector(cmd interface {
	InOrStdin() io.Reader
	OutOrStdout() io.Writer
}, projectRoot string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	registry, err := rulesetRegistryFetcher(ctx)
	if err != nil {
		return err
	}
	if len(registry) == 0 {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "No registry rulesets found.")
		return err
	}

	entries, err := buildRegistrySelectorEntries(projectRoot, registry)
	if err != nil {
		return err
	}
	if err := selectRegistryRulesets(cmd.InOrStdin(), cmd.OutOrStdout(), entries); err != nil {
		return err
	}

	summary, err := applyRegistryRulesetSelection(projectRoot, entries)
	if err != nil {
		return err
	}
	return printRegistryRulesetSummary(cmd.OutOrStdout(), summary)
}

func fetchGitHubRulesetRegistry(ctx context.Context) ([]registryRuleset, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rulesetRegistryAPIURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ruleset registry from GitHub: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch ruleset registry from GitHub: %s", resp.Status)
	}

	var entries []githubContentEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub ruleset registry: %w", err)
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	var rulesets []registryRuleset
	for _, entry := range entries {
		if entry.Type != "file" || !strings.HasSuffix(entry.Name, ".md") {
			continue
		}
		if strings.TrimSpace(entry.DownloadURL) == "" {
			return nil, fmt.Errorf("registry ruleset %s has no download URL", entry.Name)
		}
		ruleset, err := fetchGitHubRegistryRuleset(ctx, entry)
		if err != nil {
			return nil, err
		}
		rulesets = append(rulesets, ruleset)
	}
	return rulesets, nil
}

func fetchGitHubRegistryRuleset(ctx context.Context, entry githubContentEntry) (registryRuleset, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, entry.DownloadURL, nil)
	if err != nil {
		return registryRuleset{}, err
	}
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return registryRuleset{}, fmt.Errorf("failed to fetch registry ruleset %s: %w", entry.Name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return registryRuleset{}, fmt.Errorf("failed to fetch registry ruleset %s: %s", entry.Name, resp.Status)
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return registryRuleset{}, fmt.Errorf("failed to read registry ruleset %s: %w", entry.Name, err)
	}

	slug := strings.TrimSuffix(entry.Name, ".md")
	parsed := parseRuleset(string(content), entry.Name)
	if issues := validateRulesetDocument(parsed, slug); len(issues) > 0 {
		return registryRuleset{}, fmt.Errorf("registry ruleset %s is invalid: %s", entry.Name, strings.Join(issues, "; "))
	}
	return registryRuleset{
		Slug:     parsed.Metadata.Slug,
		Content:  string(content),
		Metadata: parsed.Metadata,
	}, nil
}

func buildRegistrySelectorEntries(projectRoot string, registry []registryRuleset) ([]registrySelectorEntry, error) {
	sort.SliceStable(registry, func(i, j int) bool {
		return registry[i].Slug < registry[j].Slug
	})

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
		}
		entries = append(entries, entry)
	}
	return entries, nil
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

	reader := bufio.NewReader(in)
	cursor := 0
	for {
		renderRegistryRulesetSelector(out, entries, cursor)
		key, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("failed to read ruleset selector input: %w", err)
		}
		switch key {
		case 3:
			return fmt.Errorf("ruleset selection cancelled")
		case 'q', 'Q':
			return fmt.Errorf("ruleset selection cancelled")
		case ' ', 'x', 'X':
			toggleRegistrySelectorEntry(&entries[cursor])
		case 'v', 'V', '?':
			renderRegistryRulesetPreview(out, entries[cursor])
			if _, err := reader.ReadByte(); err != nil {
				return fmt.Errorf("failed to read ruleset preview input: %w", err)
			}
		case '\r', '\n':
			_, _ = fmt.Fprint(out, "\n")
			return nil
		case 'j', 'J':
			if cursor < len(entries)-1 {
				cursor++
			}
		case 'k', 'K':
			if cursor > 0 {
				cursor--
			}
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
				if cursor > 0 {
					cursor--
				}
			case 'B':
				if cursor < len(entries)-1 {
					cursor++
				}
			}
		}
	}
}

func toggleRegistrySelectorEntry(entry *registrySelectorEntry) {
	entry.DesiredActive = !entry.DesiredActive
	entry.Touched = true
}

func renderRegistryRulesetSelector(out io.Writer, entries []registrySelectorEntry, cursor int) {
	style := styleForWriter(out)
	if cursor >= 0 {
		_, _ = fmt.Fprint(out, "\x1b[H\x1b[2J")
	}
	_, _ = fmt.Fprintln(out, style.selectionTitle("Select registry rulesets"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.muted("Registry: "+rulesetRegistrySourceDescription()))
	if cursor >= 0 {
		_, _ = fmt.Fprintln(out, style.muted("Use Up/Down or j/k to move, Space to toggle, v to view, Enter to apply, q to cancel."))
	} else {
		_, _ = fmt.Fprintln(out, style.muted("Type rule numbers to toggle active/inactive state, then press Enter to apply. Use `kit rules view <slug>` to preview full text."))
	}
	_, _ = fmt.Fprintln(out)
	for i := range entries {
		prefix := " "
		if i == cursor {
			prefix = ">"
		}
		state := formatRegistrySelectorState(style, entries[i])
		modified := ""
		if entries[i].Modified {
			modified = " " + formatRulesetStateToken(style, "MODIFIED", constitution)
		}
		checkbox := "[ ]"
		if entries[i].DesiredActive {
			checkbox = "[x]"
		}
		description := truncateRulesetDescription(selectorRulesetDescription(entries[i]), 88)
		_, _ = fmt.Fprintf(out, "%s [%d] %s %-28s %s%s  %s\n", prefix, i+1, checkbox, entries[i].Registry.Slug, state, modified, description)
	}
}

func renderRegistryRulesetPreview(out io.Writer, entry registrySelectorEntry) {
	style := styleForWriter(out)
	content := entry.Registry.Content
	source := rulesetRegistryRulesetURL(entry.Registry.Slug)
	if entry.Installed {
		content = entry.LocalContent
		source = rulesetTarget(entry.Registry.Slug)
	}
	_, _ = fmt.Fprint(out, "\x1b[H\x1b[2J")
	_, _ = fmt.Fprintln(out, style.selectionTitle("Preview: "+entry.Registry.Slug))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.muted("Source: "+source))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprint(out, ensureTrailingNewline(content))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.muted("Press any key to return."))
}

func selectorRulesetDescription(entry registrySelectorEntry) string {
	if entry.Local != nil {
		if description := strings.TrimSpace(entry.Local.Metadata.Description); description != "" {
			return description
		}
	}
	if description := strings.TrimSpace(entry.Registry.Metadata.Description); description != "" {
		return description
	}
	return "No description provided."
}

func truncateRulesetDescription(description string, maxLength int) string {
	description = strings.Join(strings.Fields(description), " ")
	if maxLength <= 0 || len(description) <= maxLength {
		return description
	}
	if maxLength <= 3 {
		return description[:maxLength]
	}
	return strings.TrimSpace(description[:maxLength-3]) + "..."
}

func formatRegistrySelectorState(style humanOutputStyle, entry registrySelectorEntry) string {
	switch {
	case entry.DesiredActive:
		return formatRulesetStateToken(style, "ACTIVE", plan)
	case entry.Installed:
		return formatRulesetStateToken(style, "INACTIVE", implement)
	default:
		return formatRulesetStateToken(style, "AVAILABLE", dim)
	}
}

func formatRulesetStateToken(style humanOutputStyle, label string, color string) string {
	if !style.enabled {
		return label
	}
	return color + whiteBold + label + reset
}

func applyRegistryRulesetSelection(projectRoot string, entries []registrySelectorEntry) (registrySelectorSummary, error) {
	var summary registrySelectorSummary
	for _, entry := range entries {
		if !entry.Installed && !entry.DesiredActive {
			summary.Unchanged++
			continue
		}
		if entry.Installed && entry.DesiredActive == entry.CurrentActive {
			summary.Unchanged++
			continue
		}

		targetStatus := inactiveRulesetStatus
		if entry.DesiredActive {
			targetStatus = document.ReferenceStatusActive
		}

		content := entry.Registry.Content
		if entry.Installed {
			content = entry.LocalContent
		}
		updated, err := setRulesetStatus(content, targetStatus)
		if err != nil {
			return summary, fmt.Errorf("failed to update ruleset %s status: %w", entry.Registry.Slug, err)
		}
		path := rulesetPath(projectRoot, entry.Registry.Slug)
		if err := document.Write(path, updated); err != nil {
			return summary, fmt.Errorf("failed to write %s: %w", rulesetTarget(entry.Registry.Slug), err)
		}

		switch {
		case !entry.Installed && entry.DesiredActive:
			summary.Imported++
		case entry.DesiredActive:
			summary.Activated++
		default:
			summary.Deactivated++
		}
	}
	return summary, nil
}

func printRegistryRulesetSummary(out io.Writer, summary registrySelectorSummary) error {
	if summary.Imported == 0 && summary.Activated == 0 && summary.Deactivated == 0 {
		_, err := fmt.Fprintln(out, "No ruleset changes selected.")
		return err
	}
	_, err := fmt.Fprintf(
		out,
		"Rulesets updated. Imported: %d, Activated: %d, Deactivated: %d, Unchanged: %d\n",
		summary.Imported,
		summary.Activated,
		summary.Deactivated,
		summary.Unchanged,
	)
	return err
}

func setRulesetStatus(content, status string) (string, error) {
	if !validRulesetStatus(status) {
		return "", fmt.Errorf("invalid ruleset status %q", status)
	}
	raw, body, err := splitRulesetFrontMatter(content)
	if err != nil {
		return "", err
	}
	lines := strings.Split(raw, "\n")
	changed := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "status:") {
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			lines[i] = indent + "status: " + status
			changed = true
			break
		}
	}
	if !changed {
		return "", fmt.Errorf("front matter status is missing")
	}

	frontMatter := strings.TrimRight(strings.Join(lines, "\n"), "\n")
	return "---\n" + frontMatter + "\n---\n\n" + strings.TrimLeft(body, "\n"), nil
}

func rulesetRegistrySourceDescription() string {
	return fmt.Sprintf(
		"https://github.com/%s/%s/tree/%s/%s",
		rulesetRegistryOwner,
		rulesetRegistryRepo,
		rulesetRegistryBranch,
		rulesetDirRelPath,
	)
}

func loadRulesetViewContent(ctx context.Context, projectRoot, slug string) (string, string, error) {
	localPath := rulesetPath(projectRoot, slug)
	if document.Exists(localPath) {
		content, err := os.ReadFile(localPath)
		if err != nil {
			return "", "", fmt.Errorf("failed to read %s: %w", rulesetTarget(slug), err)
		}
		parsed := parseRuleset(string(content), localPath)
		if issues := validateRulesetDocument(parsed, slug); len(issues) > 0 {
			return "", "", fmt.Errorf("local ruleset %s is invalid: %s", rulesetTarget(slug), strings.Join(issues, "; "))
		}
		return string(content), rulesetTarget(slug), nil
	}

	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	registry, err := rulesetRegistryFetcher(ctx)
	if err != nil {
		return "", "", err
	}
	for _, item := range registry {
		if item.Slug == slug {
			return item.Content, rulesetRegistryRulesetURL(slug), nil
		}
	}
	return "", "", fmt.Errorf("ruleset %q was not found locally or in the Kit registry", slug)
}

func rulesetRegistryRulesetURL(slug string) string {
	return fmt.Sprintf(
		"https://github.com/%s/%s/blob/%s/%s/%s.md",
		rulesetRegistryOwner,
		rulesetRegistryRepo,
		rulesetRegistryBranch,
		rulesetDirRelPath,
		slug,
	)
}

func ensureTrailingNewline(content string) string {
	if strings.HasSuffix(content, "\n") {
		return content
	}
	return content + "\n"
}
