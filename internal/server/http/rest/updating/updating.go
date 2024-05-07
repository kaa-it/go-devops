package updating

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kaa-it/go-devops/internal/api"
	"github.com/kaa-it/go-devops/internal/gzip"
	"github.com/kaa-it/go-devops/internal/server/updating"
)

type Logger interface {
	RequestLogger(h http.HandlerFunc) http.HandlerFunc
	Error(args ...interface{})
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

	mux.Post("/", h.l.RequestLogger(gzip.Middleware(h.updateJSON)))
	mux.Post("/{category}/{name}/{value}", h.l.RequestLogger(h.update))

	return mux
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")

	if category != "gauge" && category != "counter" {
		h.l.Error("Metric type is not supported")
		http.Error(w, "Metric type is not supported", http.StatusNotImplemented)
		return
	}

	name := chi.URLParam(r, "name")

	valueStr := chi.URLParam(r, "value")

	switch category {
	case "gauge":
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			h.l.Error("Invalid metric value")
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}

		if err := h.a.UpdateGauge(name, value); err != nil {
			h.l.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "counter":
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}

		if err := h.a.UpdateCounter(name, value); err != nil {
			h.l.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) updateJSON(w http.ResponseWriter, r *http.Request) {
	var req api.Metrics

	dec := json.NewDecoder(r.Body)
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.l.Error(fmt.Sprintf("failed to close body: %v", err))
		}
	}()

	if err := dec.Decode(&req); err != nil {
		h.l.Error(fmt.Sprintf("failed decoding body for update: %v", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch req.MType {
	case api.GaugeType:
		if req.Value == nil {
			h.l.Error("metric value not found")
			http.Error(w, "Metric value not found", http.StatusBadRequest)
			return
		}

		if err := h.a.UpdateGauge(req.ID, *req.Value); err != nil {
			h.l.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		val, err := h.a.Gauge(req.ID)
		if err != nil {
			h.l.Error(fmt.Sprintf("failed to get %s metric: %v", req.ID, err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		*req.Value = val

	case api.CounterType:
		if req.Delta == nil {
			h.l.Error("metric value not found")
			http.Error(w, "Metric value not found", http.StatusBadRequest)
			return
		}

		if err := h.a.UpdateCounter(req.ID, *req.Delta); err != nil {
			h.l.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		val, err := h.a.Counter(req.ID)
		if err != nil {
			h.l.Error(fmt.Sprintf("failed to get %s metric: %v", req.ID, err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		*req.Delta = val

	default:
		h.l.Error(fmt.Sprintf("metric type %s not supported", req.MType))
		http.Error(w, "Metric type is not supported", http.StatusNotImplemented)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(req); err != nil {
		h.l.Error(fmt.Sprintf("failed encoding body for update: %v", err))
		return
	}
}
