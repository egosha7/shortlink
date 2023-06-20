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
	"strings"
)

func ShortenURLuseDB(w http.ResponseWriter, r *http.Request, cfg *config.Config, repo storage.URLRepository) {
	id := helpers.GenerateID(6)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Отправлена ссылка:", string(body))
	repo.PrintAllURLs()
	var existingID string
	var switchBool bool
	existingID, switchBool = repo.AddURL(id, string(body))
	if existingID != "" && switchBool == false {
		existingID = strings.TrimRight(existingID, "\n")
		shortURLout := fmt.Sprintf("%s/%s", cfg.BaseURL, existingID)
		fmt.Println("По этому адресу уже зарегистрирован другой адрес:")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(shortURLout))
		return
	} else if existingID != "" && switchBool == true {
		id = existingID
	}
	shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, id)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}

func HandleShortenURLuseDB(w http.ResponseWriter, r *http.Request, cfg *config.Config, repo storage.URLRepository) (string, error) {
	var req ShortenURLRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return "", fmt.Errorf("failed to decode request body: %w", err)
	}

	id := helpers.GenerateID(6)
	fmt.Println("Отправлена ссылка:", req.URL)

	var existingID string
	var switchBool bool
	existingID, switchBool = repo.AddURL(id, req.URL)
	if existingID != "" && switchBool == false {
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
		return "", fmt.Errorf("failed to save URL to database")
	} else if existingID != "" && switchBool == false {
		id = existingID
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

func RedirectURLuseDB(w http.ResponseWriter, r *http.Request, repo storage.URLRepository) {
	id := chi.URLParam(r, "id")

	url, ok := repo.GetURLByID(id)
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func HandleShortenBatchUseDB(w http.ResponseWriter, r *http.Request, cfg *config.Config, repo storage.URLRepository) {
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

	// Выполняем вставку каждой записи
	for _, record := range records {
		correlationID := record["correlation_id"]
		originalURL := record["original_url"]
		shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, correlationID)

		_, ok := repo.AddURL(correlationID, originalURL)
		if !ok {
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

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}
