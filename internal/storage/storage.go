package storage

import (
	"encoding/json"
	"fmt"
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
