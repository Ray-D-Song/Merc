package githubrunner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ray-d-song/merc/app/dto"
	"github.com/ray-d-song/merc/app/model"
	"go.uber.org/zap"
)

type fakeCommandRunner struct {
	runName  string
	runArgs  []string
	startDir string
}

func (r *fakeCommandRunner) Run(_ context.Context, _ string, name string, args ...string) ([]byte, error) {
	r.runName = name
	r.runArgs = args
	return []byte("configured token secret-token"), nil
}

func (r *fakeCommandRunner) Start(_ context.Context, dir string, _ string, _ ...string) (*os.Process, error) {
	r.startDir = dir
	return os.FindProcess(os.Getpid())
}

func TestManagerExecuteCreateTaskConfiguresAndStartsRunner(t *testing.T) {
	tmp := t.TempDir()
	installDir := filepath.Join(tmp, "runners", "7")
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installDir, strings.TrimPrefix(configScript(), "./")), []byte(""), 0o755); err != nil {
		t.Fatal(err)
	}

	fake := &fakeCommandRunner{}
	manager := NewManager(tmp, "test", zap.NewNop())
	manager.Runner = fake

	report := manager.ExecuteTask(context.Background(), dto.RunnerTaskResp{
		ID:          3,
		RunnerID:    7,
		Type:        model.RunnerTaskTypeCreate,
		PayloadJSON: `{"repositoryUrl":"https://github.com/openai/codex","registrationToken":"secret-token","name":"runner-1","labels":"self-hosted","workDir":"_work"}`,
	})

	if report.Status != model.RunnerTaskStatusSucceeded {
		t.Fatalf("status = %s, want %s: %s", report.Status, model.RunnerTaskStatusSucceeded, report.LastError)
	}
	if report.RunnerStatus != model.RunnerStatusRunning {
		t.Fatalf("runner status = %s, want %s", report.RunnerStatus, model.RunnerStatusRunning)
	}
	if fake.runName != configScript() {
		t.Fatalf("config command = %s, want %s", fake.runName, configScript())
	}
	if !containsArg(fake.runArgs, "--url") || !containsArg(fake.runArgs, "https://github.com/openai/codex") {
		t.Fatalf("config args missing repository URL: %#v", fake.runArgs)
	}
	if strings.Contains(report.LogSummary, "secret-token") {
		t.Fatalf("log summary leaked token: %s", report.LogSummary)
	}
	if fake.startDir != installDir {
		t.Fatalf("start dir = %s, want %s", fake.startDir, installDir)
	}
}

func containsArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}
