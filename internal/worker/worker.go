package worker

import "github.com/egosha7/shortlink/internal/storage"

// Worker представляет собой структуру для обработки запросов на удаление URL.
type Worker struct {
	urlsChan chan deleteRequest // Канал для запросов на удаление URL.
	store    *storage.URLStore  // Хранилище URL.
}

// deleteRequest представляет собой запрос на удаление URL.
type deleteRequest struct {
	urls   []string // Список URL для удаления.
	userID string   // Идентификатор пользователя, связанный с запросом на удаление URL.
}

// NewWorker создает и возвращает новый экземпляр Worker.
func NewWorker(store *storage.URLStore) *Worker {
	// Инициализация канала для запросов на удаление URL.
	urlsChan := make(chan deleteRequest)

	// Запуск горутины для обработки запросов на удаление.
	go processDeleteRequests(urlsChan, store)

	return &Worker{
		urlsChan: urlsChan,
		store:    store,
	}
}

// DeleteURLs добавляет запрос на удаление URL в канал для последующей обработки.
func (w *Worker) DeleteURLs(urls []string, userID string) {
	// Создаем deleteRequest и отправляем его в канал для обработки.
	req := deleteRequest{
		urls:   urls,
		userID: userID,
	}
	w.urlsChan <- req
}

// processDeleteRequests обрабатывает запросы на удаление URL из канала.
func processDeleteRequests(urlsChan <-chan deleteRequest, store *storage.URLStore) {
	for req := range urlsChan {
		// Выполняем операции с URL, например, вызываем метод DeleteURLs для удаления указанных URL.
		store.DeleteURLs(req.urls, req.userID)
	}
}
