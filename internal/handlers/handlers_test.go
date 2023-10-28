package handlers_test

import (
	"bytes"
	"fmt"
	"github.com/egosha7/shortlink/internal/config"
	"github.com/egosha7/shortlink/internal/helpers"
	"github.com/egosha7/shortlink/internal/loger"
	"github.com/egosha7/shortlink/internal/storage"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/egosha7/shortlink/internal/handlers"
	"github.com/go-chi/chi"
)

func TestShortenURL(t *testing.T) {
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
		fmt.Fprintf(os.Stderr, "Error creating loger: %v\n", err)
		os.Exit(1)
	}

	// Указываем экземпляр URLStore
	store := storage.NewURLStore(cfg.FilePath, cfg.DataBase, conn, logger, pool)

	// Создаем тестовый запрос
	body := []byte("http://example.com")
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	// Создаем тестовый ответ
	rr := httptest.NewRecorder()

	// Создаем маршрутизатор chi
	r := chi.NewRouter()

	// Регистрируем обработчик
	r.HandleFunc(
		`/`, func(w http.ResponseWriter, r *http.Request) {
			handlers.ShortenURL(w, r, cfg.BaseURL, store, logger)
		},
	)

	// Вызываем функцию-обработчи
	r.ServeHTTP(rr, req)

	// Проверяем код ответа
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf(
			"handler returned wrong status code: got %v want %v",
			status, http.StatusCreated,
		)
	}

	// Проверяем содержимое ответа
	expected := "http://localhost:8080/"
	if rr.Body.String()[:len(expected)] != expected {
		t.Errorf(
			"handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected,
		)
	}
}

func TestRedirectURL(t *testing.T) {
	cfg := &config.Config{
		Addr:     "localhost:8080",
		BaseURL:  "http://localhost:8080",
		FilePath: "tmp\\some3.json",
		DataBase: "",
	}

	pool := &pgxpool.Pool{}
	conn := &pgx.Conn{}

	logger, err := loger.SetupLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating loger: %v\n", err)
		os.Exit(1)
	}

	// Указываем экземпляр URLStore
	store := storage.NewURLStore(cfg.FilePath, cfg.DataBase, conn, logger, pool)

	link := "http://example.com"
	formData := strings.NewReader(link)

	req, err := http.NewRequest("POST", "/", formData)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "text-plain")
	rr := httptest.NewRecorder()

	// Создаем маршрутизатор chi
	r := chi.NewRouter()

	// Регистрируем обработчик
	r.Post(
		"/", func(w http.ResponseWriter, r *http.Request) {
			handlers.ShortenURL(w, r, cfg.BaseURL, store, logger)
		},
	)

	// Вызываем функцию-обработчик
	r.ServeHTTP(rr, req)

	// Проверяем код ответа
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf(
			"handler returned wrong status code: got %v want %v",
			status, http.StatusCreated,
		)
	}

	req, err = http.NewRequest("GET", "/"+rr.Body.String()[len(rr.Body.String())-6:], nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()

	// Создаем маршрутизатор chi
	r2 := chi.NewRouter()

	// Регистрируем обработчик для GET-запросов на маршруте /{id}
	r2.Get(
		"/{id}", func(w http.ResponseWriter, r *http.Request) {
			handlers.RedirectURL(w, r, store)
		},
	)

	// Вызываем функцию-обработчик
	r2.ServeHTTP(rr, req)

	// Проверяем код ответа
	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf(
			"handler returned wrong status code: got %v want %v",
			status, http.StatusTemporaryRedirect,
		)
	}

	expected := link
	if rr.Header().Get("Location") != expected {
		t.Errorf(
			"handler returned unexpected location: got %v want %v",
			rr.Header().Get("Location"), expected,
		)
	}
}

func TestHandleShortenURLRequest(t *testing.T) {
	// Создайте необходимые зависимости и экземпляры для вашего теста
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
		fmt.Fprintf(os.Stderr, "Error creating loger: %v\n", err)
		os.Exit(1)
	}

	store := storage.NewURLStore(cfg.FilePath, cfg.DataBase, conn, logger, pool)

	// Создайте тестовый запрос с телом JSON-объекта, представляющего URL
	requestData := `{"url": "http://example.com"}`
	req, err := http.NewRequest("POST", "/", bytes.NewBufferString(requestData))
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{
		Name:  "userID",
		Value: "testuser",
	})

	// Создайте тестовый ответ
	rr := httptest.NewRecorder()

	// Создайте маршрутизатор chi
	r := chi.NewRouter()

	// Регистрируйте обработчик
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleShortenURL(w, r, cfg.BaseURL, store)
	})

	// Вызовите функцию-обработчик
	r.ServeHTTP(rr, req)

	// Проверяем код ответа
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf(
			"handler returned wrong status code: got %v want %v",
			status, http.StatusCreated,
		)
	}

	actual := strings.TrimSpace(rr.Body.String())
	expected := `{"result":"http://localhost:8080/` + helpers.LastGeneratedID + `"}`
	if actual != expected {
		t.Errorf(
			"handler returned unexpected body: got %s, want %s",
			actual, expected,
		)
	}
}
