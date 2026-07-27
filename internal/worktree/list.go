package worktree

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type listOptions struct {
	sortBy       string
	rootPosition string
	reverse      bool
	plain        bool
}

type listSelectorFunc func(context.Context, []worktreeEntry) (worktreeEntry, bool, error)

func parseListOptions(args []string) (listOptions, error) {
	options := listOptions{sortBy: "updated", rootPosition: "top"}
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--reverse":
			options.reverse = true
		case args[i] == "--plain":
			options.plain = true
		case args[i] == "--sort":
			if i+1 >= len(args) {
				return listOptions{}, errors.New("usage: git wt list [--sort <updated|state|head|path>] [--reverse] [--plain]")
			}
			i++
			options.sortBy = strings.ToLower(args[i])
		case strings.HasPrefix(args[i], "--sort="):
			options.sortBy = strings.ToLower(strings.TrimPrefix(args[i], "--sort="))
		case args[i] == "--root-position":
			if i+1 >= len(args) {
				return listOptions{}, errors.New("usage: git wt list [--sort <updated|state|head|path>] [--root-position <top|bottom>] [--reverse] [--plain]")
			}
			i++
			options.rootPosition = strings.ToLower(args[i])
		case strings.HasPrefix(args[i], "--root-position="):
			options.rootPosition = strings.ToLower(
				strings.TrimPrefix(args[i], "--root-position="),
			)
		default:
			return listOptions{}, fmt.Errorf("unknown list flag %q", args[i])
		}
	}
	if options.sortBy != "updated" && options.sortBy != "state" && options.sortBy != "head" && options.sortBy != "path" {
		return listOptions{}, fmt.Errorf("unsupported list sort %q (want updated, state, head, or path)", options.sortBy)
	}
	if options.rootPosition != "top" && options.rootPosition != "bottom" {
		return listOptions{}, fmt.Errorf(
			"unsupported root position %q (want top or bottom)",
			options.rootPosition,
		)
	}
	return options, nil
}

func (a *App) list(ctx context.Context, cwd string, args []string) error {
	options, err := parseListOptions(args)
	if err != nil {
		return err
	}
	repo, err := a.repository(ctx, cwd)
	if err != nil {
		return err
	}
	entries, err := a.worktrees(ctx, repo.top)
	if err != nil {
		return err
	}
	a.populateListMetadata(ctx, entries)
	a.populateListPullRequests(ctx, repo.top, entries)
	applyListUpdatedDisplay(entries, time.Local)
	sortListEntries(entries, options)

	if !options.plain && a.isTerminal() {
		selected, ok, selectErr := a.selectList(ctx, entries)
		if selectErr != nil {
			return selectErr
		}
		if !ok {
			return nil
		}
		return a.runShell(ctx, selected.path)
	}
	return a.writeList(entries)
}

func (a *App) populateListMetadata(ctx context.Context, entries []worktreeEntry) {
	for i := range entries {
		entries[i].updatedText = "unknown"
		updated, updateErr := a.gitText(ctx, entries[i].path, "log", "-1", "--format=%cI", "HEAD")
		if updateErr == nil {
			if parsed, parseErr := time.Parse(time.RFC3339, updated); parseErr == nil {
				entries[i].lastUpdated = parsed
				entries[i].updatedText = parsed.Format("Jan 02, 2006")
			}
		}

		entries[i].state = "clean"
		dirty, statusErr := a.status(ctx, entries[i].path, false)
		if statusErr != nil {
			entries[i].state = "unknown"
		} else if dirty != "" {
			entries[i].state = "dirty"
		}
	}
}

func applyListUpdatedDisplay(entries []worktreeEntry, location *time.Location) {
	for i := range entries {
		if entries[i].lastUpdated.IsZero() {
			continue
		}
		entries[i].updatedText = entries[i].lastUpdated.
			In(location).
			Format("Jan 02, 2006 15:04")
	}
}

func sortListEntries(entries []worktreeEntry, options listOptions) {
	sort.SliceStable(entries, func(i, j int) bool {
		comparison := compareListEntries(entries[i], entries[j], options.sortBy)
		if options.reverse {
			return comparison > 0
		}
		return comparison < 0
	})
	pinPrimaryListEntry(entries, options.rootPosition)
}

func pinPrimaryListEntry(entries []worktreeEntry, position string) {
	primary := -1
	for i := range entries {
		if entries[i].primary {
			primary = i
			break
		}
	}
	if primary < 0 {
		return
	}
	entry := entries[primary]
	switch position {
	case "top":
		copy(entries[1:primary+1], entries[:primary])
		entries[0] = entry
	case "bottom":
		copy(entries[primary:], entries[primary+1:])
		entries[len(entries)-1] = entry
	}
}

func compareListEntries(left, right worktreeEntry, sortBy string) int {
	var comparison int
	switch sortBy {
	case "updated":
		switch {
		case left.lastUpdated.After(right.lastUpdated):
			comparison = -1
		case left.lastUpdated.Before(right.lastUpdated):
			comparison = 1
		}
	case "state":
		comparison = strings.Compare(left.state, right.state)
	case "head":
		comparison = strings.Compare(displayHead(left), displayHead(right))
	case "path":
		comparison = strings.Compare(left.path, right.path)
	}
	if comparison == 0 {
		return strings.Compare(left.path, right.path)
	}
	return comparison
}

func (a *App) writeList(entries []worktreeEntry) error {
	if err := a.writef("STATE\tHEAD\tPR#\tLAST UPDATED\tPATH\n"); err != nil {
		return err
	}
	for _, entry := range entries {
		if err := a.writef("%s\t%s\t%s\t%s\t%s\n", entry.state, displayHead(entry), entry.prText, entry.updatedText, entry.path); err != nil {
			return err
		}
	}
	return nil
}

func displayHead(entry worktreeEntry) string {
	if entry.branch != "" {
		return entry.branch
	}
	return "detached@" + shortOID(entry.head)
}
