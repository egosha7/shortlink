package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/db"
	"github.com/egosha7/shortlink/internal/helpers"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v4"
	"io"
	"net/http"
	"os"
)

func ShortenURLuseDB(w http.ResponseWriter, r *http.Request, cfg *config.Config, conn *pgx.Conn) {
	id := helpers.GenerateID(6)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Отправлена ссылка:", string(body))
	err = db.PrintAllURLs(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating table: %v\n", err)
		os.Exit(1)
	}
	stmt := `INSERT INTO urls (id, url) VALUES ($1, $2) ON CONFLICT (url) DO NOTHING RETURNING id`
	row := conn.QueryRow(context.Background(), stmt, id, string(body))

	var existingID string
	err = row.Scan(&existingID)
	if err != nil {
		if existingID == "" {
			err = conn.QueryRow(
				context.Background(), "SELECT id FROM urls WHERE url = $1", string(body),
			).Scan(&existingID)
			if err == nil {
				shortURLout := fmt.Sprintf("%s/%s", cfg.BaseURL, existingID)
				fmt.Println("По этому адресу уже зарегистрирован другой адрес:", existingID)
				http.Error(w, shortURLout, http.StatusConflict)
			} else {
				fmt.Println("Ошибка:", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, id)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}

func HandleShortenURLuseDB(w http.ResponseWriter, r *http.Request, cfg *config.Config, conn *pgx.Conn) (string, error) {

	var req ShortenURLRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return "", fmt.Errorf("failed to decode request body: %w", err)
	}

	id := helpers.GenerateID(6)
	fmt.Println("Отправлена ссылка:", req.URL)
	err = db.PrintAllURLs(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating table: %v\n", err)
		os.Exit(1)
	}
	// Сохранение URL в базе данных
	stmt := `INSERT INTO urls (id, url) VALUES ($1, $2) ON CONFLICT (url) DO NOTHING RETURNING id`
	row := conn.QueryRow(context.Background(), stmt, id, req.URL)

	var existingID string
	err = row.Scan(&existingID)
	if err != nil {
		if existingID == "" {
			err = conn.QueryRow(
				context.Background(), "SELECT id FROM urls WHERE url = $1", req.URL,
			).Scan(&existingID)
			if err == nil {

				fmt.Println("По этому адресу уже зарегистрирован другой адрес:", existingID)

				response := struct {
					Result string `json:"result"`
				}{
					Result: existingID,
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)

				err = json.NewEncoder(w).Encode(response)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return "", fmt.Errorf("failed to encode response: %w", err)
				}
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return "", fmt.Errorf("failed to save URL to database: %w", err)
		}
	}

	shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := struct {
		Result string `json:"result"`
	}{
		Result: shortURL,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return "", fmt.Errorf("failed to encode response: %w", err)
	}

	return shortURL, nil
}

func RedirectURLuseDB(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
	id := chi.URLParam(r, "id")

	var url string
	err := conn.QueryRow(context.Background(), "SELECT url FROM urls WHERE id = $1", id).Scan(&url)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func HandleShortenBatchUseDB(w http.ResponseWriter, r *http.Request, cfg *config.Config, conn *pgx.Conn) {
	var records []map[string]string
	err := json.NewDecoder(r.Body).Decode(&records)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Проверяем, что есть записи для обработки
	if len(records) == 0 {
		http.Error(w, "Empty batch", http.StatusBadRequest)
		return
	}

	// Создаем слайс для хранения результата
	res := make([]map[string]string, 0, len(records))

	// Начинаем транзакцию
	tx, err := conn.Begin(context.Background())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Выполняем вставку каждой записи
	for _, record := range records {
		correlationID := record["correlation_id"]
		originalURL := record["original_url"]
		shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, correlationID)

		_, err := tx.Exec(
			context.Background(), "INSERT INTO urls (id, URL) VALUES ($1, $2)", correlationID, originalURL,
		)
		if err != nil {
			// Ошибка вставки, откатываем транзакцию
			tx.Rollback(context.Background())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Добавляем результат в ответ
		res = append(
			res, map[string]string{
				"correlation_id": correlationID,
				"short_url":      shortURL,
			},
		)
	}

	// Завершаем транзакцию
	err = tx.Commit(context.Background())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}
