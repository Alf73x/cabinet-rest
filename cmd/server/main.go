package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"CabinetREST/internal/config"
	"CabinetREST/internal/http-server/handlers"
	auth "CabinetREST/internal/http-server/handlers/auth"
	appmiddleware "CabinetREST/internal/http-server/middleware"
	mwLogger "CabinetREST/internal/http-server/middleware/logger"
	jwtservice "CabinetREST/internal/jwt"
	"CabinetREST/internal/lib/logger/sl"
	"CabinetREST/internal/storage/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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
	tokenService := jwtservice.NewTokenService(cfg.JWT.Secret, cfg.JWT.TTL)

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

	/* Публичный маршрут входа.
	   Здесь OptionalJWT и JWT не нужны, потому что пользователь получает токен именно через этот endpoint. */
	router.Post(handlers.Url_login,
		auth.NewLogin(log, storage, tokenService),
	)

	/* Защищённый маршрут.
	   JWT обязателен. Если заголовка Authorization нет или токен некорректный, middleware вернёт: 401 Unauthorized
	   Handler auth.NewMe будет вызван только после успешной проверки токена. */
	router.With(appmiddleware.JWT(tokenService)).Get(
		handlers.Url_Me,
		auth.NewMe(log, storage),
	)

	/* Группа публичных маршрутов с необязательной JWT-аутентификацией.
	   OptionalJWT выполняется перед каждым маршрутом внутри этой группы.
	   Если токена нет:
	   - запрос продолжает выполняться;
	   - пользователь считается анонимным;
	   - claims в context отсутствуют.

	   Если токен есть и он корректный:
	   - JWT проверяется;
	   - claims сохраняются в context;
	   - handler может определить текущего пользователя.

	   Если токен есть, но он некорректный:
	   - запрос продолжает выполняться как анонимный;
	   - claims в context отсутствуют. */

	router.Group(func(r chi.Router) {
		r.Use(appmiddleware.OptionalJWT(tokenService))
		r.Route(handlers.Url_Territories, func(r chi.Router) {
			r.Get("/search", handlers.NewTerritorySearch(log, storage))
			r.Get("/path", handlers.NewTerritoryPath(log, storage))
			r.Get("/{"+handlers.Url_Territories_ID+":[0-9]+}", handlers.NewTerritoryChildren(log, storage))
		})
		r.Route(handlers.Url_Seasons, func(r chi.Router) { r.Get("/", handlers.NewSeasons(log, storage)) })
		r.Route(handlers.Url_Sports, func(r chi.Router) { r.Get("/", handlers.NewSports(log, storage)) })
		r.Route(handlers.Url_Teams, func(r chi.Router) { r.Get("/", handlers.NewTeams(log, storage)) })
		r.Route(handlers.Url_Team_Matches, func(r chi.Router) { r.Get("/", handlers.NewTeamMatches(log, storage)) })
		r.Route(handlers.Url_Tournament, func(r chi.Router) { r.Get("/", handlers.NewTournament(log, storage)) })
		r.Route(handlers.Url_Team, func(r chi.Router) { r.Get("/", handlers.NewTeam(log, storage)) })

		r.Get(handlers.Url_OpponentOptions, handlers.NewOpponentOptions(log, storage))
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
