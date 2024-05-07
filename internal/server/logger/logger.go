package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Logger struct {
	log *zap.SugaredLogger
}

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

func (l *Logger) Info(args ...interface{}) {
	l.log.Info(args...)
}

func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	l.log.Infow(msg, keysAndValues)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.log.Fatal(args)
}

func (l *Logger) Error(args ...interface{}) {
	l.log.Error(args)
}
