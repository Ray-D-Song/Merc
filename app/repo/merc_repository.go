package repo

import (
	"context"
	"errors"
	"time"

	"github.com/ray-d-song/merc/app/infra/dbx"
	"github.com/ray-d-song/merc/app/model"
	"gorm.io/gorm"
)

type ServerNodeRepository interface {
	Create(ctx context.Context, node *model.ServerNode) error
	FindByID(ctx context.Context, id uint) (*model.ServerNode, error)
	FindByNodeKey(ctx context.Context, nodeKey string) (*model.ServerNode, error)
	List(ctx context.Context, offset, limit int, status, os string) ([]*model.ServerNode, int64, error)
	Update(ctx context.Context, node *model.ServerNode) error
}

type serverNodeRepository struct {
	db *gorm.DB
}

func NewServerNodeRepository(db *gorm.DB) ServerNodeRepository {
	return &serverNodeRepository{db: db}
}

func (r *serverNodeRepository) Create(ctx context.Context, node *model.ServerNode) error {
	return r.db.WithContext(ctx).Create(node).Error
}

func (r *serverNodeRepository) FindByID(ctx context.Context, id uint) (*model.ServerNode, error) {
	var node model.ServerNode
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&node).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &node, err
}

func (r *serverNodeRepository) FindByNodeKey(ctx context.Context, nodeKey string) (*model.ServerNode, error) {
	var node model.ServerNode
	err := r.db.WithContext(ctx).Where("node_key = ?", nodeKey).First(&node).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &node, err
}

func (r *serverNodeRepository) List(ctx context.Context, offset, limit int, status, os string) ([]*model.ServerNode, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.ServerNode{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if os != "" {
		query = query.Where("os = ?", os)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var nodes []*model.ServerNode
	if err := query.Scopes(dbx.Paginate(offset, limit)).Order("id DESC").Find(&nodes).Error; err != nil {
		return nil, 0, err
	}
	return nodes, total, nil
}

func (r *serverNodeRepository) Update(ctx context.Context, node *model.ServerNode) error {
	return r.db.WithContext(ctx).Save(node).Error
}

type AgentTokenRepository interface {
	Create(ctx context.Context, token *model.AgentToken) error
	FindByHash(ctx context.Context, tokenHash string) (*model.AgentToken, error)
	Update(ctx context.Context, token *model.AgentToken) error
}

type agentTokenRepository struct {
	db *gorm.DB
}

func NewAgentTokenRepository(db *gorm.DB) AgentTokenRepository {
	return &agentTokenRepository{db: db}
}

func (r *agentTokenRepository) Create(ctx context.Context, token *model.AgentToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *agentTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*model.AgentToken, error) {
	var token model.AgentToken
	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &token, err
}

func (r *agentTokenRepository) Update(ctx context.Context, token *model.AgentToken) error {
	return r.db.WithContext(ctx).Save(token).Error
}

type RunnerRepository interface {
	Create(ctx context.Context, runner *model.Runner) error
	FindByID(ctx context.Context, id uint) (*model.Runner, error)
	List(ctx context.Context, offset, limit int, projectID, nodeID uint, status string) ([]*model.Runner, int64, error)
	Update(ctx context.Context, runner *model.Runner) error
}

type runnerRepository struct {
	db *gorm.DB
}

func NewRunnerRepository(db *gorm.DB) RunnerRepository {
	return &runnerRepository{db: db}
}

func (r *runnerRepository) Create(ctx context.Context, runner *model.Runner) error {
	return r.db.WithContext(ctx).Create(runner).Error
}

func (r *runnerRepository) FindByID(ctx context.Context, id uint) (*model.Runner, error) {
	var runner model.Runner
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&runner).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &runner, err
}

func (r *runnerRepository) List(ctx context.Context, offset, limit int, projectID, nodeID uint, status string) ([]*model.Runner, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Runner{})
	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	}
	if nodeID > 0 {
		query = query.Where("node_id = ?", nodeID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var runners []*model.Runner
	if err := query.Scopes(dbx.Paginate(offset, limit)).Order("id DESC").Find(&runners).Error; err != nil {
		return nil, 0, err
	}
	return runners, total, nil
}

func (r *runnerRepository) Update(ctx context.Context, runner *model.Runner) error {
	return r.db.WithContext(ctx).Save(runner).Error
}

type RunnerTaskRepository interface {
	Create(ctx context.Context, task *model.RunnerTask) error
	FindByID(ctx context.Context, id uint) (*model.RunnerTask, error)
	ListQueuedForNode(ctx context.Context, nodeID uint, limit int) ([]*model.RunnerTask, error)
	Update(ctx context.Context, task *model.RunnerTask) error
}

type runnerTaskRepository struct {
	db *gorm.DB
}

func NewRunnerTaskRepository(db *gorm.DB) RunnerTaskRepository {
	return &runnerTaskRepository{db: db}
}

func (r *runnerTaskRepository) Create(ctx context.Context, task *model.RunnerTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *runnerTaskRepository) FindByID(ctx context.Context, id uint) (*model.RunnerTask, error) {
	var task model.RunnerTask
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &task, err
}

func (r *runnerTaskRepository) ListQueuedForNode(ctx context.Context, nodeID uint, limit int) ([]*model.RunnerTask, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}
	var tasks []*model.RunnerTask
	err := r.db.WithContext(ctx).
		Where("node_id = ? AND status = ?", nodeID, model.RunnerTaskStatusQueued).
		Order("id ASC").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

func (r *runnerTaskRepository) Update(ctx context.Context, task *model.RunnerTask) error {
	task.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(task).Error
}
