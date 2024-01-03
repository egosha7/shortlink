package loger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"time"
)

// SetupLogger настраивает и возвращает новый экземпляр логгера.
func SetupLogger() (*zap.Logger, error) {
	loggerConfig := zap.Config{
		Encoding:         "console",
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "time",
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		},
	}

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	return logger, nil
}

// LogMiddleware - это промежуточное ПО для логирования HTTP-запросов.
func LogMiddleware(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			duration := time.Since(start)
			logger.Info(
				"Request received",
				zap.String("URI", r.RequestURI),
				zap.String("Method", r.Method),
				zap.Duration("Duration", duration),
			)

			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)

			logger.Info(
				"Request completed",
				zap.Int("Status", rw.Status()),
				zap.Int("Size", rw.Size()),
				zap.Duration("Duration", duration),
			)
		},
	)
}

// responseWriter - это обертка над http.ResponseWriter, позволяющая отслеживать статус и размер ответа.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// newResponseWriter создает новый экземпляр responseWriter.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		size:           0,
	}
}

// WriteHeader устанавливает статус код.
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write перехватывает запись данных и отслеживает размер ответа.
func (rw *responseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.size += size
	return size, err
}

// Status возвращает статус код ответа.
func (rw *responseWriter) Status() int {
	return rw.statusCode
}

// Size возвращает размер ответа.
func (rw *responseWriter) Size() int {
	return rw.size
}
