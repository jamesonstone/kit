package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestMintControllerSkipsAutomaticPublication(t *testing.T) {
	root := mintFixture(t)
	changes, notes, err := planInitMintIntegration(root, true, mintPipelineSelection{})
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].relativePath != ".github/workflows/control.yaml" || len(notes) != 0 {
		t.Fatalf("changes=%+v notes=%v", changes, notes)
	}
	if !strings.Contains(changes[0].after, "# Scaffolded by Kit with Mint v0.9.0") || !strings.Contains(changes[0].after, "jamesonstone/mint@"+mintPipelineActionRef) {
		t.Fatal("controller version/pin differs from pipeline")
	}
	if strings.Contains(changes[0].after, "command: release-publish") {
		t.Fatal("second publisher in environment controller")
	}
}

func TestMintControllerRejectsAllWorkflowSelections(t *testing.T) {
	for _, selection := range []mintPipelineSelection{
		{workflow: ".github/workflows/deploy.yml"}, {job: "deploy"}, {beforeStep: "build"},
	} {
		root := t.TempDir()
		if _, _, err := planInitMintIntegration(root, true, selection); err == nil {
			t.Fatalf("controller accepted selection %+v", selection)
		}
		if _, err := buildInitRefreshPlan(context.Background(), root, initRefreshOptions{mint: true, mintSelection: selection}); err == nil {
			t.Fatalf("refresh accepted selection %+v", selection)
		}
		if config.Exists(root) {
			t.Fatal("conflict partially initialized config")
		}
	}
	if err := validateMintControllerOptions(true, true, mintPipelineSelection{}); err == nil {
		t.Fatal("controller accepted --mint-workflows")
	}
}

func TestMintIncompleteNotesRemainInRawPromptAndOutputOnly(t *testing.T) {
	notes := []string{"ordinary diagnostic", "INCOMPLETE: select the existing deployment job"}
	prompt := appendMintIncompletePrompt("raw initialization prompt", notes)
	if !strings.HasPrefix(prompt, "raw initialization prompt") || !strings.Contains(prompt, notes[1]) || strings.Contains(prompt, notes[0]) {
		t.Fatal(prompt)
	}
	output := captureStdout(t, func() { printInitRefreshNotes(notes, initRefreshOptions{outputOnly: true}) })
	if !strings.Contains(output, notes[1]) || strings.Contains(output, notes[0]) {
		t.Fatal(output)
	}
}

func TestMintEverySelectionFieldActivatesRefreshPlanning(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	if err := config.Save(root, defaultInitConfig()); err != nil {
		t.Fatal(err)
	}
	for _, selection := range []mintPipelineSelection{{workflow: ".github/workflows/deploy.yml"}, {job: "deploy"}, {beforeStep: "publish"}} {
		_, err := buildInitRefreshPlan(context.Background(), root, initRefreshOptions{mintSelection: selection})
		if err == nil || !strings.Contains(err.Error(), "incomplete Mint selection") {
			t.Fatalf("selection %+v ignored: %v", selection, err)
		}
	}
}

func TestInitMintIncompleteSelectionIsVisibleWithoutExtraPublisher(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	writeFile(t, filepath.Join(root, ".github/workflows/deploy.yaml"), "name: Deploy\non: push\npermissions:\n  contents: read\njobs:\n  deploy:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo existing\n")
	withInitFlags(t, func() {
		initOutputOnly = true
		output := captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatal(err)
			}
		})
		if !strings.Contains(output, "INCOMPLETE") || !strings.Contains(output, "Incomplete Mint integration") {
			t.Fatal("raw prompt hid incomplete adoption")
		}
	})
	if _, err := os.Stat(filepath.Join(root, ".github/workflows/mint-release.yaml")); !os.IsNotExist(err) {
		t.Fatal("existing deployment got competing publisher")
	}
}

func TestMintRefreshOutputOnlyStillReportsIncomplete(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	if err := config.Save(root, defaultInitConfig()); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, ".github/workflows/deploy.yaml"), "name: Deploy\non: push\npermissions: {}\njobs: {}\n")
	output := captureStdout(t, func() {
		if _, err := runInitRefreshWithSnapshot(root, initRefreshOptions{mintWorkflows: true, outputOnly: true}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(output, "INCOMPLETE") {
		t.Fatal("output-only refresh hid incomplete adoption")
	}
	output = captureStdout(t, func() {
		if err := printCurrentMintIncompleteNotes(root, initRefreshOptions{mintWorkflows: true}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(output, "INCOMPLETE") {
		t.Fatal("already-current reconcile hid incomplete adoption")
	}
}
