//go:build windows

package worktree

import (
	"context"
	"fmt"
	"io"
	"os"
	"syscall"
)

const terminalReadWaitMilliseconds = 10

type contextTerminalReader struct {
	ctx   context.Context
	input *os.File
}

func newContextTerminalReader(ctx context.Context, input *os.File) (io.Reader, func() error, error) {
	return &contextTerminalReader{ctx: ctx, input: input}, func() error { return nil }, nil
}

func (r *contextTerminalReader) Read(buffer []byte) (int, error) {
	if len(buffer) == 0 {
		return 0, nil
	}
	for {
		if err := r.ctx.Err(); err != nil {
			return 0, err
		}
		event, err := syscall.WaitForSingleObject(syscall.Handle(r.input.Fd()), terminalReadWaitMilliseconds)
		if err != nil {
			return 0, fmt.Errorf("wait for terminal input: %w", err)
		}
		switch event {
		case syscall.WAIT_OBJECT_0:
			return r.input.Read(buffer[:1])
		case syscall.WAIT_TIMEOUT:
			continue
		default:
			return 0, fmt.Errorf("wait for terminal input returned event %d", event)
		}
	}
}
