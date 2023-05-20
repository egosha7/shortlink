package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"egosha7/shortlink/internal/handlers"
)

func TestShortenURL(t *testing.T) {
	// Создаем тестовый запрос
	body := []byte("http://example.com")
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	// Создаем тестовый ответ
	rr := httptest.NewRecorder()

	// Вызываем функцию-обработчик
	handler := http.HandlerFunc(handlers.ShortenURL)
	handler.ServeHTTP(rr, req)

	// Проверяем код ответа
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	// Проверяем содержимое ответа
	expected := "http://localhost:8080/"
	if rr.Body.String()[:len(expected)] != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestRedirectURL(t *testing.T) {
	link := "https://example.com"
	formData := strings.NewReader(link)

	req, err := http.NewRequest("POST", "/", formData)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "text-plain")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.ShortenURL)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	req, err = http.NewRequest("GET", "/"+rr.Body.String()[len(rr.Body.String())-6:], nil)
	if err != nil {
		t.Fatal(err)
	}
	handler = http.HandlerFunc(handlers.RedirectURL)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusTemporaryRedirect)
	}

	expected := link
	if rr.Header().Get("Location") != expected {
		t.Errorf("handler returned unexpected location: got %v want %v",
			rr.Header().Get("Location"), expected)
	}
}
