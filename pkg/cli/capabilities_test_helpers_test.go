package cli

import (
	"bytes"
)

func executeCapabilitiesCommand(args ...string) (string, error) {
	cmd := newCapabilitiesCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}
