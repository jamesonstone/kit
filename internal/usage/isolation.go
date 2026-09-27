package usage

import (
	"os"
	"strings"
	"testing"
)

// DisableEnv turns off usage recording for a process when set to a non-empty
// value other than "0" or "false".
const DisableEnv = "KIT_USAGE_DISABLED"

// recordingInTests lets this package's own tests exercise the store.
var recordingInTests = false

// suppressed keeps tests, development builds, and explicit opt-outs from
// writing to the user's real usage log.
func suppressed(version string) bool {
	if value := strings.TrimSpace(strings.ToLower(os.Getenv(DisableEnv))); value != "" && value != "0" && value != "false" {
		return true
	}
	if testing.Testing() && !recordingInTests {
		return true
	}
	version = strings.TrimSpace(version)
	return version == "" || version == "dev"
}
