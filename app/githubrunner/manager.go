package githubrunner

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ray-d-song/merc/app/dto"
	"github.com/ray-d-song/merc/app/model"
	"github.com/ray-d-song/merc/app/service"
	"go.uber.org/zap"
)

type CommandRunner interface {
	Run(ctx context.Context, dir string, name string, args ...string) ([]byte, error)
	Start(ctx context.Context, dir string, name string, args ...string) (*os.Process, error)
}

type ExecCommandRunner struct{}

func (r ExecCommandRunner) Run(ctx context.Context, dir string, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func (r ExecCommandRunner) Start(ctx context.Context, dir string, name string, args ...string) (*os.Process, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd.Process, nil
}

type Manager struct {
	DataDir          string
	Version          string
	Runner           CommandRunner
	Logger           *zap.Logger
	Client           *http.Client
	LatestReleaseURL string
}

type createPayload struct {
	RepositoryURL     string `json:"repositoryUrl"`
	RegistrationToken string `json:"registrationToken"`
	Name              string `json:"name"`
	Labels            string `json:"labels"`
	WorkDir           string `json:"workDir"`
}

func NewManager(dataDir, version string, logger *zap.Logger) *Manager {
	if version == "" {
		version = "latest"
	}
	return &Manager{
		DataDir:          dataDir,
		Version:          version,
		Runner:           ExecCommandRunner{},
		Logger:           logger.Named("githubrunner"),
		Client:           &http.Client{Timeout: 10 * time.Minute},
		LatestReleaseURL: "https://api.github.com/repos/actions/runner/releases/latest",
	}
}

func (m *Manager) ExecuteTask(ctx context.Context, task dto.RunnerTaskResp) dto.ReportRunnerTaskReq {
	report := dto.ReportRunnerTaskReq{
		Status:       model.RunnerTaskStatusSucceeded,
		RunnerStatus: model.RunnerStatusRunning,
	}

	var err error
	switch task.Type {
	case model.RunnerTaskTypeCreate, model.RunnerTaskTypeReconfigure:
		report, err = m.create(ctx, task)
	case model.RunnerTaskTypeStart:
		report, err = m.start(ctx, task.RunnerID)
	case model.RunnerTaskTypeStop:
		report, err = m.stop(ctx, task.RunnerID)
	case model.RunnerTaskTypeRemove:
		report, err = m.remove(ctx, task.RunnerID, task.PayloadJSON)
	default:
		err = fmt.Errorf("unsupported runner task type %q", task.Type)
	}

	if err != nil {
		report.Status = model.RunnerTaskStatusFailed
		report.RunnerStatus = model.RunnerStatusError
		report.LastError = service.RedactSecrets(err.Error())
	}
	return report
}

func (m *Manager) create(ctx context.Context, task dto.RunnerTaskResp) (dto.ReportRunnerTaskReq, error) {
	var payload createPayload
	if err := json.Unmarshal([]byte(task.PayloadJSON), &payload); err != nil {
		return dto.ReportRunnerTaskReq{}, fmt.Errorf("decode create payload: %w", err)
	}
	if payload.RepositoryURL == "" || payload.RegistrationToken == "" || payload.Name == "" {
		return dto.ReportRunnerTaskReq{}, fmt.Errorf("missing repository URL, token, or runner name")
	}

	installDir := m.runnerDir(task.RunnerID)
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return dto.ReportRunnerTaskReq{}, fmt.Errorf("create runner directory: %w", err)
	}
	if err := m.ensureInstalled(ctx, installDir); err != nil {
		return dto.ReportRunnerTaskReq{}, err
	}

	workDir := payload.WorkDir
	if workDir == "" {
		workDir = "_work"
	}
	args := []string{
		"--url", payload.RepositoryURL,
		"--token", payload.RegistrationToken,
		"--name", payload.Name,
		"--work", workDir,
		"--unattended",
		"--replace",
	}
	if payload.Labels != "" {
		args = append(args, "--labels", payload.Labels)
	}
	output, err := m.Runner.Run(ctx, installDir, configScript(), args...)
	safeOutput := redactValue(string(output), payload.RegistrationToken)
	if err != nil {
		return dto.ReportRunnerTaskReq{}, fmt.Errorf("configure runner: %w: %s", err, safeOutput)
	}

	report, err := m.start(ctx, task.RunnerID)
	if err != nil {
		return report, err
	}
	report.ResultJSON = fmt.Sprintf(`{"installDir":%q}`, installDir)
	report.LogSummary = safeOutput
	return report, nil
}

func (m *Manager) start(ctx context.Context, runnerID uint) (dto.ReportRunnerTaskReq, error) {
	installDir := m.runnerDir(runnerID)
	process, err := m.Runner.Start(ctx, installDir, runScript())
	if err != nil {
		return dto.ReportRunnerTaskReq{}, fmt.Errorf("start runner: %w", err)
	}
	return dto.ReportRunnerTaskReq{
		Status:       model.RunnerTaskStatusSucceeded,
		RunnerStatus: model.RunnerStatusRunning,
		ProcessID:    process.Pid,
		ResultJSON:   fmt.Sprintf(`{"pid":%d}`, process.Pid),
	}, nil
}

func (m *Manager) stop(_ context.Context, runnerID uint) (dto.ReportRunnerTaskReq, error) {
	return dto.ReportRunnerTaskReq{
		Status:       model.RunnerTaskStatusSucceeded,
		RunnerStatus: model.RunnerStatusStopped,
		ResultJSON:   fmt.Sprintf(`{"runnerId":%d}`, runnerID),
	}, nil
}

func (m *Manager) remove(ctx context.Context, runnerID uint, payloadJSON string) (dto.ReportRunnerTaskReq, error) {
	var payload createPayload
	_ = json.Unmarshal([]byte(payloadJSON), &payload)
	installDir := m.runnerDir(runnerID)
	if payload.RegistrationToken != "" && fileExists(filepath.Join(installDir, configScript())) {
		output, err := m.Runner.Run(ctx, installDir, configScript(), "remove", "--token", payload.RegistrationToken)
		if err != nil {
			return dto.ReportRunnerTaskReq{}, fmt.Errorf("remove runner config: %w: %s", err, service.RedactSecrets(string(output)))
		}
	}
	if err := os.RemoveAll(installDir); err != nil {
		return dto.ReportRunnerTaskReq{}, fmt.Errorf("remove runner directory: %w", err)
	}
	return dto.ReportRunnerTaskReq{
		Status:       model.RunnerTaskStatusSucceeded,
		RunnerStatus: model.RunnerStatusRemoved,
		ResultJSON:   fmt.Sprintf(`{"removedDir":%q}`, installDir),
	}, nil
}

func (m *Manager) ensureInstalled(ctx context.Context, installDir string) error {
	if fileExists(filepath.Join(installDir, configScript())) {
		return nil
	}
	version, err := m.resolveVersion(ctx)
	if err != nil {
		return err
	}
	archivePath := filepath.Join(installDir, runnerArchiveName(version))
	if err := m.download(ctx, runnerDownloadURL(version), archivePath); err != nil {
		return err
	}
	if strings.HasSuffix(archivePath, ".zip") {
		return unzip(archivePath, installDir)
	}
	return untargz(archivePath, installDir)
}

func (m *Manager) resolveVersion(ctx context.Context) (string, error) {
	version := strings.TrimSpace(m.Version)
	if version == "" || strings.EqualFold(version, "latest") {
		return m.latestRunnerVersion(ctx)
	}
	return strings.TrimPrefix(version, "v"), nil
}

func (m *Manager) latestRunnerVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.LatestReleaseURL, nil)
	if err != nil {
		return "", fmt.Errorf("create latest runner release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := m.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch latest runner release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch latest runner release: unexpected HTTP %d", resp.StatusCode)
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("decode latest runner release: %w", err)
	}
	version := strings.TrimPrefix(strings.TrimSpace(release.TagName), "v")
	if version == "" {
		return "", fmt.Errorf("latest runner release did not include tag_name")
	}
	return version, nil
}

func (m *Manager) download(ctx context.Context, url, target string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create runner download request: %w", err)
	}
	resp, err := m.Client.Do(req)
	if err != nil {
		return fmt.Errorf("download runner: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download runner: unexpected HTTP %d", resp.StatusCode)
	}
	file, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("create runner archive: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("write runner archive: %w", err)
	}
	return nil
}

func (m *Manager) runnerDir(runnerID uint) string {
	return filepath.Join(m.DataDir, "runners", fmt.Sprintf("%d", runnerID))
}

func runnerDownloadURL(version string) string {
	return fmt.Sprintf("https://github.com/actions/runner/releases/download/v%s/%s", version, runnerArchiveName(version))
}

func runnerArchiveName(version string) string {
	goos := runtime.GOOS
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x64"
	}
	if arch == "arm64" {
		arch = "arm64"
	}
	if goos == "darwin" {
		goos = "osx"
	}
	if goos == "windows" {
		return fmt.Sprintf("actions-runner-win-%s-%s.zip", arch, version)
	}
	return fmt.Sprintf("actions-runner-%s-%s-%s.tar.gz", goos, arch, version)
}

func configScript() string {
	if runtime.GOOS == "windows" {
		return "config.cmd"
	}
	return "./config.sh"
}

func runScript() string {
	if runtime.GOOS == "windows" {
		return "run.cmd"
	}
	return "./run.sh"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func redactValue(value, secret string) string {
	value = service.RedactSecrets(value)
	if secret == "" {
		return value
	}
	return strings.ReplaceAll(value, secret, "[redacted]")
}

func unzip(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer reader.Close()
	for _, file := range reader.File {
		target := filepath.Join(dest, file.Name)
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, file.Mode()); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := file.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			in.Close()
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		in.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func untargz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open tar.gz: %w", err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open gzip: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		target := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			if err := out.Close(); err != nil {
				return err
			}
		}
	}
}
