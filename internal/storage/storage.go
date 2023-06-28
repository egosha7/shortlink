package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/egosha7/shortlink/internal/helpers"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v4/stdlib"
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
	ID     string
	URL    string
	UserID string
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

func (s *URLStore) DeleteURLs(url string, userID string) error {
	if s.DBstring != "" {
		repo := NewPostgresURLRepository(s.db, s.logger)
		return repo.DeleteURLs(url, userID, s.DBstring)
	}
	s.logger.Error("Database string no exist")
	return nil
}

func (s *URLStore) AddURL(id, url, userID string) (string, bool) {
	if s.DBstring != "" {
		repo := NewPostgresURLRepository(s.db, s.logger)
		return repo.AddURL(id, url, userID, s.DBstring)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверка наличия дубликата ID
	for _, u := range s.urls {
		if u.ID == id {
			// ID уже существует в хранилище, генерируем новый
			newID := helpers.GenerateID(6)
			return s.AddURL(newID, url, userID)
		}
	}

	// Проверка наличия дубликата URL
	for _, u := range s.urls {
		if u.URL == url {
			// URL уже существует в хранилище, возвращаем соответствующий ID
			return u.ID, false
		}
	}

	newURL := URL{ID: id, URL: url, UserID: userID}
	s.urls = append(s.urls, newURL)

	// Сохранение данных в файл
	err := s.SaveToFile()
	if err != nil {
		s.logger.Error("Error saving data to file", zap.Error(err))
	}

	return id, true
}

func (s *URLStore) AddURLwithTx(records []map[string]string, ctx context.Context, BaseURL string, userID string) ([]map[string]string, bool) {
	if s.DBstring != "" {
		repo := NewPostgresURLRepository(s.db, s.logger)
		return repo.AddURLwithTx(records, ctx, s.DBstring, BaseURL, userID)
	}
	s.logger.Error("Database string no exist")
	return nil, false
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

func (s *URLStore) GetURLsByUserID(userID string) []URL {
	if s.DBstring != "" {
		repo := NewPostgresURLRepository(s.db, s.logger)
		return repo.GetURLsByUserID(userID)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	userURLs := make([]URL, 0)

	for _, u := range s.urls {
		if u.UserID == userID {
			userURLs = append(userURLs, u)
		}
	}

	return userURLs
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

func (r *PostgresURLRepository) DeleteURLs(url string, userID string, DataBase string) error {
	// Создание конфигурации пула подключений
	config, err := pgxpool.ParseConfig(DataBase)
	if err != nil {
		fmt.Printf(err.Error())
	}

	// Создание пула подключений
	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		fmt.Printf(err.Error())
	}

	// Использование пула подключений для выполнения запросов
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		fmt.Printf(err.Error())
	}
	defer conn.Release()

	query := `
		UPDATE user_urls
		SET delFLAG = true
		WHERE userID = $1 AND IDshortURL = $2
	`

	// Выполняем запрос на множественное обновление
	_, err = conn.Exec(context.Background(), query, userID, url)
	if err != nil {
		fmt.Printf(err.Error())
		return err
	}
	return nil
}

func (r *PostgresURLRepository) AddURL(id string, url string, userID string, DataBase string) (string, bool) {
	return r.addURLWithRetry(id, url, userID, 10, DataBase)
}

func (r *PostgresURLRepository) addURLWithRetry(id string, url string, userID string, attempts int, DataBase string) (string, bool) {
	// Создание конфигурации пула подключений
	config, err := pgxpool.ParseConfig(DataBase)
	if err != nil {
		fmt.Printf(err.Error())
	}

	// Создание пула подключений
	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		fmt.Printf(err.Error())
	}

	// Использование пула подключений для выполнения запросов
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		fmt.Printf(err.Error())
	}
	defer conn.Release()

	query := "INSERT INTO urls (id, url) VALUES ($1, $2)"
	_, err = conn.Exec(context.Background(), query, id, url)
	if err != nil {
		pgErr, ok := err.(*pgconn.PgError)
		if ok && pgErr.Code == pgerrcode.UniqueViolation {
			switch pgErr.ConstraintName {
			case "urls_pkey":
				// ID уже существует в базе данных, генерируем новый
				if attempts > 0 {
					newID := helpers.GenerateID(6)
					return r.addURLWithRetry(newID, url, userID, attempts-1, DataBase)
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

	// Добавляем данные в таблицу user_urls
	userQuery := "INSERT INTO user_urls (idshorturl, userid) VALUES ($1, $2)"
	_, userErr := conn.Exec(context.Background(), userQuery, id, userID)
	if userErr != nil {
		r.logger.Error("Failed to add user URL", zap.Error(userErr))
		return "", false
	}

	return id, true
}

func (r *PostgresURLRepository) AddURLwithTx(records []map[string]string, ctx context.Context, DBString string, BaseURL string, userID string) ([]map[string]string, bool) {

	conn, err := sql.Open("pgx", DBString)
	if err != nil {
		r.logger.Error("Error sql.Open", zap.Error(err))
		return nil, false
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Error BeginTx", zap.Error(err))
		return nil, false
	}
	defer tx.Rollback()

	res := make([]map[string]string, 0, len(records))

	// Обрабатываем каждую запись
	for _, record := range records {
		correlationID := record["correlation_id"]
		originalURL := record["original_url"]

		_, err = tx.Exec("INSERT INTO urls (id, url) VALUES ($1, $2)", correlationID, originalURL)
		if err != nil {
			r.logger.Error("Error Exec", zap.Error(err))
			return nil, false
		}

		_, err = tx.Exec("INSERT INTO user_urls (idshorturl, userid) VALUES ($1, $2)", correlationID, userID)
		if err != nil {
			r.logger.Error("Error Exec", zap.Error(err))
			return nil, false
		}

		shortURL := fmt.Sprintf("%s/%s", BaseURL, correlationID)

		// Добавляем результат в ответ
		res = append(
			res, map[string]string{
				"correlation_id": correlationID,
				"short_url":      shortURL,
			},
		)
	}

	err = tx.Commit()
	if err != nil {
		r.logger.Error("Error commit", zap.Error(err))
		return nil, false
	}
	return res, true
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

	var delFlag bool
	query = "SELECT delFLAG FROM user_urls WHERE IDshortURL = $1"
	err = r.db.QueryRow(context.Background(), query, id).Scan(&delFlag)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", false
		}
		r.logger.Error("Failed to get delFLAG by IDshortURL", zap.Error(err))
		return "", false
	}

	if delFlag {
		return url, false
	}

	return url, true
}

func (r *PostgresURLRepository) GetURLsByUserID(userID string) []URL {
	var userURLs []URL
	query := `
        SELECT u.URL, uu.IDshortURL
        FROM urls u
        JOIN user_urls uu ON u.ID = uu.IDshortURL
        WHERE uu.userID = $1
    `
	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		r.logger.Error("Failed to get URLs by UserID", zap.Error(err))
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var url, shortURL string
		err := rows.Scan(&url, &shortURL)
		if err != nil {
			r.logger.Error("Failed to scan URL and ShortURL", zap.Error(err))
			return nil
		}
		userURLs = append(userURLs, URL{ID: shortURL, URL: url, UserID: userID})
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error occurred while iterating over rows", zap.Error(err))
		return nil
	}

	return userURLs
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

	_, err = r.db.Exec(
		context.Background(), `
		CREATE TABLE IF NOT EXISTS user_urls (
			ID SERIAL PRIMARY KEY,
			IDshortURL TEXT,
			userID TEXT,
			delFLAG BOOL DEFAULT false
		)
	`,
	)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		context.Background(), `
		ALTER TABLE user_urls
		ADD CONSTRAINT fk_name_IDshortURL
		FOREIGN KEY (IDshortURL) REFERENCES urls (ID);

	`,
	)
	if err != nil {
		return err
	}

	return nil
}
