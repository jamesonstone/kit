//go:build darwin || linux

package worktree

import (
	"context"
	"errors"
	"io"
	"os"
	"syscall"
	"time"
)

const terminalReadRetryInterval = 10 * time.Millisecond

type contextTerminalReader struct {
	ctx   context.Context
	input *os.File
}

func newContextTerminalReader(ctx context.Context, input *os.File) (io.Reader, func() error, error) {
	originalFlags, err := terminalFileFlags(input.Fd())
	if err != nil {
		return nil, nil, err
	}
	if err := setTerminalFileFlags(input.Fd(), originalFlags|syscall.O_NONBLOCK); err != nil {
		return nil, nil, err
	}
	restore := func() error {
		return setTerminalFileFlags(input.Fd(), originalFlags)
	}
	return &contextTerminalReader{ctx: ctx, input: input}, restore, nil
}

func (r *contextTerminalReader) Read(buffer []byte) (int, error) {
	if len(buffer) == 0 {
		return 0, nil
	}
	for {
		if err := r.ctx.Err(); err != nil {
			return 0, err
		}
		count, err := r.input.Read(buffer[:1])
		if err == nil || !errors.Is(err, syscall.EAGAIN) && !errors.Is(err, syscall.EWOULDBLOCK) {
			return count, err
		}
		timer := time.NewTimer(terminalReadRetryInterval)
		select {
		case <-r.ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return 0, r.ctx.Err()
		case <-timer.C:
		}
	}
}

func terminalFileFlags(fd uintptr) (uintptr, error) {
	flags, _, errno := syscall.Syscall(syscall.SYS_FCNTL, fd, uintptr(syscall.F_GETFL), 0)
	if errno != 0 {
		return 0, errno
	}
	return flags, nil
}

func setTerminalFileFlags(fd, flags uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_FCNTL, fd, uintptr(syscall.F_SETFL), flags)
	if errno != 0 {
		return errno
	}
	return nil
}
