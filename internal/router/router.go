package routes

import (
	"github.com/egosha7/shortlink/internal/cookiemw"
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

func SetupRoutes(cfg *config.Config, conn *pgx.Conn, logger *zap.Logger) http.Handler {

	// Создание хранилища
	store := storage.NewURLStore(cfg.FilePath, cfg.DataBase, conn, logger)
	repo := storage.NewPostgresURLRepository(conn, logger)
	if cfg.DataBase != "" {
		repo.CreateTable()
	}

	// Загрузка данных из файла
	err := store.LoadFromFile()
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

			route.Get(
				"/cookie/set", func(w http.ResponseWriter, r *http.Request) {
					handlers.SetCookieHandler(w, r)
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
					handlers.GetUserURLsHandler(w, r, cfg.BaseURL, store)
				},
			)

			route.Post(
				"/", func(w http.ResponseWriter, r *http.Request) {
					handlers.ShortenURL(w, r, cfg.BaseURL, store)
				},
			)

			route.Post(
				"/api/shorten", func(w http.ResponseWriter, r *http.Request) {
					handlers.HandleShortenURL(w, r, cfg.BaseURL, store)
				},
			)

			route.Post(
				"/api/shorten/batch", func(w http.ResponseWriter, r *http.Request) {
					handlers.HandleShortenBatch(w, r, cfg.BaseURL, store)
				},
			)

		},
	)

	return r
}
