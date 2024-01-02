package routes

import (
	"context"
	httphandlers "github.com/egosha7/shortlink/handlers"
	"github.com/egosha7/shortlink/internal/auth"
	"github.com/egosha7/shortlink/internal/cookiemw"
	"github.com/egosha7/shortlink/internal/worker"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"net/http"
	"os"

	"github.com/egosha7/shortlink/internal/compress"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/db"
	"github.com/egosha7/shortlink/internal/handlers"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v4"
)

// SetupRoutes настраивает и возвращает обработчик HTTP-маршрутов.
func SetupRoutes(cfg *config.Config, conn *pgx.Conn, logger *zap.Logger) http.Handler {
	config, err := pgxpool.ParseConfig(cfg.DataBase)
	if err != nil {
		logger.Error("Error parse config", zap.Error(err))
	}
	config.MaxConns = 1000
	// Создание пула подключений
	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		logger.Error("Error connect config", zap.Error(err))
	}
	// Создание хранилища
	store := storage.NewURLStore(cfg.FilePath, cfg.DataBase, conn, logger, pool)
	repo := storage.NewPostgresURLRepository(conn, logger, pool)
	wkr := worker.NewWorker(store)

	if cfg.DataBase != "" {
		repo.CreateTable()
	}

	// Загрузка данных из файла.
	err = store.LoadFromFile()
	if err != nil {
		logger.Error("Error loading data from file", zap.Error(err)) // Используем логер для вывода ошибки
		os.Exit(1)
	}

	// Создание роутера
	r := chi.NewRouter()

	gzipMiddleware := compress.GzipMiddleware{}

	// Создание группы роутера
	r.Group(
		func(route chi.Router) {
			route.Use(cookiemw.CookieMiddleware)
			route.Use(gzipMiddleware.Apply)

			route.Delete(
				"/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
					handlers.DeleteUserURLsHandler(w, r, wkr)
				},
			)

			route.Get(
				"/cookie/set", func(w http.ResponseWriter, r *http.Request) {
					auth.SetCookieHandler(w, r)
				},
			)

			route.Get(
				"/{id}", func(w http.ResponseWriter, r *http.Request) {
					handlers.RedirectURL(w, r, store)
				},
			)

			route.Get(
				"/ping", func(w http.ResponseWriter, r *http.Request) {
					db.PingDB(w, r, conn)
				},
			)

			route.Get(
				"/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
					handlers.GetUserURLsHandler(w, r, cfg.BaseURL, store, logger)
				},
			)

			route.Post(
				"/", func(w http.ResponseWriter, r *http.Request) {
					httphandlers.ShortenURL(w, r, cfg.BaseURL, store, logger)
				},
			)

			route.Post(
				"/api/shorten", func(w http.ResponseWriter, r *http.Request) {
					handlers.HandleShortenURL(w, r, cfg.BaseURL, store)
				},
			)

			route.Post(
				"/api/shorten/batch", func(w http.ResponseWriter, r *http.Request) {
					handlers.HandleShortenBatch(w, r, cfg.BaseURL, store, logger)
				},
			)

			route.Get(
				"/api/internal/stats", func(w http.ResponseWriter, r *http.Request) {
					handlers.StatsHandler(w, r, store, cfg.TrustedSubnet)
				},
			)

		},
	)

	return r
}
