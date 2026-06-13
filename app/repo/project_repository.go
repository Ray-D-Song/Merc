package repo

import (
	"context"
	"errors"

	"github.com/ray-d-song/merc/app/infra/dbx"
	"github.com/ray-d-song/merc/app/model"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *model.Project) error
	FindByID(ctx context.Context, id uint) (*model.Project, error)
	List(ctx context.Context, offset, limit int) ([]*model.Project, int64, error)
	Update(ctx context.Context, project *model.Project) error
	Delete(ctx context.Context, id uint) error
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *projectRepository) FindByID(ctx context.Context, id uint) (*model.Project, error) {
	var project model.Project
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&project).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) List(ctx context.Context, offset, limit int) ([]*model.Project, int64, error) {
	var projects []*model.Project
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.Project{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).
		Scopes(dbx.Paginate(offset, limit)).
		Order("id DESC").
		Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

func (r *projectRepository) Update(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *projectRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Project{}, id).Error
}
