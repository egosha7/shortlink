package worker

import "github.com/egosha7/shortlink/internal/storage"

type Worker struct {
	urlsChan chan deleteRequest
	store    *storage.URLStore
}

type deleteRequest struct {
	urls   []string
	userID string
}

func NewWorker(store *storage.URLStore) *Worker {
	// Инициализация канала
	urlsChan := make(chan deleteRequest)

	// Запуск горутины для обработки запросов на удаление
	go processDeleteRequests(urlsChan, store)

	return &Worker{
		urlsChan: urlsChan,
		store:    store,
	}
}

func (w *Worker) DeleteURLs(urls []string, userID string) {
	// Создаем deleteRequest и отправляем его в канал
	req := deleteRequest{
		urls:   urls,
		userID: userID,
	}
	w.urlsChan <- req
}

func processDeleteRequests(urlsChan <-chan deleteRequest, store *storage.URLStore) {
	for req := range urlsChan {
		// Выполняем операции с ссылками, например, вызываем метод DeleteURLs
		store.DeleteURLs(req.urls, req.userID)
	}
}
