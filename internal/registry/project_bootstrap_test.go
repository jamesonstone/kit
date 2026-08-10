package registry

import (
	"strings"
	"testing"
)

func TestProjectConfigCarriesCanonicalBootstrapSettings(t *testing.T) {
	config := NewProjectConfig(SourceConfig{Repo: "owner/catalog"})
	if config.Bootstrap.ScaffoldVersion != 1 || config.Bootstrap.Workflow != "repository-bootstrap" {
		t.Fatalf("bootstrap config = %#v", config.Bootstrap)
	}
	content, err := MarshalProject(config)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"bootstrap:", "scaffold_version: 1", "workflow: repository-bootstrap"} {
		if !strings.Contains(string(content), expected) {
			t.Errorf("project config is missing %q\n%s", expected, content)
		}
	}
}
