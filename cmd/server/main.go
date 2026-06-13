package main

import (
	"context"
	"errors"
	"flag"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/ray-d-song/merc/app/agent"
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
	switch command := firstArg(); command {
	case "", "server":
		runServer(false)
	case "agent":
		runAgentOnly(os.Args[2:])
	case "start":
		runStart(os.Args[2:])
	default:
		runServer(false)
	}
}

func runServer(enableAgent bool) {
	options := []fx.Option{
		config.Module,
		fx.Provide(
			newRouter,
			newSessionStore,
			repo.NewAgentTokenRepository,
			repo.NewRunnerRepository,
			repo.NewRunnerTaskRepository,
			repo.NewServerNodeRepository,
			repo.NewUserRepository,
			repo.NewProjectRepository,
			repo.NewSessionRepository,
			service.NewAuthService,
			service.NewMercService,
			service.NewUserService,
			service.NewProjectService,
			v1.NewAuthHandler,
			v1.NewMercHandler,
			v1.NewUserHanlder,
			v1.NewProjectHandler,
			agent.NewRuntime,
			cron.NewCron,
		),
		fx.Invoke(
			model.AutoMigrate,
			registerRoutes,
			runHTTPServer,
			startCron,
		),
	}
	if enableAgent {
		options = append(options, fx.Invoke(startAgentRuntime))
	}

	app := fx.New(options...)
	app.Run()
}

func runStart(args []string) {
	flags := flag.NewFlagSet("start", flag.ExitOnError)
	enableAgent := flags.Bool("agent", false, "run a worker agent in the same process")
	serverURL := flags.String("server-url", "", "control node URL for the local agent")
	token := flags.String("token", "", "agent join token")
	_ = flags.Parse(args)
	applyAgentFlagOverrides(*serverURL, *token)
	runServer(*enableAgent)
}

func runAgentOnly(args []string) {
	flags := flag.NewFlagSet("agent", flag.ExitOnError)
	serverURL := flags.String("server-url", "", "control node URL")
	token := flags.String("token", "", "agent join token")
	dataDir := flags.String("data-dir", "", "agent data directory")
	_ = flags.Parse(args)
	applyAgentFlagOverrides(*serverURL, *token)
	if *dataDir != "" {
		_ = os.Setenv("MERC_AGENT_DATA_DIR", *dataDir)
	}

	v, err := config.NewViper()
	if err != nil {
		panic(err)
	}
	cfg, err := config.NewAppConfig(v)
	if err != nil {
		panic(err)
	}
	logger, err := config.NewLogger(cfg)
	if err != nil {
		panic(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := agent.NewRuntime(cfg, logger).Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Fatal("agent stopped", zap.Error(err))
	}
}

func firstArg() string {
	if len(os.Args) < 2 {
		return ""
	}
	return os.Args[1]
}

func applyAgentFlagOverrides(serverURL, token string) {
	if serverURL != "" {
		_ = os.Setenv("MERC_AGENT_SERVER_URL", serverURL)
	}
	if token != "" {
		_ = os.Setenv("MERC_AGENT_TOKEN", token)
	}
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
	mercHandler *v1.MercHandler,
	sessionStore *store.SessionStore,
	userRepo repo.UserRepository,
	cfg *config.AppConfig,
) {
	router.Route("/api/v1", func(r chi.Router) {
		// Auth routes (mix of public and protected)
		r.Route("/auth", func(r chi.Router) {
			authHandler.RegisterRoutes(r, sessionStore, cfg.Session.Secret)
		})

		// Agent routes use bearer tokens instead of browser sessions.
		r.Route("/agent", func(r chi.Router) {
			mercHandler.RegisterAgentRoutes(r)
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
			r.Route("/server", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireAdmin(userRepo))
					mercHandler.RegisterServerNodeRoutes(r)
				})
			})
			r.Route("/runner", func(r chi.Router) {
				r.Use(middleware.RequireAdmin(userRepo))
				mercHandler.RegisterRunnerRoutes(r)
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

func startAgentRuntime(lc fx.Lifecycle, runtime *agent.Runtime, logger *zap.Logger) {
	var cancel context.CancelFunc
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			agentCtx, stop := context.WithCancel(context.Background())
			cancel = stop
			go func() {
				if err := runtime.Run(agentCtx); err != nil && !errors.Is(err, context.Canceled) {
					logger.Error("agent runtime stopped", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if cancel != nil {
				cancel()
			}
			return nil
		},
	})
}
