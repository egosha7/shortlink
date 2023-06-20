package routes

import (
	"context"
	"fmt"
	"github.com/egosha7/shortlink/internal/compress"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/handlers"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/jackc/pgx/v4"
	"net/http"

	"github.com/go-chi/chi"
)

func SetupRoutes(cfg *config.Config, store *storage.URLStore, conn *pgx.Conn) http.Handler {
	PrintAllURLs(conn)
	// Создание роутера
	r := chi.NewRouter()
	gzipMiddleware := compress.GzipMiddleware{}

	if conn != nil {
		r.Group(
			// Группа хэндлеров для работы с БД
			func(route chi.Router) {
				route.Use(gzipMiddleware.Apply)
				route.Get(
					"/{id}", func(w http.ResponseWriter, r *http.Request) {
						handlers.RedirectURLuseDB(w, r, conn)
					},
				)

				route.Get(
					"/ping", func(w http.ResponseWriter, r *http.Request) {
						err := conn.Ping(context.Background())
						if err != nil {
							http.Error(w, "Database connection error", http.StatusInternalServerError)
							return
						}

						w.WriteHeader(http.StatusOK)
					},
				)

				route.Post(
					"/", func(w http.ResponseWriter, r *http.Request) {
						handlers.ShortenURLuseDB(w, r, cfg, conn)
					},
				)

				route.Post(
					"/api/shorten", func(w http.ResponseWriter, r *http.Request) {
						handlers.HandleShortenURLuseDB(w, r, cfg, conn)
					},
				)

				route.Post(
					"/api/shorten/batch", func(w http.ResponseWriter, r *http.Request) {
						handlers.HandleShortenBatchUseDB(w, r, cfg, conn)
					},
				)
			},
		)
	} else {
		r.Group(
			// Группа хэндлеров для работы с паматью и файлом
			func(route chi.Router) {
				route.Use(gzipMiddleware.Apply)
				route.Get(
					"/{id}", func(w http.ResponseWriter, r *http.Request) {
						handlers.RedirectURL(w, r, store)
					},
				)

				route.Get(
					"/ping", func(w http.ResponseWriter, r *http.Request) {
						err := conn.Ping(context.Background())
						if err != nil {
							http.Error(w, "Database connection error", http.StatusInternalServerError)
							return
						}

						w.WriteHeader(http.StatusOK)
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
	}

	return r
}

func PrintAllURLs(conn *pgx.Conn) error {
	rows, err := conn.Query(context.Background(), "SELECT * FROM urls")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, url string
		err := rows.Scan(&id, &url)
		if err != nil {
			return err
		}
		fmt.Printf("ID: %s, URL: %s\n", id, url)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
