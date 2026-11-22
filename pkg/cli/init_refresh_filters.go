package cli

import (
	"path/filepath"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func initRefreshTargetMatches(targets map[string]bool, relativePath string) bool {
	if len(targets) == 0 {
		return true
	}
	_, ok := targets[filepath.ToSlash(relativePath)]
	return ok
}

func (opts initRefreshOptions) resetsProjectConfig(targets map[string]bool) bool {
	return opts.force && targets[config.ConfigFileName]
}

func (s *initRefreshStats) recordFileChange(change initRefreshFileChange) {
	s.recordResult(change.result)
}

func (s *initRefreshStats) recordResult(result instructionFileWriteResult) {
	switch result {
	case instructionFileCreated:
		s.created++
	case instructionFileUpdated:
		s.updated++
	case instructionFileMerged:
		s.merged++
	case instructionFileRemoved:
		s.removed++
	case instructionFileSkipped:
		s.skipped++
	}
}
