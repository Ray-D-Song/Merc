package v1

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/ray-d-song/merc/app/dto"
	"github.com/ray-d-song/merc/app/infra/httpx"
	"github.com/ray-d-song/merc/app/model"
	"github.com/ray-d-song/merc/app/service"
)

type MercHandler struct {
	service *service.MercService
	auth    *service.AuthService
}

func NewMercHandler(svc *service.MercService, auth *service.AuthService) *MercHandler {
	return &MercHandler{service: svc, auth: auth}
}

func (h *MercHandler) RegisterServerNodeRoutes(r chi.Router) {
	r.Post("/list", h.handleListNodes)
	r.Get("/{id}", h.handleGetNode)
	r.Post("/token/create", h.handleCreateAgentToken)
}

func (h *MercHandler) RegisterRunnerRoutes(r chi.Router) {
	r.Post("/list", h.handleListRunners)
	r.Get("/{id}", h.handleGetRunner)
	r.Post("/create", h.handleCreateRunner)
	r.Post("/{id}/start", h.handleStartRunner)
	r.Post("/{id}/stop", h.handleStopRunner)
	r.Post("/{id}/remove", h.handleRemoveRunner)
}

func (h *MercHandler) RegisterAgentRoutes(r chi.Router) {
	r.Use(h.requireAgentToken)
	r.Post("/register", h.handleAgentRegister)
	r.Post("/heartbeat", h.handleAgentHeartbeat)
	r.Post("/tasks/poll", h.handlePollRunnerTasks)
	r.Post("/tasks/{id}/report", h.handleReportRunnerTask)
}

func (h *MercHandler) requireAgentToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" {
			httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing agent token"})
			return
		}
		if _, err := h.service.AuthenticateAgentToken(r.Context(), token); err != nil {
			httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid agent token"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *MercHandler) handleListNodes(w http.ResponseWriter, r *http.Request) {
	var req dto.ListServerNodeReq
	if !bindAndValidate(w, r, &req) {
		return
	}
	nodes, total, err := h.service.ListNodes(r.Context(), req)
	if err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list server nodes"})
		return
	}
	items := make([]dto.ServerNodeResp, len(nodes))
	for i, node := range nodes {
		items[i] = toServerNodeResp(node)
	}
	httpx.WriteJSON(w, http.StatusOK, dto.ListServerNodeResp{
		PaginatedResponse: dto.PaginatedResponse[dto.ServerNodeResp]{
			Data: items,
			Pagination: dto.PaginationMeta{
				Page:       req.Page,
				PageSize:   req.PageSize,
				Total:      total,
				TotalPages: req.CalculateTotalPages(total),
			},
		},
	})
}

func (h *MercHandler) handleGetNode(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	node, err := h.service.GetNode(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrServerNodeNotFound) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get server node"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toServerNodeResp(node))
}

func (h *MercHandler) handleCreateAgentToken(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateAgentTokenReq
	if !bindAndValidate(w, r, &req) {
		return
	}
	userID, username, ok := currentUserIdentity(w, r, h.auth)
	if !ok {
		return
	}
	token, plain, err := h.service.CreateAgentToken(r.Context(), req, userID, username)
	if err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create agent token"})
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, dto.CreateAgentTokenResp{
		ID:        token.ID,
		Name:      token.Name,
		Token:     plain,
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
	})
}

func (h *MercHandler) handleAgentRegister(w http.ResponseWriter, r *http.Request) {
	var req dto.AgentRegisterReq
	if !bindAndValidate(w, r, &req) {
		return
	}
	node, err := h.service.RegisterNode(r.Context(), req)
	if err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to register node"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toServerNodeResp(node))
}

func (h *MercHandler) handleAgentHeartbeat(w http.ResponseWriter, r *http.Request) {
	var req dto.AgentHeartbeatReq
	if !bindAndValidate(w, r, &req) {
		return
	}
	node, err := h.service.HeartbeatNode(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrServerNodeNotFound) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update heartbeat"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toServerNodeResp(node))
}

func (h *MercHandler) handlePollRunnerTasks(w http.ResponseWriter, r *http.Request) {
	var req dto.PollRunnerTasksReq
	if !bindAndValidate(w, r, &req) {
		return
	}
	tasks, err := h.service.PollRunnerTasks(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrServerNodeNotFound) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to poll runner tasks"})
		return
	}
	items := make([]dto.RunnerTaskResp, len(tasks))
	for i, task := range tasks {
		items[i] = toRunnerTaskResp(task)
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *MercHandler) handleReportRunnerTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req dto.ReportRunnerTaskReq
	if !bindAndValidate(w, r, &req) {
		return
	}
	task, err := h.service.ReportRunnerTask(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrRunnerTaskNotFound) || errors.Is(err, service.ErrServerNodeNotFound) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to report runner task"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toRunnerTaskResp(task))
}

func (h *MercHandler) handleCreateRunner(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRunnerReq
	if !bindAndValidate(w, r, &req) {
		return
	}
	userID, username, ok := currentUserIdentity(w, r, h.auth)
	if !ok {
		return
	}
	runner, err := h.service.CreateRunner(r.Context(), req, userID, username)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProjectNotFound), errors.Is(err, service.ErrServerNodeNotFound):
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create runner"})
		}
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toRunnerResp(runner))
}

func (h *MercHandler) handleListRunners(w http.ResponseWriter, r *http.Request) {
	var req dto.ListRunnerReq
	if !bindAndValidate(w, r, &req) {
		return
	}
	runners, total, err := h.service.ListRunners(r.Context(), req)
	if err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list runners"})
		return
	}
	items := make([]dto.RunnerResp, len(runners))
	for i, runner := range runners {
		items[i] = toRunnerResp(runner)
	}
	httpx.WriteJSON(w, http.StatusOK, dto.ListRunnerResp{
		PaginatedResponse: dto.PaginatedResponse[dto.RunnerResp]{
			Data: items,
			Pagination: dto.PaginationMeta{
				Page:       req.Page,
				PageSize:   req.PageSize,
				Total:      total,
				TotalPages: req.CalculateTotalPages(total),
			},
		},
	})
}

func (h *MercHandler) handleGetRunner(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	runner, err := h.service.GetRunner(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrRunnerNotFound) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get runner"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toRunnerResp(runner))
}

func (h *MercHandler) handleStartRunner(w http.ResponseWriter, r *http.Request) {
	h.handleRunnerTaskCommand(w, r, model.RunnerTaskTypeStart)
}

func (h *MercHandler) handleStopRunner(w http.ResponseWriter, r *http.Request) {
	h.handleRunnerTaskCommand(w, r, model.RunnerTaskTypeStop)
}

func (h *MercHandler) handleRemoveRunner(w http.ResponseWriter, r *http.Request) {
	h.handleRunnerTaskCommand(w, r, model.RunnerTaskTypeRemove)
}

func (h *MercHandler) handleRunnerTaskCommand(w http.ResponseWriter, r *http.Request, taskType string) {
	id, err := parseIDParam(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	task, err := h.service.EnqueueRunnerTask(r.Context(), id, taskType)
	if err != nil {
		if errors.Is(err, service.ErrRunnerNotFound) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to enqueue runner task"})
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toRunnerTaskResp(task))
}

func bindAndValidate(w http.ResponseWriter, r *http.Request, req any) bool {
	if err := httpx.BindJSON(r, req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return false
	}
	if err := httpx.ValidateStruct(req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return false
	}
	return true
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func toServerNodeResp(node *model.ServerNode) dto.ServerNodeResp {
	return dto.ServerNodeResp{
		ID:            node.ID,
		NodeKey:       node.NodeKey,
		Name:          node.Name,
		Hostname:      node.Hostname,
		Roles:         node.Roles,
		OS:            node.OS,
		Arch:          node.Arch,
		Version:       node.Version,
		Status:        node.Status,
		ResourceJSON:  node.ResourceJSON,
		LastHeartbeat: node.LastHeartbeat,
		CreatedAt:     node.CreatedAt,
		UpdatedAt:     node.UpdatedAt,
	}
}

func toRunnerResp(runner *model.Runner) dto.RunnerResp {
	return dto.RunnerResp{
		ID:                  runner.ID,
		ProjectID:           runner.ProjectID,
		NodeID:              runner.NodeID,
		Name:                runner.Name,
		Labels:              runner.Labels,
		WorkDir:             runner.WorkDir,
		InstallDir:          runner.InstallDir,
		OS:                  runner.OS,
		Arch:                runner.Arch,
		Status:              runner.Status,
		ProcessID:           runner.ProcessID,
		LastError:           runner.LastError,
		LastSeenAt:          runner.LastSeenAt,
		CreaterID:           runner.CreaterID,
		CreaterUserName:     runner.CreaterUserName,
		LastUpdaterID:       runner.LastUpdaterID,
		LastUpdaterUserName: runner.LastUpdaterUserName,
		CreatedAt:           runner.CreatedAt,
		UpdatedAt:           runner.UpdatedAt,
	}
}

func toRunnerTaskResp(task *model.RunnerTask) dto.RunnerTaskResp {
	return dto.RunnerTaskResp{
		ID:           task.ID,
		RunnerID:     task.RunnerID,
		NodeID:       task.NodeID,
		Type:         task.Type,
		Status:       task.Status,
		PayloadJSON:  task.PayloadJSON,
		ResultJSON:   task.ResultJSON,
		LogSummary:   task.LogSummary,
		LastError:    task.LastError,
		AttemptCount: task.AttemptCount,
		ClaimedAt:    task.ClaimedAt,
		StartedAt:    task.StartedAt,
		FinishedAt:   task.FinishedAt,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
	}
}
