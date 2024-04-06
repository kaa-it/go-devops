package updating

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kaa-it/go-devops/internal/server/updating"
)

type Logger interface {
	RequestLogger(h http.HandlerFunc) http.HandlerFunc
}

type Handler struct {
	a updating.Service
	l Logger
}

func NewHandler(a updating.Service, l Logger) *Handler {
	return &Handler{a, l}
}

func (h *Handler) Route() *chi.Mux {
	mux := chi.NewRouter()

	mux.Post("/{category}/{name}/{value}", h.l.RequestLogger(h.update))

	return mux
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")

	if category != "gauge" && category != "counter" {
		http.Error(w, "Metric type is not supported", http.StatusNotImplemented)
		return
	}

	name := chi.URLParam(r, "name")

	valueStr := chi.URLParam(r, "value")

	switch category {
	case "gauge":
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}

		h.a.UpdateGauge(name, value)
	case "counter":
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}
		h.a.UpdateCounter(name, value)
	}

	w.WriteHeader(http.StatusOK)
}
