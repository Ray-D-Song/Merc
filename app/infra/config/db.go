package config

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewDB wires a gorm.DB using the configured driver and zap logger.
func NewDB(cfg *AppConfig, logger *zap.Logger) (*gorm.DB, error) {
	dialector, err := newDialector(cfg.Database)
	if err != nil {
		return nil, err
	}

	gormLogger := gormlogger.New(
		zap.NewStdLog(logger.Named("gorm")),
		gormlogger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  gormlogger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(dialector, &gorm.Config{Logger: gormLogger})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("obtain sql.DB: %w", err)
	}

	if cfg.Database.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	}
	if cfg.Database.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	}
	if cfg.Database.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	}

	return db, nil
}

func newDialector(cfg DatabaseConfig) (gorm.Dialector, error) {
	switch cfg.Type {
	case "sqlite":
		return sqlite.Open(cfg.URL), nil
	default:
		return nil, fmt.Errorf("unsupported database type %q", cfg.Type)
	}
}
