package main

import (
	"crypto_tracker/config"
	"crypto_tracker/internal/handlers/add"
	"crypto_tracker/internal/handlers/get"
	"crypto_tracker/internal/handlers/remove"
	"crypto_tracker/internal/storage/pg"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	config := config.MustLoad()
	log := setupLogger(config.Env)
	log.Info("start app", slog.String("env", config.Env))

	_ = log

	storage, err := pg.New(&config)
	if err != nil {
		log.Error("failed to init storage")
		os.Exit(1)
	}
	_ = storage
	log.Info("migration run is completed")
	defer storage.Close()

	router := chi.NewRouter()
	router.Use(middleware.Recoverer) // воостановление после паники (чтобы не падало приложение после 1 ошибки в хендлере)
	router.Use(middleware.URLFormat)

	// Swagger UI
	// router.Get("/swagger/*", httpSwagger.WrapHandler)

	// Настройка роутинга
	router.Post("/currency/add", add.New(log, config.ExtAPIUrl, config.APIKey, storage))
	router.Post("/currency/remove", remove.New(log))
	router.Get("/currency/price", get.New(log, storage))

	log.Info("starting server", slog.String("address", config.Address))

	srv := &http.Server{
		Addr:         config.Address,
		Handler:      router,
		ReadTimeout:  config.HTTPServer.Timeout,
		WriteTimeout: config.HTTPServer.Timeout,
		IdleTimeout:  config.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT) // graceful shutdown
	check := <-stop

	log.Debug("server stopped", slog.String("signal", check.String()))
}

// Настройка уровня логирования
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
