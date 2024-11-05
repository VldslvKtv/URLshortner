package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"url_shortner/internal/config"
	"url_shortner/internal/http_server/handlers/redirect"
	"url_shortner/internal/http_server/handlers/url/save"
	"url_shortner/internal/lib/logger/sl"
	"url_shortner/internal/storage/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("start app", slog.String("env", cfg.Env))

	fmt.Println(cfg)

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	_ = storage

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer) // воостановление полсе паники (чтобы не падало приложение после 1 ошибки в хендлере)
	router.Use(middleware.URLFormat) //

	router.Post("/url", save.New(log, storage))
	router.Get("/{alias}", redirect.New(log, storage)) // TODO: Потестить
	// TODO: сДЕЛАТЬ ЕЩЕ И DELETE

	log.Info("starting  server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case "local":
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "dev":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "prod":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log

}
