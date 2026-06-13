package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ray-d-song/merc/app/dto"
	"github.com/ray-d-song/merc/app/infra/httpx"
	"github.com/ray-d-song/merc/app/model"
	"github.com/ray-d-song/merc/app/service"
)

type ProjectHandler struct {
	service *service.ProjectService
}

func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: svc}
}

func (h *ProjectHandler) RegisterRoutes(r chi.Router) {
	r.Post("/list", h.handleList)
	r.Get("/{id}", h.handleGet)
	r.Post("/create", h.handleCreate)
	r.Put("/update", h.handleUpdate)
	r.Delete("/{id}", h.handleDelete)
}

func (h *ProjectHandler) RegisterReadRoutes(r chi.Router) {
	r.Post("/list", h.handleList)
	r.Get("/{id}", h.handleGet)
}

func (h *ProjectHandler) RegisterAdminRoutes(r chi.Router) {
	r.Post("/create", h.handleCreate)
	r.Put("/update", h.handleUpdate)
	r.Delete("/{id}", h.handleDelete)
}

func (h *ProjectHandler) handleList(w http.ResponseWriter, r *http.Request) {
	var req dto.ListProjectReq
	if err := httpx.BindJSON(r, &req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	projects, total, err := h.service.ListProjects(r.Context(), req)
	if err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list projects"})
		return
	}

	items := make([]dto.ProjectResp, len(projects))
	for i, p := range projects {
		items[i] = toProjectResp(p)
	}

	httpx.WriteJSON(w, http.StatusOK, dto.ListProjectResp{
		PaginatedResponse: dto.PaginatedResponse[dto.ProjectResp]{
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

func (h *ProjectHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	project, err := h.service.GetProject(r.Context(), id)
	if err != nil {
		if err == service.ErrProjectNotFound {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get project"})
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProjectResp(project))
}

func (h *ProjectHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProjectReq
	if err := httpx.BindJSON(r, &req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	userID, err := httpx.GetUserID(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	project, err := h.service.CreateProject(r.Context(), req, userID, "")
	if err != nil {
		if errors.Is(err, service.ErrInvalidRepositoryURL) {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create project"})
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toProjectResp(project))
}

func (h *ProjectHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateProjectReq
	if err := httpx.BindJSON(r, &req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	userID, err := httpx.GetUserID(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	project, err := h.service.UpdateProject(r.Context(), req, userID, "")
	if err != nil {
		if err == service.ErrProjectNotFound {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrInvalidRepositoryURL) {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update project"})
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProjectResp(project))
}

func (h *ProjectHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteProject(r.Context(), id); err != nil {
		if err == service.ErrProjectNotFound {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete project"})
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "project deleted successfully"})
}

func parseIDParam(r *http.Request) (uint, error) {
	raw := chi.URLParam(r, "id")
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

func toProjectResp(p *model.Project) dto.ProjectResp {
	return dto.ProjectResp{
		ID:                  p.ID,
		Name:                p.Name,
		Provider:            p.Provider,
		RepositoryURL:       p.RepositoryURL,
		Owner:               p.Owner,
		Repo:                p.Repo,
		Description:         p.Description,
		Status:              p.Status,
		CreaterID:           p.CreaterID,
		CreaterUserName:     p.CreaterUserName,
		LastUpdaterID:       p.LastUpdaterID,
		LastUpdaterUserName: p.LastUpdaterUserName,
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
	}
}
