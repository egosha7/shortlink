package main

import (
	"github.com/egosha7/shortlink/internal/const"
	"github.com/egosha7/shortlink/internal/handlers"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMainPageGet(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.MainPage)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := _const.Form
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestMainPagePost(t *testing.T) {
	link := "https://example.com"
	formData := strings.NewReader("link=" + link)

	req, err := http.NewRequest("POST", "/", formData)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.MainPage)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	expected := 6
	if len(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			len(rr.Body.String()), expected)
	}

}

func TestMainPageRedirect(t *testing.T) {
	link := "https://example.com"
	formData := strings.NewReader("link=" + link)

	req, err := http.NewRequest("POST", "/", formData)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.MainPage)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	req, err = http.NewRequest("GET", "/"+rr.Body.String(), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusSeeOther)
	}

	expected := link
	if rr.Header().Get("Location") != expected {
		t.Errorf("handler returned unexpected location: got %v want %v",
			rr.Header().Get("Location"), expected)
	}
}

func TestMainPageBadRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "/invalid", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.MainPage)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := "404"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
