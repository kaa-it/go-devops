package rest

import (
	"net/http"
	"strconv"

	"github.com/kaa-it/go-devops/internal/server/updating"
)

type Handler struct {
	a updating.Service
}

func NewHandler(a updating.Service) *Handler {
	return &Handler{a}
}

func (h *Handler) Route() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/update/{category}/{name}/{value}", h.update)

	return mux
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	category := r.PathValue("category")

	if category != "gauge" && category != "counter" {
		http.Error(w, "Metric type is not supported", http.StatusNotImplemented)
		return
	}

	name := r.PathValue("name")

	valueStr := r.PathValue("value")

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
