package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/ray-d-song/merc/app/dto"
	"github.com/ray-d-song/merc/app/githubrunner"
	"github.com/ray-d-song/merc/app/infra/config"
	"github.com/ray-d-song/merc/app/model"
	"go.uber.org/zap"
)

type Runtime struct {
	cfg     *config.AppConfig
	logger  *zap.Logger
	client  *http.Client
	manager *githubrunner.Manager
	nodeKey string
}

func NewRuntime(cfg *config.AppConfig, logger *zap.Logger) *Runtime {
	dataDir := cfg.Agent.DataDir
	if dataDir == "" {
		dataDir = "./merc-agent"
	}
	nodeKey := cfg.Node.NodeKey
	if nodeKey == "" {
		nodeKey = defaultNodeKey()
	}
	return &Runtime{
		cfg:     cfg,
		logger:  logger.Named("agent"),
		client:  &http.Client{Timeout: 30 * time.Second},
		manager: githubrunner.NewManager(dataDir, cfg.Runner.DefaultVersion, logger),
		nodeKey: nodeKey,
	}
}

func (r *Runtime) Run(ctx context.Context) error {
	if r.cfg.Agent.ServerURL == "" {
		return fmt.Errorf("agent.server_url is required")
	}
	if r.cfg.Agent.Token == "" {
		return fmt.Errorf("agent.token is required")
	}
	if err := os.MkdirAll(r.cfg.Agent.DataDir, 0o755); err != nil {
		return fmt.Errorf("create agent data dir: %w", err)
	}
	if err := r.register(ctx); err != nil {
		return err
	}

	heartbeatInterval := r.cfg.Agent.HeartbeatInterval
	if heartbeatInterval <= 0 {
		heartbeatInterval = 15 * time.Second
	}
	pollInterval := r.cfg.Agent.PollInterval
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}

	heartbeatTicker := time.NewTicker(heartbeatInterval)
	pollTicker := time.NewTicker(pollInterval)
	defer heartbeatTicker.Stop()
	defer pollTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-heartbeatTicker.C:
			if err := r.heartbeat(ctx); err != nil {
				r.logger.Warn("agent heartbeat failed", zap.Error(err))
			}
		case <-pollTicker.C:
			if err := r.pollAndExecute(ctx); err != nil {
				r.logger.Warn("agent task poll failed", zap.Error(err))
			}
		}
	}
}

func (r *Runtime) register(ctx context.Context) error {
	hostname, _ := os.Hostname()
	name := r.cfg.Node.Name
	if name == "" {
		name = hostname
	}
	req := dto.AgentRegisterReq{
		NodeKey:  r.nodeKey,
		Name:     name,
		Hostname: hostname,
		Roles:    strings.Join(agentRoles(r.cfg.Node.Roles), ","),
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Version:  r.cfg.Runner.DefaultVersion,
	}
	var resp dto.ServerNodeResp
	if err := r.post(ctx, "/api/v1/agent/register", req, &resp); err != nil {
		return fmt.Errorf("register agent: %w", err)
	}
	r.logger.Info("agent registered", zap.Uint("node_id", resp.ID), zap.String("node_key", resp.NodeKey))
	return nil
}

func (r *Runtime) heartbeat(ctx context.Context) error {
	req := dto.AgentHeartbeatReq{
		NodeKey: r.nodeKey,
	}
	var resp dto.ServerNodeResp
	return r.post(ctx, "/api/v1/agent/heartbeat", req, &resp)
}

func (r *Runtime) pollAndExecute(ctx context.Context) error {
	req := dto.PollRunnerTasksReq{
		NodeKey: r.nodeKey,
		Limit:   5,
	}
	var tasks []dto.RunnerTaskResp
	if err := r.post(ctx, "/api/v1/agent/tasks/poll", req, &tasks); err != nil {
		return err
	}
	for _, task := range tasks {
		r.logger.Info("executing runner task", zap.Uint("task_id", task.ID), zap.String("type", task.Type))
		report := r.manager.ExecuteTask(ctx, task)
		report.NodeKey = r.nodeKey
		if report.RunnerStatus == "" {
			report.RunnerStatus = statusForTask(task.Type, report.Status)
		}
		var resp dto.RunnerTaskResp
		if err := r.post(ctx, fmt.Sprintf("/api/v1/agent/tasks/%d/report", task.ID), report, &resp); err != nil {
			r.logger.Warn("report runner task failed", zap.Uint("task_id", task.ID), zap.Error(err))
		}
	}
	return nil
}

func (r *Runtime) post(ctx context.Context, path string, body any, target any) error {
	base, err := url.Parse(r.cfg.Agent.ServerURL)
	if err != nil {
		return fmt.Errorf("parse server URL: %w", err)
	}
	endpoint, err := base.Parse(path)
	if err != nil {
		return fmt.Errorf("build agent URL: %w", err)
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+r.cfg.Agent.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed: %s", resp.Status)
	}
	if target == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func defaultNodeKey() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}
	return strings.Join([]string{hostname, runtime.GOOS, runtime.GOARCH}, "-")
}

func agentRoles(roles []string) []string {
	seenAgent := false
	for _, role := range roles {
		if role == model.NodeRoleAgent {
			seenAgent = true
			break
		}
	}
	if seenAgent {
		return roles
	}
	return append(roles, model.NodeRoleAgent)
}

func statusForTask(taskType, taskStatus string) string {
	if taskStatus == model.RunnerTaskStatusFailed {
		return model.RunnerStatusError
	}
	switch taskType {
	case model.RunnerTaskTypeStop:
		return model.RunnerStatusStopped
	case model.RunnerTaskTypeRemove:
		return model.RunnerStatusRemoved
	default:
		return model.RunnerStatusRunning
	}
}
