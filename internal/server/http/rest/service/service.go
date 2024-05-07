package service

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kaa-it/go-devops/internal/server/service"
)

type Logger interface {
	RequestLogger(h http.HandlerFunc) http.HandlerFunc
	Error(args ...interface{})
}

type Handler struct {
	s service.Service
	l Logger
}

func NewHandler(s service.Service, l Logger) *Handler {
	return &Handler{s, l}
}

func (h *Handler) Route() *chi.Mux {
	mux := chi.NewRouter()

	mux.Get("/", h.l.RequestLogger(h.ping))

	return mux
}

func (h *Handler) ping(w http.ResponseWriter, r *http.Request) {
	err := h.s.Ping()
	if err != nil {
		h.l.Error(fmt.Sprintf("ping failed: %v", err))
		http.Error(w, "Ping failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
