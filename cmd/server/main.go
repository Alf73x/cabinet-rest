package main

import (
	"context"
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

/*
main()
  │
  ├── config.MustLoad()
  │      │
  │      └── читаем конфигурацию
  │          env
  │          address
  │          SQLite path
  │          JWT secret
  │          JWT TTL
  │          HTTP timeouts
  │
  ├── setupLogger()
  │      │
  │      └── создаём slog.Logger
  │
  ├── sqlite.New()
  │      │
  │      └── открываем SQLite
  │
  ├── NewTokenService()
  │      │
  │      └── создаём сервис JWT
  │
  ├── chi.NewRouter()
  │
  ├── подключаем middleware
  │
  ├── регистрируем routes
  │
  ├── создаём http.Server
  │
  └── ListenAndServe()
         │
         ▼
      REST готов
*/

// Возможные окружения приложения. От окружения зависит, в частности, формат и уровень логирования.
const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	/*
		Загружаем конфигурацию приложения.

		config.MustLoad() читает настройки приложения и завершает программу при критической ошибке конфигурации.

		В cfg находятся, например:
		- адрес HTTP-сервера;
		- путь к SQLite;
		- настройки JWT;
		- таймауты HTTP;
		- текущее окружение.
	*/
	cfg := config.MustLoad()
	// fmt.Println(cfg)

	/*
		Инициализируем slog.Logger.

		Формат и минимальный уровень логирования зависят от cfg.Env.
	*/
	log := setupLogger(cfg.Env)
	// Записываем информационное сообщение о запуске REST-сервера. Вместе с сообщением добавляем структурированный атрибут env.
	log.Info("starting Cabinet REST-Server", slog.String("env", cfg.Env))
	// Debug-сообщение будет выведено только в том случае, если текущая конфигурация логгера разрешает уровень DEBUG.
	log.Debug("debug messages are enabled")

	/*
		Инициализируем хранилище SQLite.

		sqlite.New открывает базу данных и возвращает объект storage,
		который затем передаётся HTTP handlers.
	*/
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	/*
		Создаём сервис JWT.

		Он будет использоваться:
		- при login для создания токена;
		- в middleware для проверки токена.
	*/
	tokenService := jwtservice.NewTokenService(cfg.JWT.Secret, cfg.JWT.TTL)

	/*
		Создаём основной HTTP router на базе chi.
	*/
	router := chi.NewRouter()
	/*
		CORS middleware.

		Разрешаем React-клиенту обращаться к REST API с указанных адресов.

		Это особенно важно при локальной разработке, когда React работает на :5173, а REST — на другом порту.
	*/
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://192.168.7.149:5173",
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
		AllowCredentials: true,
		MaxAge:           300, // Сколько секунд браузер может кэшировать результат preflight OPTIONS-запроса.
	}))

	/*
		Стандартные middleware chi.

		RequestID: добавляет уникальный ID HTTP-запроса.

		RealIP: определяет реальный IP клиента, учитывая proxy headers.

		mwLogger.New(log): наш собственный middleware логирования HTTP-запросов.

		Recoverer: перехватывает panic внутри handlers, чтобы сервер не завершился полностью.

		URLFormat: поддержка URL format suffix в chi.
	*/
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	/*
		Публичный маршрут входа.

		Здесь JWT не требуется, потому что именно этот endpoint проверяет login/password и выдаёт JWT-токен.
	*/
	router.Post(handlers.Url_login, auth.NewLogin(log, storage, tokenService))

	/*
		Защищённый маршрут.

		appmiddleware.JWT(tokenService, storage) выполняется до вызова handler.

		Если Authorization отсутствует
		или JWT некорректен:
			-> возвращается 401 Unauthorized.

		Если JWT корректен:
			-> claims сохраняются в context;
			-> handler NewMe получает доступ к текущему пользователю.
	*/
	router.With(appmiddleware.JWT(tokenService, storage)).Get(handlers.Url_Me, auth.NewMe(log, storage))
	/* Пример ещё одного защищённого маршрута.
	router.With(appmiddleware.JWT(tokenService, storage)).Get(
		handlers.Url_Comparison,
		handlers.NewComparison(log, storage),
	)
	*/

	/*
		Группа публичных маршрутов с необязательной JWT-аутентификацией.

		OptionalJWT выполняется перед каждым маршрутом внутри этой группы.

		Если токена нет:
			- запрос продолжает выполняться;
			- пользователь считается анонимным;
			- claims в context отсутствуют.

		Если токен есть и он корректный:
			- подпись JWT проверяется;
			- claims сохраняются в context;
			- handler может определить пользователя.

		Если токен есть, но он некорректный:
			- запрос всё равно продолжается;
			- пользователь считается анонимным.
	*/

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
		r.Get(handlers.Url_Comparison, handlers.NewComparison(log, storage))
		r.Get(handlers.Url_ComparisonMatches, handlers.NewComparisonMatches(log, storage))

		r.Get(handlers.Url_SummaryCategories, handlers.NewSummaryCategories(log, storage))
		r.Get(handlers.Url_SummaryTables, handlers.NewSummaryTable(log, storage))

		r.Get(handlers.Url_SeasonInfo, handlers.NewSeasonInfo(log, storage))
		r.Get(handlers.Url_TeamInfo, handlers.NewTeamInfo(log, storage))

	})

	/*
		HTTP router полностью настроен. Теперь запускаем сервер.
	*/
	log.Info("starting server", slog.String("address", cfg.Address))

	/*
		Создаём канал для получения сигналов ОС.

		Программа будет ждать:
		- Ctrl+C;
		- SIGINT;
		- SIGTERM.

		SIGTERM особенно важен для Docker: docker stop сначала отправляет контейнеру SIGTERM.
	*/
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	/*
		Создаём HTTP server.

		Handler = router означает, что все HTTP-запросы будут передаваться в chi router.
	*/
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	/*
		Запускаем HTTP-сервер в отдельной goroutine.

		Это нужно потому, что ListenAndServe блокирует поток. Основная goroutine ниже должна продолжить работу и ждать сигнал завершения приложения.
	*/

	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			log.Error(
				"failed to start server",
				sl.Err(err),
			)
		}
	}()

	/*
		Блокируем main goroutine до тех пор, пока ОС не пришлёт сигнал завершения.
	*/
	<-done
	log.Info("stopping server")

	/*
		Graceful shutdown.

		Даём серверу до 10 секунд, чтобы завершить текущие HTTP-запросы и корректно остановиться.
	*/
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	// TODO: close storage

	log.Info("server stopped")
}

/*
setupLogger создаёт slog.Logger в зависимости от окружения приложения.

local:

	человекочитаемый текстовый формат, уровень DEBUG и выше.

dev:

	JSON-формат, уровень DEBUG и выше.

prod:

	JSON-формат, уровень INFO и выше.
*/
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
