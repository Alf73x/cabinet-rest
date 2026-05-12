package main

import (
	"CabinetREST/internal/config"
	"CabinetREST/internal/lib/logger/sl"
	"CabinetREST/internal/storage/sqlite"
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"fmt"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"CabinetREST/internal/http-server/handlers"
	territories "CabinetREST/internal/http-server/handlers"
	mwLogger "CabinetREST/internal/http-server/middleware/logger"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// config: cleanenv
	cfg := config.MustLoad()
	fmt.Println(cfg)

	// init logger: slog
	log := setupLogger(cfg.Env)
	log.Info("stating Cabinet REST-Server", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	// init storage: sqlite
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	// init router
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
		},
		MaxAge: 300,
	}))

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	// router.Post(handlers.Url_Territories, countries.New(log, storage))
	// router.Get(handlers.Url_Territories, countries.New(log, storage))
	router.Route(handlers.Url_Territories, func(r chi.Router) {
		r.Use(middleware.BasicAuth("CabinetREST", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
			// can be more users here
		}))

		r.Get("/search", territories.NewTerritorySearch(log, storage))
		r.Get("/path", territories.NewTerritoryPath(log, storage))
		r.Get("/{"+handlers.Url_Territories_ID+":[0-9]+}", territories.NewTerritoryChildren(log, storage))
	})

	router.Route(handlers.Url_Seasons, func(r chi.Router) {
		r.Use(middleware.BasicAuth("CabinetREST", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Get("/", handlers.NewSeasons(log, storage))
	})

	router.Route(handlers.Url_Sports, func(r chi.Router) {
		r.Use(middleware.BasicAuth("CabinetREST", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Get("/", handlers.NewSports(log, storage))
	})

	router.Route(handlers.Url_Teams, func(r chi.Router) {
		r.Use(middleware.BasicAuth("CabinetREST", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Get("/", handlers.NewTeams(log, storage))
	})

	router.Route(handlers.Url_Team_Matches, func(r chi.Router) {
		r.Use(middleware.BasicAuth("CabinetREST", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Get("/", handlers.NewTeamMatches(log, storage))
	})

	// run server
	log.Info("starting server", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()
	<-done
	log.Info("stopping server")

	// TODO: move timeout to config
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	// TODO: close storage

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
