package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ray-d-song/merc/app/dto"
	"github.com/ray-d-song/merc/app/model"
	"github.com/ray-d-song/merc/app/repo"
)

var (
	ErrAgentTokenInvalid  = errors.New("invalid agent token")
	ErrServerNodeNotFound = errors.New("server node not found")
	ErrRunnerNotFound     = errors.New("runner not found")
	ErrRunnerTaskNotFound = errors.New("runner task not found")
)

type MercService struct {
	nodes    repo.ServerNodeRepository
	tokens   repo.AgentTokenRepository
	runners  repo.RunnerRepository
	tasks    repo.RunnerTaskRepository
	projects repo.ProjectRepository
}

func NewMercService(
	nodes repo.ServerNodeRepository,
	tokens repo.AgentTokenRepository,
	runners repo.RunnerRepository,
	tasks repo.RunnerTaskRepository,
	projects repo.ProjectRepository,
) *MercService {
	return &MercService{
		nodes:    nodes,
		tokens:   tokens,
		runners:  runners,
		tasks:    tasks,
		projects: projects,
	}
}

func (s *MercService) CreateAgentToken(ctx context.Context, req dto.CreateAgentTokenReq, creatorID uint, creatorName string) (*model.AgentToken, string, error) {
	plain, err := generateToken()
	if err != nil {
		return nil, "", err
	}
	token := &model.AgentToken{
		Name:            req.Name,
		TokenHash:       hashToken(plain),
		ExpiresAt:       req.ExpiresAt,
		CreaterID:       creatorID,
		CreaterUserName: creatorName,
	}
	if err := s.tokens.Create(ctx, token); err != nil {
		return nil, "", fmt.Errorf("create agent token: %w", err)
	}
	return token, plain, nil
}

func (s *MercService) AuthenticateAgentToken(ctx context.Context, plain string) (*model.AgentToken, error) {
	token, err := s.tokens.FindByHash(ctx, hashToken(plain))
	if err != nil {
		return nil, fmt.Errorf("find agent token: %w", err)
	}
	if token == nil || token.RevokedAt != nil {
		return nil, ErrAgentTokenInvalid
	}
	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		return nil, ErrAgentTokenInvalid
	}
	now := time.Now()
	token.LastUsedAt = &now
	if err := s.tokens.Update(ctx, token); err != nil {
		return nil, fmt.Errorf("update agent token usage: %w", err)
	}
	return token, nil
}

func (s *MercService) RegisterNode(ctx context.Context, req dto.AgentRegisterReq) (*model.ServerNode, error) {
	now := time.Now()
	node, err := s.nodes.FindByNodeKey(ctx, req.NodeKey)
	if err != nil {
		return nil, fmt.Errorf("find node: %w", err)
	}
	if node == nil {
		node = &model.ServerNode{
			NodeKey: req.NodeKey,
			Status:  model.NodeStatusOnline,
		}
	}
	node.Name = req.Name
	if node.Name == "" {
		node.Name = req.Hostname
	}
	node.Hostname = req.Hostname
	node.Roles = req.Roles
	node.OS = req.OS
	node.Arch = req.Arch
	node.Version = req.Version
	node.ResourceJSON = req.ResourceJSON
	node.Status = model.NodeStatusOnline
	node.LastHeartbeat = &now

	if node.ID == 0 {
		if err := s.nodes.Create(ctx, node); err != nil {
			return nil, fmt.Errorf("create node: %w", err)
		}
		return node, nil
	}
	if err := s.nodes.Update(ctx, node); err != nil {
		return nil, fmt.Errorf("update node: %w", err)
	}
	return node, nil
}

func (s *MercService) HeartbeatNode(ctx context.Context, req dto.AgentHeartbeatReq) (*model.ServerNode, error) {
	node, err := s.nodes.FindByNodeKey(ctx, req.NodeKey)
	if err != nil {
		return nil, fmt.Errorf("find node: %w", err)
	}
	if node == nil {
		return nil, ErrServerNodeNotFound
	}
	now := time.Now()
	node.Status = model.NodeStatusOnline
	node.ResourceJSON = req.ResourceJSON
	node.LastHeartbeat = &now
	if err := s.nodes.Update(ctx, node); err != nil {
		return nil, fmt.Errorf("update node heartbeat: %w", err)
	}

	for _, state := range req.Runners {
		if state.ID == 0 {
			continue
		}
		runner, err := s.runners.FindByID(ctx, state.ID)
		if err != nil {
			return nil, fmt.Errorf("find runner state: %w", err)
		}
		if runner == nil || runner.NodeID != node.ID {
			continue
		}
		runner.Status = state.Status
		runner.ProcessID = state.ProcessID
		runner.LastError = state.LastError
		runner.LastSeenAt = &now
		if err := s.runners.Update(ctx, runner); err != nil {
			return nil, fmt.Errorf("update runner state: %w", err)
		}
	}

	return node, nil
}

func (s *MercService) ListNodes(ctx context.Context, req dto.ListServerNodeReq) ([]*model.ServerNode, int64, error) {
	return s.nodes.List(ctx, req.GetOffset(), req.GetLimit(), req.Status, req.OS)
}

func (s *MercService) GetNode(ctx context.Context, id uint) (*model.ServerNode, error) {
	node, err := s.nodes.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find node: %w", err)
	}
	if node == nil {
		return nil, ErrServerNodeNotFound
	}
	return node, nil
}

func (s *MercService) CreateRunner(ctx context.Context, req dto.CreateRunnerReq, creatorID uint, creatorName string) (*model.Runner, error) {
	project, err := s.projects.FindByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("find project: %w", err)
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}
	node, err := s.nodes.FindByID(ctx, req.NodeID)
	if err != nil {
		return nil, fmt.Errorf("find node: %w", err)
	}
	if node == nil {
		return nil, ErrServerNodeNotFound
	}

	runner := &model.Runner{
		ProjectID:       req.ProjectID,
		NodeID:          req.NodeID,
		Name:            req.Name,
		Labels:          req.Labels,
		WorkDir:         req.WorkDir,
		OS:              node.OS,
		Arch:            node.Arch,
		Status:          model.RunnerStatusPending,
		CreaterID:       creatorID,
		CreaterUserName: creatorName,
	}
	if err := s.runners.Create(ctx, runner); err != nil {
		return nil, fmt.Errorf("create runner: %w", err)
	}

	payload, err := json.Marshal(map[string]string{
		"repositoryUrl":     project.RepositoryURL,
		"registrationToken": req.RegistrationToken,
		"name":              req.Name,
		"labels":            req.Labels,
		"workDir":           req.WorkDir,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal runner task payload: %w", err)
	}
	task := &model.RunnerTask{
		RunnerID:    runner.ID,
		NodeID:      node.ID,
		Type:        model.RunnerTaskTypeCreate,
		Status:      model.RunnerTaskStatusQueued,
		PayloadJSON: string(payload),
	}
	if err := s.tasks.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("create runner task: %w", err)
	}
	return runner, nil
}

func (s *MercService) ListRunners(ctx context.Context, req dto.ListRunnerReq) ([]*model.Runner, int64, error) {
	return s.runners.List(ctx, req.GetOffset(), req.GetLimit(), req.ProjectID, req.NodeID, req.Status)
}

func (s *MercService) GetRunner(ctx context.Context, id uint) (*model.Runner, error) {
	runner, err := s.runners.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find runner: %w", err)
	}
	if runner == nil {
		return nil, ErrRunnerNotFound
	}
	return runner, nil
}

func (s *MercService) EnqueueRunnerTask(ctx context.Context, runnerID uint, taskType string) (*model.RunnerTask, error) {
	runner, err := s.GetRunner(ctx, runnerID)
	if err != nil {
		return nil, err
	}
	task := &model.RunnerTask{
		RunnerID:    runner.ID,
		NodeID:      runner.NodeID,
		Type:        taskType,
		Status:      model.RunnerTaskStatusQueued,
		PayloadJSON: "{}",
	}
	if err := s.tasks.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("create runner task: %w", err)
	}
	return task, nil
}

func (s *MercService) PollRunnerTasks(ctx context.Context, req dto.PollRunnerTasksReq) ([]*model.RunnerTask, error) {
	node, err := s.nodes.FindByNodeKey(ctx, req.NodeKey)
	if err != nil {
		return nil, fmt.Errorf("find node: %w", err)
	}
	if node == nil {
		return nil, ErrServerNodeNotFound
	}
	tasks, err := s.tasks.ListQueuedForNode(ctx, node.ID, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("list node tasks: %w", err)
	}
	now := time.Now()
	for _, task := range tasks {
		task.Status = model.RunnerTaskStatusClaimed
		task.ClaimedAt = &now
		task.AttemptCount++
		if err := s.tasks.Update(ctx, task); err != nil {
			return nil, fmt.Errorf("claim task: %w", err)
		}
	}
	return tasks, nil
}

func (s *MercService) ReportRunnerTask(ctx context.Context, taskID uint, req dto.ReportRunnerTaskReq) (*model.RunnerTask, error) {
	node, err := s.nodes.FindByNodeKey(ctx, req.NodeKey)
	if err != nil {
		return nil, fmt.Errorf("find node: %w", err)
	}
	if node == nil {
		return nil, ErrServerNodeNotFound
	}
	task, err := s.tasks.FindByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("find task: %w", err)
	}
	if task == nil || task.NodeID != node.ID {
		return nil, ErrRunnerTaskNotFound
	}
	now := time.Now()
	task.Status = req.Status
	task.ResultJSON = req.ResultJSON
	task.LogSummary = RedactSecrets(req.LogSummary)
	task.LastError = RedactSecrets(req.LastError)
	if task.StartedAt == nil {
		task.StartedAt = &now
	}
	if isTerminalTaskStatus(req.Status) {
		task.FinishedAt = &now
		task.PayloadJSON = RedactTaskPayload(task.PayloadJSON)
	}
	if err := s.tasks.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("update task report: %w", err)
	}

	if task.RunnerID > 0 && req.RunnerStatus != "" {
		runner, err := s.runners.FindByID(ctx, task.RunnerID)
		if err != nil {
			return nil, fmt.Errorf("find task runner: %w", err)
		}
		if runner != nil {
			runner.Status = req.RunnerStatus
			runner.ProcessID = req.ProcessID
			runner.LastError = task.LastError
			runner.LastSeenAt = &now
			if err := s.runners.Update(ctx, runner); err != nil {
				return nil, fmt.Errorf("update task runner: %w", err)
			}
		}
	}

	return task, nil
}

func RedactSecrets(value string) string {
	if value == "" {
		return value
	}
	for _, marker := range []string{"registrationToken", "token", "Authorization", "Cookie"} {
		value = strings.ReplaceAll(value, marker, "[redacted]")
	}
	return value
}

func RedactTaskPayload(payload string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return RedactSecrets(payload)
	}
	for _, key := range []string{"registrationToken", "token", "authorization", "cookie"} {
		if _, ok := data[key]; ok {
			data[key] = "[redacted]"
		}
	}
	result, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(result)
}

func isTerminalTaskStatus(status string) bool {
	return status == model.RunnerTaskStatusSucceeded ||
		status == model.RunnerTaskStatusFailed ||
		status == model.RunnerTaskStatusCanceled
}

func generateToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return "merc_" + base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
