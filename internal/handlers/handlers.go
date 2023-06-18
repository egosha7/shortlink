package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/helpers"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/go-chi/chi"
	"io"
	"net/http"
)

type Key string

func ShortenURL(w http.ResponseWriter, r *http.Request, cfg *config.Config, store *storage.URLStore) {
	id := helpers.GenerateID(6)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url, ok := store.GetURL(id)
	if !ok {
		store.AddURL(id, string(body))
	} else {
		fmt.Println("По этому адресу уже зарегистрирован другой адрес:", url)
	}
	shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, id)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}

type ShortenURLRequest struct {
	URL string `json:"url"`
}

func HandleShortenURL(w http.ResponseWriter, r *http.Request, cfg *config.Config, store *storage.URLStore) (string, error) {

	var req ShortenURLRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return "", fmt.Errorf("failed to decode request body: %w", err)
	}

	// Используем тело запроса
	id := helpers.GenerateID(6)

	url, ok := store.GetURL(id)
	if ok {
		fmt.Println("По этому адресу уже зарегистрирован другой адрес:", url)
	} else {
		store.AddURL(id, req.URL)
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

func RedirectURL(w http.ResponseWriter, r *http.Request, store *storage.URLStore) {
	id := chi.URLParam(r, "id")
	url, ok := store.GetURL(id)
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func HandleShortenBatch(w http.ResponseWriter, r *http.Request, cfg *config.Config, store *storage.URLStore) {
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

	// Обрабатываем каждую запись
	for _, record := range records {
		correlationID := record["correlation_id"]
		originalURL := record["original_url"]

		url, ok := store.GetURL(correlationID)
		if ok {
			fmt.Println("По этому адресу уже зарегистрирован другой адрес:", url)
		} else {
			store.AddURL(correlationID, originalURL)
		}

		shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, correlationID)

		// Добавляем результат в ответ
		res = append(
			res, map[string]string{
				"correlation_id": correlationID,
				"short_url":      shortURL,
			},
		)
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}
