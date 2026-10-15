package cleanup

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

func (s *Service) recoveryStorage(ctx context.Context, repo, common string) error {
	if _, e := os.Lstat(filepath.Join(common, "info", "grafts")); !os.IsNotExist(e) {
		return keep("KEEP_LOCAL_STATE", "local graft history is unsupported")
	}
	config, err := s.git(ctx, repo, "config", "--null", "--list")
	if err != nil {
		return err
	}
	for _, entry := range strings.Split(string(config), "\x00") {
		key, value, _ := strings.Cut(entry, "\n")
		if key == "extensions.refstorage" && value != "files" {
			return keep("KEEP_LOCAL_STATE", "unsupported reference storage")
		}
	}
	return nil
}
