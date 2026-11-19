// Package plugin discovers opt-in instruction bundles and reconciles their
// blocks into agent entry files. It does not run plugin code, read plugin
// config, store credentials, or use the network.
package plugin

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// RootDir is the only directory Kit searches for plugins.
	RootDir = ".kit/plugins"
	// ManifestName is the manifest file inside each plugin directory.
	ManifestName = "kit-plugin.yaml"
	// SchemaVersion is the only manifest schema this Kit reads.
	SchemaVersion = 1

	maxInstructionBytes = 64 * 1024
	maxManifestBytes    = 16 * 1024

	pluginBeginPrefix = "<!-- BEGIN KIT-PLUGIN "
	pluginEndPrefix   = "<!-- END KIT-PLUGIN "
)

// Block is one enabled plugin's reconciled instruction block.
type Block struct {
	ID         string
	Body       string
	ConfigPath string
}

var pluginIDPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)

// ValidateID reports whether id can be a plugin directory name and manifest id.
func ValidateID(id string) error {
	if !pluginIDPattern.MatchString(id) {
		return fmt.Errorf("plugin id %q must match %s", id, pluginIDPattern.String())
	}
	return nil
}

func pluginBeginMarker(id string) string {
	return pluginBeginPrefix + id + " -->"
}

func pluginEndMarker(id string) string {
	return pluginEndPrefix + id + " -->"
}

func pluginConfigComment(configPath string) string {
	return "<!-- Plugin config: " + configPath + " -->"
}

func pluginIDFromBegin(line string) (string, bool) {
	if !strings.HasPrefix(line, pluginBeginPrefix) || !strings.HasSuffix(line, " -->") {
		return "", false
	}
	id := strings.TrimSuffix(strings.TrimPrefix(line, pluginBeginPrefix), " -->")
	if id == "" || strings.ContainsAny(id, " \t") {
		return "", false
	}
	return id, true
}
