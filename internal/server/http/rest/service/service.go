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

// Handler describes common state for all handlers in package
type Handler struct {
	s service.Service
	l Logger
}

// NewHandler creates new instance of Handler
func NewHandler(s service.Service, l Logger) *Handler {
	return &Handler{s, l}
}

// Route creates router for all routes controlled by the package
func (h *Handler) Route() *chi.Mux {
	mux := chi.NewRouter()

	mux.Get("/", h.l.RequestLogger(h.ping))

	return mux
}

//		    @Tags	Info
//			@Summary Request for service health checking
//		    @Success	200
//		    @Failure	500
//	        @Router	    /ping	[get]
func (h *Handler) ping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.s.Ping(ctx)
	if err != nil {
		h.l.Error(fmt.Sprintf("ping failed: %v", err))
		http.Error(w, "Ping failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
