package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	v1 "github.com/ray-d-song/merc/app/api/v1"
	"github.com/ray-d-song/merc/app/cron"
	"github.com/ray-d-song/merc/app/infra/config"
	"github.com/ray-d-song/merc/app/infra/static"
	"github.com/ray-d-song/merc/app/middleware"
	"github.com/ray-d-song/merc/app/model"
	"github.com/ray-d-song/merc/app/repo"
	"github.com/ray-d-song/merc/app/service"
	"github.com/ray-d-song/merc/app/store"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	app := fx.New(
		config.Module,
		fx.Provide(
			newRouter,
			newSessionStore,
			repo.NewUserRepository,
			repo.NewProjectRepository,
			repo.NewSessionRepository,
			service.NewAuthService,
			service.NewUserService,
			service.NewProjectService,
			v1.NewAuthHandler,
			v1.NewUserHanlder,
			v1.NewProjectHandler,
			cron.NewCron,
		),
		fx.Invoke(
			model.AutoMigrate,
			registerRoutes,
			runHTTPServer,
			startCron,
		),
	)

	app.Run()
}

func newRouter(cfg *config.AppConfig) chi.Router {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS(cfg.Server.FrontendDomain))
	return r
}

func newSessionStore(cfg *config.AppConfig, sessionRepo repo.SessionRepository) *store.SessionStore {
	maxAge := time.Duration(cfg.Session.MaxAgeMS) * time.Millisecond
	if maxAge <= 0 {
		maxAge = 14 * 24 * time.Hour
	}
	return store.NewSessionStore(sessionRepo, maxAge)
}

func registerRoutes(
	router chi.Router,
	authHandler *v1.AuthHandler,
	userHandler *v1.UserHandler,
	projectHandler *v1.ProjectHandler,
	sessionStore *store.SessionStore,
	userRepo repo.UserRepository,
	cfg *config.AppConfig,
) {
	router.Route("/api/v1", func(r chi.Router) {
		// Auth routes (mix of public and protected)
		r.Route("/auth", func(r chi.Router) {
			authHandler.RegisterRoutes(r, sessionStore, cfg.Session.Secret)
		})

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireSession(sessionStore, cfg.Session.Secret))
			v1.RegisterHealthRoutes(r)
			r.Route("/project", func(r chi.Router) {
				projectHandler.RegisterReadRoutes(r)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireAdmin(userRepo))
					projectHandler.RegisterAdminRoutes(r)
				})
			})
		})

		// Admin-only routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireSession(sessionStore, cfg.Session.Secret))
			r.Use(middleware.RequireAdmin(userRepo))
			r.Route("/user", func(r chi.Router) {
				userHandler.RegisterRoutes(r)
			})
		})
	})

	// Static files (SPA) - must be registered last as fallback
	staticFS := static.MustGetFS()
	router.Handle("/*", middleware.SPAHandler(staticFS))
}

func runHTTPServer(lc fx.Lifecycle, logger *zap.Logger, cfg *config.AppConfig, router chi.Router) {
	addr := net.JoinHostPort(cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting HTTP server", zap.String("addr", addr))
			go func() {
				if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Error("http server stopped unexpectedly", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping HTTP server")
			return server.Shutdown(ctx)
		},
	})
}

func startCron(lc fx.Lifecycle, cron *cron.Cron) {
	cron.Register(lc)
}
