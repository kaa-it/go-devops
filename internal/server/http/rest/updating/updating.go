package updating

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/kaa-it/go-devops/internal/api"
	"github.com/kaa-it/go-devops/internal/gzip"
	"github.com/kaa-it/go-devops/internal/server/hash"
	"github.com/kaa-it/go-devops/internal/server/updating"
)

type Logger interface {
	RequestLogger(h http.HandlerFunc) http.HandlerFunc
	Error(args ...interface{})
}

// Handler describes common state for all handlers in package
type Handler struct {
	a updating.Service
	l Logger
}

// NewHandler creates new instance of Handler
func NewHandler(a updating.Service, l Logger) *Handler {
	return &Handler{a, l}
}

// Route creates router for all routes controlled by the package
func (h *Handler) Route(key string) *chi.Mux {
	mux := chi.NewRouter()

	mux.Post("/", h.l.RequestLogger(hash.Middleware(key, gzip.Middleware(h.updateJSON))))
	mux.Post("/{category}/{name}/{value}", h.l.RequestLogger(hash.Middleware(key, h.update)))

	return mux
}

// Updates returns handler for /updates route.
func (h *Handler) Updates(key string) http.HandlerFunc {

	return h.l.RequestLogger(hash.Middleware(key, gzip.Middleware(h.updates)))
}

//			@Tags	Update
//			@Summary Request to update value of metric using URL params
//			@Param	    category   path       string  true "Metric type"
//		    @Param      name       path       string  true "Metric name"
//	        @Param      value      path       string  true "New metric value"
//			@Success	200
//			@Failure    404        {string}   string
//			@Failure	501        {string}   string "Metric type is not supported"
//			@Failure    500
//			@Router	    /update/{category}/{name}/{value}	[post]
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")

	if category != "gauge" && category != "counter" {
		h.l.Error("Metric type is not supported")
		http.Error(w, "Metric type is not supported", http.StatusNotImplemented)
		return
	}

	name := chi.URLParam(r, "name")

	valueStr := chi.URLParam(r, "value")

	ctx := r.Context()

	switch category {
	case "gauge":
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			h.l.Error("Invalid metric value")
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}

		if err := h.a.UpdateGauge(ctx, name, value); err != nil {
			h.l.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "counter":
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			h.l.Error("Invalid metric value")
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}

		if err := h.a.UpdateCounter(ctx, name, value); err != nil {
			h.l.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

//			@Tags	 Update
//			@Summary Request to update metric value in JSON format
//		    @Accept     json
//			@Produce    json
//			@Param	    request    body       api.Metrics  true "Metric update request"
//			@Success	200
//	        @Failure    400        {string}   string
//			@Failure    404        {string}   string
//			@Failure	501        {string}   string "Metric type is not supported"
//			@Failure    500
//			@Router	    /update/	[post]
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

	ctx := r.Context()

	switch req.MType {
	case api.GaugeType:
		if req.Value == nil {
			h.l.Error("metric value not found")
			http.Error(w, "Metric value not found", http.StatusBadRequest)
			return
		}

		if err := h.a.UpdateGauge(ctx, req.ID, *req.Value); err != nil {
			h.l.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		val, err := h.a.Gauge(ctx, req.ID)
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

		if err := h.a.UpdateCounter(ctx, req.ID, *req.Delta); err != nil {
			h.l.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		val, err := h.a.Counter(ctx, req.ID)
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

//			@Tags	 Update
//			@Summary Request to update some metric value simultaneously in JSON format
//		    @Accept     json
//			@Produce    json
//			@Param	    request    body       []api.Metrics  true "Batch metric update request"
//			@Success	200
//	        @Failure    400        {string}   string
//			@Failure    404        {string}   string
//			@Failure    500
//			@Router	    /updates	[post]
func (h *Handler) updates(w http.ResponseWriter, r *http.Request) {
	var req []api.Metrics

	dec := json.NewDecoder(r.Body)
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.l.Error(fmt.Sprintf("failed to close body: %v", err))
		}
	}()

	if err := dec.Decode(&req); err != nil {
		h.l.Error(fmt.Sprintf("failed decoding body for updates: %v", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req) == 0 {
		h.l.Error("metric batch is empty")
		http.Error(w, "Metric batch is empty", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	if err := h.a.Updates(ctx, req); err != nil {
		h.l.Error(fmt.Sprintf("batch update failed: %v", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
