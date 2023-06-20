package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/egosha7/shortlink/internal/helpers"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"os"
	"sync"
)

type URLStore struct {
	urls     []URL
	mu       sync.RWMutex
	filePath string
}

type URL struct {
	ID  string
	URL string
}

func NewURLStore(filePath string) *URLStore {
	return &URLStore{
		urls:     make([]URL, 0),
		filePath: filePath,
	}
}

func (s *URLStore) AddURL(id, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	newURL := URL{ID: id, URL: url}
	s.urls = append(s.urls, newURL)

	// Сохранение данных в файл
	err := s.SaveToFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving data to file: %v\n", err)
	}
}

func (s *URLStore) GetURL(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.urls {
		if u.ID == id {
			return u.URL, true
		}
	}
	return "", false
}

func (s *URLStore) LoadFromFile() error {
	// Проверка наличия флага или переменной окружения
	if s.filePath == "" {
		return nil // Если значение не установлено, выходим без загрузки данных
	}

	// Проверка существования файла
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		// Создание нового файла
		file, err := os.Create(s.filePath)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	// Открываем файл
	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Читаем данные из файла
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	if fileInfo.Size() == 0 {
		return nil
	}

	if err = json.NewDecoder(file).Decode(&s.urls); err != nil {
		return err
	}

	return nil
}

func (s *URLStore) SaveToFile() error {
	// Проверка наличия флага или переменной окружения
	if s.filePath == "" {
		return nil // Если значение не установлено, выходим без сохранения на диск
	}

	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(s.urls)
	if err != nil {
		return err
	}

	return nil
}

type URLRepository interface {
	AddURL(id string, url string) (string, bool)
	GetIDByURL(url string) (string, bool)
	GetURLByID(id string) (string, bool)
	PrintAllURLs()
}

type PostgresURLRepository struct {
	db *pgx.Conn
}

func NewPostgresURLRepository(db *pgx.Conn) *PostgresURLRepository {
	return &PostgresURLRepository{
		db: db,
	}
}

func (r *PostgresURLRepository) AddURL(id string, url string) (string, bool) {
	query := "INSERT INTO urls (id, url) VALUES ($1, $2)"
	_, err := r.db.Exec(context.Background(), query, id, url)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "urls_pkey" {
				// ID уже существует в базе данных, генерируем новый
				newID := helpers.GenerateID(6)
				return r.AddURL(newID, url)
			} else if pgErr.ConstraintName == "urls_url_key" {
				// URL уже существует в базе данных, возвращаем соответствующий ID
				urlInDB, ok := r.GetIDByURL(url)
				if !ok {
					fmt.Println("Failed to get ID by URL:", err)
					return "", false
				}
				return urlInDB, false
			} else {
				fmt.Println("Failed to add URL:", err)
			}
		} else {
			fmt.Println("Failed to add URL:", err)
		}
		return "", false
	}
	return id, true
}

func (r *PostgresURLRepository) GetIDByURL(url string) (string, bool) {
	var id string
	query := "SELECT id FROM urls WHERE url = $1"
	err := r.db.QueryRow(context.Background(), query, url).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", false
		}
		fmt.Println("Failed to get ID by URL:", err)
		return "", false
	}
	return id, true
}

func (r *PostgresURLRepository) GetURLByID(id string) (string, bool) {
	var url string
	query := "SELECT url FROM urls WHERE id = $1"
	err := r.db.QueryRow(context.Background(), query, id).Scan(&url)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", false
		}
		fmt.Println("Failed to get URL by ID:", err)
		return "", false
	}
	return url, true
}

func (r *PostgresURLRepository) PrintAllURLs() {
	rows, err := r.db.Query(context.Background(), "SELECT id, url FROM urls")
	if err != nil {
		fmt.Println("Failed to query URLs:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, url string
		err := rows.Scan(&id, &url)
		if err != nil {
			fmt.Println("Failed to scan row:", err)
			continue
		}
		fmt.Printf("ID: %s, URL: %s\n", id, url)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("Error iterating over rows:", err)
	}
}
