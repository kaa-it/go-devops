// Package logger contains implementation of logger that uses zap package.
package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Logger describes logger instance.
type Logger struct {
	log *zap.SugaredLogger
}

// New creates new logger instance with given minimal level.
func New(level string) (*Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()

	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{log: zl.Sugar()}, nil
}

// RequestLogger wraps request handler for logging response.
func (l *Logger) RequestLogger(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggerResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		uri := r.RequestURI

		method := r.Method

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		l.log.Infoln(
			"uri", uri,
			"method", method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
}

// Info sends values into log at info level.
func (l *Logger) Info(args ...interface{}) {
	l.log.Info(args...)
}

// Infow sends values with message into log at info level.
func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	l.log.Infow(msg, keysAndValues)
}

// Fatal sends message into log and terminates application.
func (l *Logger) Fatal(args ...interface{}) {
	l.log.Fatal(args)
}

// Error sends values into log at error level.
func (l *Logger) Error(args ...interface{}) {
	l.log.Error(args)
}
