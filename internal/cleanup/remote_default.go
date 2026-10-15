package cleanup

import (
	"context"
	"strings"
)

func (s *Service) remoteDefault(ctx context.Context, repository, remote string) (string, string, error) {
	out, err := s.git(ctx, repository, "ls-remote", "--symref", remote, "HEAD")
	if err != nil {
		return "", "", err
	}
	var ref, head string
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[0] == "ref:" && fields[2] == "HEAD" {
			ref = fields[1]
		}
		if len(fields) == 2 && fields[1] == "HEAD" {
			head = fields[0]
		}
	}
	if !strings.HasPrefix(ref, "refs/heads/") || requireCanonicalOID(head) != nil {
		return "", "", keep("UNAVAILABLE", "live remote default branch is unavailable")
	}
	return ref, head, nil
}
