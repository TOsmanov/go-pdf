package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/TOsmanov/go-pdf/internal/config"
	handlers "github.com/TOsmanov/go-pdf/internal/http-server"
	logger "github.com/TOsmanov/go-pdf/internal/http-server/middleware"
	"github.com/TOsmanov/go-pdf/internal/lib/utils"
	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	var err error
	cfg := config.MustLoad()
	log := SetupLogger(cfg.Env)
	err = cfg.InitTemplates()
	if err != nil {
		log.Error("error init templates", slog.Any("err", err))
	}

	sem := utils.Semaphore{
		C: make(chan struct{}, cfg.Limit),
	}

	log.Info("starting converter service", slog.String("environment", cfg.Env))
	log.Debug("DEBUG messages are enabled")

	log.Info("enabled limit of simultaneously running goroutines", slog.Int("limit", cfg.Limit))

	router := chi.NewRouter()

	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.HandleFunc("/pdf", handlers.PDFHandler(log, cfg, &sem))
	router.HandleFunc("/docx", handlers.DOCXHandler(log, cfg))
	if cfg.Face.FaceEnable {
		router.HandleFunc("/face", handlers.FaceHandler(log, cfg))
	}
	router.HandleFunc("/reload", handlers.ReloadHandler(cfg))

	cfg.Reload = make(chan error, 1)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	log.Info("starting server", slog.String("address", srv.Addr))

	go func() {
		defer srv.Close()
		if err = srv.ListenAndServe(); err != nil {
			log.Error("failed to serve server", slog.Any("error", err))
		}
	}()

	go func() {
		if err = <-cfg.Reload; err != nil {
			Reload(err, log)
		}
	}()

	<-done
	log.Info("stopping server")

	log.Info("server stopped")
}

func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case "local":
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "docker-debug":
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "prod":
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}

func Reload(err error, log *slog.Logger) {
	log.Error("Stopping server", slog.Any("error", err))
	os.Exit(1)
}
