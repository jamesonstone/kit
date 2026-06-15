package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeCICommandRunner struct {
	outputs map[string]string
	errors  map[string]error
	calls   []string
}

func (f *fakeCICommandRunner) Output(dir string, name string, args ...string) ([]byte, error) {
	output, err := f.OutputAllowError(dir, name, args...)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (f *fakeCICommandRunner) OutputAllowError(dir string, name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	f.calls = append(f.calls, key)
	output, ok := f.outputs[key]
	if !ok {
		return nil, fmt.Errorf("unexpected command: %s", key)
	}
	return []byte(output), f.errors[key]
}

func TestRunCIDiagnosesDefaultBranchFailureAndCachesGitHubConfig(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	restoreCWD := chdirForTest(t, projectRoot)
	defer restoreCWD()

	fake := newFakeCIRunner()
	fake.outputs["gh run list --branch main --status completed --limit 30 --json "+ciRunListJSONFields+" --repo jamesonstone/kit"] = `[
  {"databaseId":101,"conclusion":"failure","status":"completed","workflowName":"Tests","displayTitle":"go test","headBranch":"main","headSha":"abc","url":"https://github.com/jamesonstone/kit/actions/runs/101"}
]`
	fake.outputs["gh run view 101 --json attempt,conclusion,createdAt,databaseId,displayTitle,event,headBranch,headSha,jobs,name,number,startedAt,status,updatedAt,url,workflowDatabaseId,workflowName --repo jamesonstone/kit"] = `{
  "databaseId":101,
  "conclusion":"failure",
  "status":"completed",
  "workflowName":"Tests",
  "displayTitle":"go test",
  "headBranch":"main",
  "headSha":"abc",
  "url":"https://github.com/jamesonstone/kit/actions/runs/101",
  "jobs":[{"databaseId":201,"name":"test","conclusion":"failure","status":"completed","steps":[{"name":"go test ./...","conclusion":"failure"}]}]
}`
	fake.outputs["gh run view 101 --log-failed --job 201 --repo jamesonstone/kit"] = "test\tgo test ./...\tError: expected nil\n"
	restoreRunner := stubCIRunner(fake)
	defer restoreRunner()

	var out bytes.Buffer
	exitCode, err := runCIWithOptions(ciOptions{LogLines: 200}, &out)
	if err != nil {
		t.Fatalf("runCIWithOptions() error = %v", err)
	}
	if exitCode != 1 {
		t.Fatalf("exitCode = %d, want 1", exitCode)
	}
	content := out.String()
	for _, check := range []string{"CI Diagnosis", "Root Cause", "Error: expected nil", "Agent Prompt"} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected output to contain %q, got:\n%s", check, content)
		}
	}
	configData, err := os.ReadFile(filepath.Join(projectRoot, ".kit.yaml"))
	if err != nil {
		t.Fatalf("ReadFile(.kit.yaml) error = %v", err)
	}
	for _, check := range []string{"github:", "repository: jamesonstone/kit", "default_branch: main"} {
		if !strings.Contains(string(configData), check) {
			t.Fatalf("expected cached config to contain %q, got:\n%s", check, configData)
		}
	}
}

func TestRunCIJSONIncludesAgentPrompt(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	restoreCWD := chdirForTest(t, projectRoot)
	defer restoreCWD()

	fake := newFakeCIRunner()
	fake.outputs["gh run list --branch main --status completed --limit 30 --json "+ciRunListJSONFields+" --repo jamesonstone/kit"] = `[]`
	restoreRunner := stubCIRunner(fake)
	defer restoreRunner()

	var out bytes.Buffer
	exitCode, err := runCIWithOptions(ciOptions{JSON: true, LogLines: 200}, &out)
	if err != nil {
		t.Fatalf("runCIWithOptions() error = %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0", exitCode)
	}
	var diagnosis ciDiagnosis
	if err := json.Unmarshal(out.Bytes(), &diagnosis); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; output:\n%s", err, out.String())
	}
	if diagnosis.FailureFound {
		t.Fatal("expected no failure")
	}
	if !strings.Contains(diagnosis.AgentPrompt, "No CI fix is currently required") {
		t.Fatalf("expected agent prompt in JSON, got %#v", diagnosis.AgentPrompt)
	}
}

func TestRunCIPRURLUsesActionCheckRun(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	restoreCWD := chdirForTest(t, projectRoot)
	defer restoreCWD()

	fake := newFakeCIRunner()
	fake.outputs["gh pr view 67 --json number,headRefName,headRefOid,title,url --repo Patient-Driven-Care/cortex"] = `{"number":67,"headRefName":"GH-66","headRefOid":"def","title":"prod integration","url":"https://github.com/Patient-Driven-Care/cortex/pull/67"}`
	fake.outputs["gh pr checks 67 --json bucket,completedAt,description,event,link,name,startedAt,state,workflow --repo Patient-Driven-Care/cortex"] = `[
  {"bucket":"fail","name":"test","state":"failure","workflow":"Tests","link":"https://github.com/Patient-Driven-Care/cortex/actions/runs/777/jobs/888"},
  {"bucket":"fail","name":"Codecov","state":"failure","link":"https://codecov.example/check"}
]`
	fake.outputs["gh run view 777 --json attempt,conclusion,createdAt,databaseId,displayTitle,event,headBranch,headSha,jobs,name,number,startedAt,status,updatedAt,url,workflowDatabaseId,workflowName --repo Patient-Driven-Care/cortex"] = `{
  "databaseId":777,
  "conclusion":"failure",
  "status":"completed",
  "workflowName":"Tests",
  "displayTitle":"go test",
  "headBranch":"GH-66",
  "headSha":"def",
  "url":"https://github.com/Patient-Driven-Care/cortex/actions/runs/777",
  "jobs":[{"databaseId":888,"name":"test","conclusion":"failure","status":"completed"}]
}`
	fake.outputs["gh run view 777 --log-failed --job 888 --repo Patient-Driven-Care/cortex"] = "test\tRun tests\tfatal: missing config\n"
	restoreRunner := stubCIRunner(fake)
	defer restoreRunner()

	var out bytes.Buffer
	exitCode, err := runCIWithOptions(ciOptions{
		PRRef:    "https://github.com/Patient-Driven-Care/cortex/pull/67",
		LogLines: 200,
	}, &out)
	if err != nil {
		t.Fatalf("runCIWithOptions() error = %v", err)
	}
	if exitCode != 1 {
		t.Fatalf("exitCode = %d, want 1", exitCode)
	}
	content := out.String()
	for _, check := range []string{"Patient-Driven-Care/cortex", "fatal: missing config", "external: Codecov"} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected output to contain %q, got:\n%s", check, content)
		}
	}
}

func TestRunCIParsesPRChecksWhenGHReturnsFailureExit(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	restoreCWD := chdirForTest(t, projectRoot)
	defer restoreCWD()

	fake := newFakeCIRunner()
	checksKey := "gh pr checks 67 --json bucket,completedAt,description,event,link,name,startedAt,state,workflow --repo jamesonstone/kit"
	fake.outputs["gh pr view 67 --json number,headRefName,headRefOid,title,url --repo jamesonstone/kit"] = `{"number":67,"headRefName":"GH-67","headRefOid":"def","title":"ci","url":"https://github.com/jamesonstone/kit/pull/67"}`
	fake.outputs[checksKey] = `[{"bucket":"fail","name":"test","state":"failure","workflow":"Tests","link":"https://github.com/jamesonstone/kit/actions/runs/777"}]`
	fake.errors[checksKey] = fmt.Errorf("exit status 1")
	fake.outputs["gh run view 777 --json attempt,conclusion,createdAt,databaseId,displayTitle,event,headBranch,headSha,jobs,name,number,startedAt,status,updatedAt,url,workflowDatabaseId,workflowName --repo jamesonstone/kit"] = `{
  "databaseId":777,
  "conclusion":"failure",
  "status":"completed",
  "workflowName":"Tests",
  "displayTitle":"go test",
  "jobs":[{"databaseId":888,"name":"test","conclusion":"failure","status":"completed"}]
}`
	fake.outputs["gh run view 777 --log-failed --job 888 --repo jamesonstone/kit"] = "test\tRun tests\tError: failed\n"
	restoreRunner := stubCIRunner(fake)
	defer restoreRunner()

	var out bytes.Buffer
	exitCode, err := runCIWithOptions(ciOptions{PRRef: "67", LogLines: 200}, &out)
	if err != nil {
		t.Fatalf("runCIWithOptions() error = %v", err)
	}
	if exitCode != 1 {
		t.Fatalf("exitCode = %d, want 1", exitCode)
	}
	if !strings.Contains(out.String(), "Error: failed") {
		t.Fatalf("expected failure output from parsed PR checks, got:\n%s", out.String())
	}
}

func TestRunIDTakesPriorityOverWorkflowFlag(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	restoreCWD := chdirForTest(t, projectRoot)
	defer restoreCWD()

	fake := newFakeCIRunner()
	fake.outputs["gh run view 101 --json attempt,conclusion,createdAt,databaseId,displayTitle,event,headBranch,headSha,jobs,name,number,startedAt,status,updatedAt,url,workflowDatabaseId,workflowName --repo jamesonstone/kit"] = `{
  "databaseId":101,
  "conclusion":"failure",
  "status":"completed",
  "workflowName":"Tests",
  "displayTitle":"go test",
  "url":"https://github.com/jamesonstone/kit/actions/runs/101",
  "jobs":[{"databaseId":201,"name":"test","conclusion":"failure","status":"completed"}]
}`
	fake.outputs["gh run view 101 --log-failed --job 201 --repo jamesonstone/kit"] = "test\tRun tests\tError: failed\n"
	restoreRunner := stubCIRunner(fake)
	defer restoreRunner()

	var out bytes.Buffer
	exitCode, err := runCIWithOptions(ciOptions{
		RunID:       "101",
		WorkflowRef: "does-not-exist",
		LogLines:    200,
	}, &out)
	if err != nil {
		t.Fatalf("runCIWithOptions() error = %v", err)
	}
	if exitCode != 1 {
		t.Fatalf("exitCode = %d, want 1", exitCode)
	}
	for _, call := range fake.calls {
		if strings.Contains(call, "workflow list") {
			t.Fatalf("did not expect workflow resolution when --run is set; calls=%v", fake.calls)
		}
	}
}

func TestMatchCIWorkflowExactThenUniqueSubstring(t *testing.T) {
	workflows := []ciWorkflow{
		{Name: "Tests", Path: ".github/workflows/test.yml"},
		{Name: "Lint", Path: ".github/workflows/lint.yml"},
	}
	got, err := matchCIWorkflow(".github/workflows/test.yml", workflows)
	if err != nil {
		t.Fatalf("matchCIWorkflow(path) error = %v", err)
	}
	if got.Name != "Tests" {
		t.Fatalf("path match = %#v, want Tests", got)
	}
	got, err = matchCIWorkflow("lin", workflows)
	if err != nil {
		t.Fatalf("matchCIWorkflow(substring) error = %v", err)
	}
	if got.Name != "Lint" {
		t.Fatalf("substring match = %#v, want Lint", got)
	}
	if _, err := matchCIWorkflow("t", workflows); err == nil {
		t.Fatal("expected ambiguous workflow substring to fail")
	}
}

func newFakeCIRunner() *fakeCICommandRunner {
	return &fakeCICommandRunner{
		outputs: map[string]string{
			"git remote get-url origin": "git@github.com:jamesonstone/kit.git\n",
			"gh auth status":            "github.com\n",
			"gh repo view --json nameWithOwner,defaultBranchRef --repo jamesonstone/kit": `{"nameWithOwner":"jamesonstone/kit","defaultBranchRef":{"name":"main"}}`,
		},
		errors: map[string]error{},
	}
}

func stubCIRunner(fake *fakeCICommandRunner) func() {
	previous := ciRunner
	previousDispatchResolver := dispatchCurrentRepoResolver
	ciRunner = fake
	dispatchCurrentRepoResolver = func() (string, string, error) {
		return "jamesonstone", "kit", nil
	}
	return func() {
		ciRunner = previous
		dispatchCurrentRepoResolver = previousDispatchResolver
	}
}
