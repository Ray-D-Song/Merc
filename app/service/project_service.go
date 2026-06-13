package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ray-d-song/merc/app/dto"
	"github.com/ray-d-song/merc/app/githubrepo"
	"github.com/ray-d-song/merc/app/model"
	"github.com/ray-d-song/merc/app/repo"
)

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrInvalidRepositoryURL = errors.New("invalid repository URL")
)

type ProjectService struct {
	repo repo.ProjectRepository
}

func NewProjectService(repo repo.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) CreateProject(ctx context.Context, req dto.CreateProjectReq, creatorID uint, creatorName string) (*model.Project, error) {
	repository, err := githubrepo.Parse(req.RepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRepositoryURL, err)
	}
	name := req.Name
	if name == "" {
		name = repository.Repo
	}
	project := &model.Project{
		Name:            name,
		Provider:        "github",
		RepositoryURL:   repository.URL,
		Owner:           repository.Owner,
		Repo:            repository.Repo,
		Description:     req.Description,
		Status:          1,
		CreaterID:       creatorID,
		CreaterUserName: creatorName,
	}
	if err := s.repo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return project, nil
}

func (s *ProjectService) GetProject(ctx context.Context, id uint) (*model.Project, error) {
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find project: %w", err)
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}
	return project, nil
}

func (s *ProjectService) ListProjects(ctx context.Context, req dto.ListProjectReq) ([]*model.Project, int64, error) {
	return s.repo.List(ctx, req.GetOffset(), req.GetLimit())
}

func (s *ProjectService) UpdateProject(ctx context.Context, req dto.UpdateProjectReq, updaterID uint, updaterName string) (*model.Project, error) {
	project, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("find project: %w", err)
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	if req.RepositoryURL != "" {
		repository, err := githubrepo.Parse(req.RepositoryURL)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidRepositoryURL, err)
		}
		project.RepositoryURL = repository.URL
		project.Owner = repository.Owner
		project.Repo = repository.Repo
		project.Provider = "github"
	}
	project.Name = req.Name
	project.Description = req.Description
	project.Status = req.Status
	project.LastUpdaterID = updaterID
	project.LastUpdaterUserName = updaterName

	if err := s.repo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("update project: %w", err)
	}
	return project, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, id uint) error {
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find project: %w", err)
	}
	if project == nil {
		return ErrProjectNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
}
