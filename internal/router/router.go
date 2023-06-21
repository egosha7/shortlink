package routes

import (
	"fmt"
	"github.com/egosha7/shortlink/internal/compress"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/db"
	"github.com/egosha7/shortlink/internal/handlers"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/jackc/pgx/v4"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

func SetupRoutes(cfg *config.Config, conn *pgx.Conn) http.Handler {

	// Создание хранилища
	store := storage.NewURLStore(cfg.FilePath, cfg.DataBase, conn)
	repo := storage.NewPostgresURLRepository(conn)
	if cfg.DataBase != "" {
		repo.CreateTable()
	}

	// Загрузка данных из файла
	err := store.LoadFromFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading data from file: %v\n", err)
		os.Exit(1)
	}

	// Создание роутера
	r := chi.NewRouter()

	gzipMiddleware := compress.GzipMiddleware{}

	// Создание группы роутера
	r.Group(
		func(route chi.Router) {
			route.Use(gzipMiddleware.Apply)
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

			route.Post(
				"/", func(w http.ResponseWriter, r *http.Request) {
					handlers.ShortenURL(w, r, cfg, store)
				},
			)

			route.Post(
				"/api/shorten", func(w http.ResponseWriter, r *http.Request) {
					handlers.HandleShortenURL(w, r, cfg, store)
				},
			)

			route.Post(
				"/api/shorten/batch", func(w http.ResponseWriter, r *http.Request) {
					handlers.HandleShortenBatch(w, r, cfg, store)
				},
			)

		},
	)

	return r
}
