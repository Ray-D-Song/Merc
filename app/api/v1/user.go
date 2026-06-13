package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ray-d-song/merc/app/dto"
	"github.com/ray-d-song/merc/app/infra/httpx"
	"github.com/ray-d-song/merc/app/model"
	"github.com/ray-d-song/merc/app/service"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHanlder(service *service.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) RegisterRoutes(router chi.Router) {
	router.Get("/list", h.handleListUser)
	router.Post("/create", h.handleCreateUser)
	router.Post("/disable", h.handleDisableUser)
	router.Post("/enable", h.handleEnableUser)
	router.Post("/update-role", h.handleUpdateUserRole)
}

func (h *UserHandler) handleListUser(w http.ResponseWriter, r *http.Request) {
	// Bind request parameters from query string or JSON body
	var req dto.ListUserRequest
	if err := httpx.BindQuery(r, &req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Call service layer to get user list and total count
	users, total, err := h.service.ListUser(r.Context(), req)
	if err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list users"})
		return
	}

	// Convert domain models to DTOs (handler layer responsibility)
	userDTOs := make([]dto.UserDTO, len(users))
	for i, user := range users {
		userDTOs[i] = toUserDTO(user)
	}

	// Build paginated response
	resp := dto.ListUserResponse{
		PaginatedResponse: dto.PaginatedResponse[dto.UserDTO]{
			Data: userDTOs,
			Pagination: dto.PaginationMeta{
				Page:       req.Page,
				PageSize:   req.PageSize,
				Total:      total,
				TotalPages: req.CalculateTotalPages(total),
			},
		},
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	// Bind request body
	var req dto.CreateUserRequest
	if err := httpx.BindJSON(r, &req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Call service layer to create user
	user, err := h.service.CreateUser(r.Context(), req)
	if err != nil {
		switch err {
		case service.ErrUserAlreadyExists:
			httpx.WriteJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		default:
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		}
		return
	}

	// Convert to DTO and return
	resp := dto.CreateUserResponse{
		UserDTO: toUserDTO(user),
	}

	httpx.WriteJSON(w, http.StatusCreated, resp)
}

func (h *UserHandler) handleDisableUser(w http.ResponseWriter, r *http.Request) {
	// Bind request body
	var req dto.UpdateUserStatusRequest
	if err := httpx.BindJSON(r, &req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Call service layer to disable user
	err := h.service.DisableUser(r.Context(), req.UserID)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to disable user"})
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "user disabled successfully"})
}

func (h *UserHandler) handleEnableUser(w http.ResponseWriter, r *http.Request) {
	// Bind request body
	var req dto.UpdateUserStatusRequest
	if err := httpx.BindJSON(r, &req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Call service layer to enable user
	err := h.service.EnableUser(r.Context(), req.UserID)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to enable user"})
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "user enabled successfully"})
}

func (h *UserHandler) handleUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	// Bind request body
	var req dto.UpdateUserRoleRequest
	if err := httpx.BindJSON(r, &req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Call service layer to update user role
	err := h.service.UpdateUserRole(r.Context(), req.UserID, req.Role)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update user role"})
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "user role updated successfully"})
}

// toUserDTO converts a domain model User to a UserDTO
func toUserDTO(user *model.User) dto.UserDTO {
	return dto.UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Realname:  user.Realname,
		Email:     user.Email,
		Role:      user.Role,
		IsAdmin:   user.IsAdmin,
		IsEnabled: user.IsEnabled,
	}
}
