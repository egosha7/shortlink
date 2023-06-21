package storage

import (
	"context"
	"encoding/json"
	"github.com/egosha7/shortlink/internal/helpers"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"os"
	"sync"
)

type URLStore struct {
	urls     []URL
	mu       sync.RWMutex
	filePath string
	DBstring string
	db       *pgx.Conn
	logger   *zap.Logger
}

type URL struct {
	ID  string
	URL string
}

func NewURLStore(filePath string, DBstring string, db *pgx.Conn, logger *zap.Logger) *URLStore {
	return &URLStore{
		urls:     make([]URL, 0),
		filePath: filePath,
		DBstring: DBstring,
		db:       db,
		logger:   logger,
	}
}

func (s *URLStore) AddURL(id, url string) (string, bool) {
	if s.DBstring != "" {
		repo := NewPostgresURLRepository(s.db, s.logger)
		return repo.AddURL(id, url)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверка наличия дубликата ID
	for _, u := range s.urls {
		if u.ID == id {
			// ID уже существует в хранилище, генерируем новый
			newID := helpers.GenerateID(6)
			return s.AddURL(newID, url)
		}
	}

	// Проверка наличия дубликата URL
	for _, u := range s.urls {
		if u.URL == url {
			// URL уже существует в хранилище, возвращаем соответствующий ID
			return u.ID, false
		}
	}

	newURL := URL{ID: id, URL: url}
	s.urls = append(s.urls, newURL)

	// Сохранение данных в файл
	err := s.SaveToFile()
	if err != nil {
		s.logger.Error("Error saving data to file", zap.Error(err))
	}

	return id, true
}

func (s *URLStore) GetURL(id string) (string, bool) {
	if s.DBstring != "" {
		repo := NewPostgresURLRepository(s.db, s.logger)
		return repo.GetURLByID(id)
	}

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
	CreateTable()
	PrintAllURLs()
}

type PostgresURLRepository struct {
	db     *pgx.Conn
	logger *zap.Logger
}

func NewPostgresURLRepository(db *pgx.Conn, logger *zap.Logger) *PostgresURLRepository {
	return &PostgresURLRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresURLRepository) AddURL(id string, url string) (string, bool) {
	return r.addURLWithRetry(id, url, 10)
}

func (r *PostgresURLRepository) addURLWithRetry(id string, url string, attempts int) (string, bool) {
	query := "INSERT INTO urls (id, url) VALUES ($1, $2)"
	_, err := r.db.Exec(context.Background(), query, id, url)
	if err != nil {
		pgErr, ok := err.(*pgconn.PgError)
		if ok && pgErr.Code == pgerrcode.UniqueViolation {
			switch pgErr.ConstraintName {
			case "urls_pkey":
				// ID уже существует в базе данных, генерируем новый
				if attempts > 0 {
					newID := helpers.GenerateID(6)
					return r.addURLWithRetry(newID, url, attempts-1)
				} else {
					r.logger.Warn("Exceeded maximum retry attempts")
				}
			case "urls_url_key":
				// URL уже существует в базе данных, возвращаем соответствующий ID
				urlInDB, ok := r.GetIDByURL(url)
				if !ok {
					r.logger.Error("Failed to get ID by URL", zap.Error(err))
					return "", false
				}
				return urlInDB, false
			default:
				r.logger.Error("Failed to add URL", zap.Error(err))
			}
		} else {
			r.logger.Error("Failed to add URL", zap.Error(err))
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
		r.logger.Error("Failed to get ID by URL", zap.Error(err))
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
		r.logger.Error("Failed to get URL by ID", zap.Error(err))
		return "", false
	}
	return url, true
}

func (r *PostgresURLRepository) PrintAllURLs() {
	rows, err := r.db.Query(context.Background(), "SELECT id, url FROM urls")
	if err != nil {
		r.logger.Error("Failed to query URLs", zap.Error(err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, url string
		err := rows.Scan(&id, &url)
		if err != nil {
			r.logger.Error("Failed to scan row", zap.Error(err))
			continue
		}
		r.logger.Info(
			"URL",
			zap.String("ID", id),
			zap.String("URL", url),
		)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating over rows", zap.Error(err))
	}
}

func (r *PostgresURLRepository) CreateTable() error {
	_, err := r.db.Exec(
		context.Background(), `
		CREATE TABLE IF NOT EXISTS urls (
			ID TEXT PRIMARY KEY,
			URL TEXT,
			UNIQUE (URL)
		)
	`,
	)
	if err != nil {
		return err
	}

	return nil
}
