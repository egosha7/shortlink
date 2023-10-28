package example

import (
	"bytes"
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/handlers"
	"github.com/egosha7/shortlink/internal/loger"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
)

func ExampleHandleShortenURL() {
	// Создаем необходимые зависимости и экземпляры для вашего теста
	cfg := &config.Config{
		Addr:     "localhost:8080",
		BaseURL:  "http://localhost:8080",
		FilePath: "tmp\\some3.json",
		DataBase: "",
	}
	conn := &pgx.Conn{}
	pool := &pgxpool.Pool{}

	logger, err := loger.SetupLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating logger: %v\n", err)
		os.Exit(1)
	}

	store := storage.NewURLStore(cfg.FilePath, cfg.DataBase, conn, logger, pool)

	// Создаем тестовый запрос с телом JSON-объекта, представляющего URL
	requestData := `{"url": "http://example.com"}`
	req, err := http.NewRequest("POST", "/", bytes.NewBufferString(requestData))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	req.AddCookie(&http.Cookie{
		Name:  "userID",
		Value: "testuser",
	})

	// Создаем тестовый ответ
	rr := httptest.NewRecorder()

	// Создаем маршрутизатор chi
	r := chi.NewRouter()

	// Регистрируем обработчик
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleShortenURL(w, r, cfg.BaseURL, store)
	})

	// Вызываем функцию-обработчик
	r.ServeHTTP(rr, req)

	// Проверяем код ответа
	if status := rr.Code; status != http.StatusCreated {
		fmt.Println("Unexpected status code:", status)
		return
	}

	actual := strings.TrimSpace(rr.Body.String())
	expected := `{"result":"http://localhost:8080/"`
	if !strings.HasPrefix(actual, expected) {
		fmt.Printf("Unexpected body: got %s, want prefix %s\n", actual, expected)
	}
}
