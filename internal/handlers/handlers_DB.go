package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/helpers"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v4"
	"io"
	"net/http"
)

func ShortenURLuseDB(w http.ResponseWriter, r *http.Request, cfg *config.Config, conn *pgx.Conn) {
	id := helpers.GenerateID(6)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверка наличия URL в базе данных
	var count int
	err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM urls WHERE id = $1", id).Scan(&count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count > 0 {
		fmt.Println("По этому адресу уже зарегистрирован другой адрес")
		http.Error(w, "Conflict", http.StatusConflict)
		return
	}

	// Сохранение URL в базе данных
	_, err = conn.Exec(context.Background(), "INSERT INTO urls (id, url) VALUES ($1, $2)", id, string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	// Проверка наличия URL в базе данных
	var count int
	err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM urls WHERE id = $1", id).Scan(&count)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return "", fmt.Errorf("failed to check URL existence: %w", err)
	}
	if count > 0 {
		fmt.Println("По этому адресу уже зарегистрирован другой адрес")
		http.Error(w, "Conflict", http.StatusConflict)
		return "", fmt.Errorf("URL already exists")
	}

	// Сохранение URL в базе данных
	_, err = conn.Exec(context.Background(), "INSERT INTO urls (id, url) VALUES ($1, $2)", id, req.URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return "", fmt.Errorf("failed to save URL to database: %w", err)
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
