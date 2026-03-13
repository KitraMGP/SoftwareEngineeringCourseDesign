package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"backend/internal/kb"
	"backend/internal/platform/config"
	"backend/internal/platform/db"
	"backend/internal/platform/storage"
	"backend/internal/task"
	"backend/internal/worker"
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

	taskRepo := task.NewRepository(pool)
	taskService := task.NewService(taskRepo)
	kbRepo := kb.NewRepository(pool)
	storageService, err := storage.NewFromConfig(cfg.Storage)
	if err != nil {
		logger.Error("failed to initialize storage", slog.Any("error", err))
		os.Exit(1)
	}

	worker := worker.New(logger, taskService, kbRepo, storageService, cfg.Task.PollInterval)

	runCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := worker.Run(runCtx); err != nil {
		logger.Error("worker stopped unexpectedly", slog.Any("error", err))
		os.Exit(1)
	}
}
