package cleanup

import (
	"context"
	"strings"
)

// checkoutFilters checks effective attributes from the incoming tree, info and global
// attributes. Merely configuring an unused driver cannot cause a checkout command.
// Unsupported Git attribute inspection fails closed.
func (s *Service) checkoutFilters(ctx context.Context, repo, head string) error {
	config, err := s.git(ctx, repo, "config", "--null", "--list")
	if err != nil {
		return err
	}
	configured := false
	for _, entry := range strings.Split(string(config), "\x00") {
		key, _, _ := strings.Cut(entry, "\n")
		if strings.HasPrefix(key, "filter.") {
			configured = true
		}
	}
	if !configured {
		return nil
	}
	paths, err := s.git(ctx, repo, "ls-tree", "-rz", "--name-only", head)
	if err != nil {
		return err
	}
	args := []string{"check-attr", "--source=" + head, "-z", "filter", "--"}
	for _, path := range strings.Split(string(paths), "\x00") {
		if path != "" {
			args = append(args, path)
		}
	}
	if len(args) == 5 {
		return nil
	}
	output, err := s.git(ctx, repo, args...)
	if err != nil {
		return err
	}
	fields := strings.Split(strings.TrimSuffix(string(output), "\x00"), "\x00")
	if len(fields)%3 != 0 {
		return keep("UNAVAILABLE", "effective filter attributes could not be established")
	}
	for i := 2; i < len(fields); i += 3 {
		if fields[i] != "unspecified" && fields[i] != "unset" {
			return keep("KEEP_LOCAL_STATE", "incoming checkout could execute a configured filter")
		}
	}
	return nil
}
