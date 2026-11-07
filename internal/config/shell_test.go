package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestShellConfigurationDefaultsAndOverrides(t *testing.T) {
	t.Setenv("SHELL", "/personal/shell-not-to-save")
	for _, tc := range []struct {
		name, body string
		command    []string
	}{
		{name: "omitted"},
		{name: "null", body: "shell: null\n"},
		{name: "empty mapping", body: "shell: {}\n"},
		{name: "empty command", body: "shell:\n  command: []\n"},
		{name: "null command", body: "shell:\n  command: null\n"},
		{name: "zsh", body: "shell:\n  command: [zsh, -lic]\n", command: []string{"zsh", "-lic"}},
		{name: "alias", body: "invocation: &shell\n  command: [zsh, -lic]\nshell: *shell\n", command: []string{"zsh", "-lic"}},
		{name: "powershell", body: "shell:\n  command: [pwsh, -Login, -Command]\n", command: []string{"pwsh", "-Login", "-Command"}},
		{name: "literal arguments", body: "shell:\n  command: ['$SHELL', '~/profile', 'path with spaces']\n", command: []string{"$SHELL", "~/profile", "path with spaces"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, ConfigFileName)
			if err := os.WriteFile(path, []byte("schema_version: 2\n"+tc.body), 0644); err != nil {
				t.Fatal(err)
			}
			cfg, inspection, err := LoadWithInspection(root)
			if err != nil || inspection.HasErrors() {
				t.Fatalf("load: %v, findings: %+v", err, inspection.Findings)
			}
			wantSource := "user"
			if len(tc.command) != 0 {
				wantSource = "command"
				if !reflect.DeepEqual(cfg.Shell.Command, tc.command) {
					t.Fatalf("command = %#v, want %#v", cfg.Shell.Command, tc.command)
				}
			}
			if cfg.Shell.Source() != wantSource {
				t.Fatalf("source = %s, want %s", cfg.Shell.Source(), wantSource)
			}
			if err := Save(root, cfg); err != nil {
				t.Fatal(err)
			}
			loaded, err := Load(root)
			if err != nil || loaded.Shell.Source() != wantSource {
				t.Fatalf("round trip: %v, config: %+v", err, loaded)
			}
			if len(tc.command) != 0 && !reflect.DeepEqual(loaded.Shell.Command, tc.command) {
				t.Fatal("round trip changed the exact command prefix")
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(data), "/personal/shell-not-to-save") {
				t.Fatal("configuration captured the current user's environment")
			}
		})
	}
	if Default().Shell.Source() != "user" || len(Default().Shell.Command) != 0 {
		t.Fatal("new projects must resolve the current user's environment")
	}
}

func TestShellConfigurationRejectsInvalidArguments(t *testing.T) {
	for _, command := range [][]string{{""}, {"  "}, {"zsh", ""}, {"zsh", "line\nbreak"}, {"zsh", "nul\x00byte"}} {
		root := t.TempDir()
		cfg := Default()
		cfg.Shell.Command = command
		if err := Save(root, cfg); err != nil {
			t.Fatal(err)
		}
		_, inspection, err := LoadWithInspection(root)
		if err != nil || !inspection.HasErrors() {
			t.Fatalf("invalid argument accepted: err=%v findings=%+v", err, inspection.Findings)
		}
		found := false
		for _, finding := range inspection.Findings {
			if finding.Field == "shell.command" && !finding.Repairable {
				found = true
			}
		}
		if !found {
			t.Fatal("missing nonrepairable shell.command finding")
		}
		if _, err := Load(root); err == nil {
			t.Fatal("normal load must not silently accept an invalid override")
		}
	}
}

func TestShellConfigurationRejectsMisspelledFieldsAndWrongShapes(t *testing.T) {
	for _, body := range []string{
		"shell: zsh\n", "shell: [zsh, -lic]\n",
		"shell:\n  command: [123]\n", "shell:\n  command: [zsh, true]\n",
		"shell:\n  command: [null]\n", "shell:\n  command: [null, zsh, -lic]\n",
		"shell: &cycle\n  command: [*cycle]\n",
		"shell:\n  command: zsh -lic\n", "shell:\n  commands: [zsh, -lic]\n",
		"shell:\n  command: [zsh]\n  command: [bash]\n",
	} {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, ConfigFileName), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(root); err == nil {
			t.Fatalf("invalid shell configuration accepted: %s", body)
		}
	}
}
