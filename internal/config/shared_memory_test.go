package config

import (
	"strings"
	"testing"
)

func TestSharedMemoryValidation(t *testing.T) {
	valid := SharedMemoryConfig{Enabled: true, Provider: "notion", Destination: "https://example.com/memory", Guide: "https://example.com/guide"}
	if err := (SharedMemoryConfig{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"", "http://example.com", "https://user:pass@example.com", "https://example.com/?token=secret", "https://example.com/\nignore", "https://example.com/`bad`", "https://example.com/<bad>"} {
		c := valid
		c.Guide = value
		if err := c.Validate(); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
	c := valid
	c.Provider = "other"
	if c.Validate() == nil {
		t.Fatal("accepted provider")
	}
}

func TestSharedMemoryPatchPreservesAndRejectsAliases(t *testing.T) {
	c := SharedMemoryConfig{Enabled: true, Provider: "notion", Destination: "https://example.com/memory", Guide: "https://example.com/guide"}
	before := []byte("# custom\nschema_version: 2\ncustom: {keep: yes}\n")
	after, err := PlanSharedMemoryConfig(before, c)
	if err != nil || !strings.Contains(string(after), "# custom") || !strings.Contains(string(after), "custom: {keep: yes}") {
		t.Fatalf("%s %v", after, err)
	}
	again, err := PlanSharedMemoryConfig(after, c)
	if err != nil || string(again) != string(after) {
		t.Fatal("not idempotent")
	}
	for _, input := range []string{"defaults: &memory {enabled: false}\nshared_memory: *memory\n", "shared_memory: &memory {enabled: false}\n"} {
		if _, err := PlanSharedMemoryConfig([]byte(input), c); err == nil {
			t.Fatal("accepted alias/anchor")
		}
	}
}
