package cli

import (
	"path/filepath"
)

func initRefreshTargetMatches(targets map[string]bool, relativePath string) bool {
	if len(targets) == 0 {
		return true
	}
	_, ok := targets[filepath.ToSlash(relativePath)]
	return ok
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
	case instructionFileSkipped:
		s.skipped++
	}
}
