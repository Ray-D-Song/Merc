package cron

import (
	"context"
	"sync"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Job is a recurring background task managed by Cron.
type Job interface {
	Name() string
	Interval() time.Duration
	Run(ctx context.Context) error
}

type Cron struct {
	logger *zap.Logger
	jobs   []Job
}

func NewCron(logger *zap.Logger) *Cron {
	return &Cron{logger: logger}
}

func (c *Cron) AddJob(job Job) {
	c.jobs = append(c.jobs, job)
}

func (c *Cron) Register(lc fx.Lifecycle) {
	var cancel context.CancelFunc
	var wg sync.WaitGroup

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if len(c.jobs) == 0 {
				c.logger.Info("cron started with no registered jobs")
				return nil
			}

			runCtx, stop := context.WithCancel(context.Background())
			cancel = stop

			for _, job := range c.jobs {
				wg.Add(1)
				go c.runJob(runCtx, &wg, job)
			}

			c.logger.Info("cron started", zap.Int("jobCount", len(c.jobs)))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if cancel != nil {
				cancel()
			}
			wg.Wait()
			c.logger.Info("cron stopped")
			return nil
		},
	})
}

func (c *Cron) runJob(ctx context.Context, wg *sync.WaitGroup, job Job) {
	defer wg.Done()

	interval := job.Interval()
	if interval <= 0 {
		c.logger.Warn("cron job skipped because interval is not positive", zap.String("job", job.Name()))
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := job.Run(ctx); err != nil {
				c.logger.Error("cron job failed", zap.String("job", job.Name()), zap.Error(err))
			}
		}
	}
}
