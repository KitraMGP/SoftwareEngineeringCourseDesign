package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/account"
	"backend/internal/admin"
	"backend/internal/chat"
	"backend/internal/kb"
	"backend/internal/platform/auth"
	"backend/internal/platform/config"
	"backend/internal/platform/db"
	"backend/internal/platform/httpx"
	"backend/internal/platform/storage"
	"backend/internal/task"

	"github.com/go-chi/chi/v5"
)

func main() {
	if err := config.LoadEnvFiles(".env", ".env.local"); err != nil {
		panic(err)
	}

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx := context.Background()
	pool, err := db.Open(ctx, cfg.Database)
	if err != nil {
		logger.Error("failed to connect database", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	tokenManager := auth.NewTokenManager(
		cfg.Auth.JWTSecret,
		cfg.Auth.Issuer,
		cfg.Auth.AccessTokenTTL,
		cfg.Auth.RefreshTokenTTL,
	)

	accountRepo := account.NewRepository(pool)
	accountService := account.NewService(accountRepo, tokenManager)
	accountHandler := account.NewHandler(accountService, cfg.Auth)

	taskRepo := task.NewRepository(pool)
	taskService := task.NewService(taskRepo)
	storageService, err := storage.NewFromConfig(cfg.Storage)
	if err != nil {
		logger.Error("failed to initialize storage", slog.Any("error", err))
		os.Exit(1)
	}

	chatService := chat.NewService(
		chat.NewRepository(pool),
		chat.NewProvider(cfg.AI),
		chat.ServiceConfig{
			DefaultModel:       cfg.AI.DefaultChatModel,
			SystemPrompt:       cfg.AI.SystemPrompt,
			RequestTimeout:     cfg.AI.ChatTimeout,
			MaxHistoryMessages: cfg.AI.MaxHistoryMessages,
			Temperature:        cfg.AI.ChatTemperature,
		},
	)

	chatHandler := chat.NewHandler(chatService, cfg.AI.SSEHeartbeatInterval)
	kbHandler := kb.NewHandler(kb.NewService(kb.NewRepository(pool), taskService, storageService, cfg.Storage.MaxUploadBytes))
	adminHandler := admin.NewHandler()

	router := chi.NewRouter()
	router.Use(httpx.RequestID)
	router.Use(httpx.Recoverer(logger))
	router.Use(httpx.Logger(logger))

	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthz", httpx.Adapt(func(w http.ResponseWriter, r *http.Request) error {
			httpx.Success(w, http.StatusOK, map[string]any{
				"service": cfg.App.Name,
				"time":    time.Now().UTC(),
			})
			return nil
		}))

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", httpx.Adapt(accountHandler.Register))
			r.Post("/login", httpx.Adapt(accountHandler.Login))
			r.Post("/refresh", httpx.Adapt(accountHandler.Refresh))

			r.Group(func(r chi.Router) {
				r.Use(auth.Middleware(tokenManager))
				r.Post("/logout", httpx.Adapt(accountHandler.Logout))
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(tokenManager))

			r.Get("/users/me", httpx.Adapt(accountHandler.GetCurrentUser))
			r.Put("/users/me", httpx.Adapt(accountHandler.UpdateCurrentUser))
			r.Put("/users/me/password", httpx.Adapt(accountHandler.ChangePassword))

			chatHandler.RegisterRoutes(r)
			kbHandler.RegisterRoutes(r)

			r.Route("/admin", func(r chi.Router) {
				r.Use(auth.RequireRole(account.RoleAdmin))
				adminHandler.RegisterRoutes(r)
			})
		})
	})

	server := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		logger.Info("api server started", slog.String("addr", cfg.HTTP.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server stopped unexpectedly", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("shutting down api server")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown api server", slog.Any("error", err))
	}
}
