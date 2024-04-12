package viewing

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/kaa-it/go-devops/internal/api"
	"github.com/kaa-it/go-devops/internal/gzip"
	"github.com/kaa-it/go-devops/internal/server/viewing"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type Logger interface {
	RequestLogger(h http.HandlerFunc) http.HandlerFunc
	Error(args ...interface{})
}

type Handler struct {
	a viewing.Service
	l Logger
}

func NewHandler(a viewing.Service, l Logger) *Handler {
	return &Handler{a, l}
}

func (h *Handler) Route() *chi.Mux {
	mux := chi.NewRouter()

	mux.Get("/", h.l.RequestLogger(gzip.Middleware(h.home)))
	mux.Post("/value/", h.l.RequestLogger(gzip.Middleware(h.valueJSON)))
	mux.Get("/value/{category}/{name}", h.l.RequestLogger(h.value))

	return mux
}

func (h *Handler) home(w http.ResponseWriter, _ *http.Request) {
	const templ = `
		<table style='
			border-collapse: collapse;
			border: 2px solid rgb(140 140 140);
		'>
        	<caption style='font-weight: bold; padding: 10px;'>
				Metrics
			</caption>
            <thead style='background-color: rgb(228 240 245);'>
                <tr style='text-align: left'>
                    <th style='
						border: 1px solid rgb(160 160 160);
						padding: 8px 10px;
						text-align: left'
                    >
                        Name
                    </th>
                    <th style='
						border: 1px solid rgb(160 160 160);
						padding: 8px 10px;
						text-align: left'
                    >
                        Value
                    </th>
                </tr>
            </thead>
            <tbody>
            {{range .Gauges}}
              <tr>
					<th style='
						border: 1px solid rgb(160 160 160);
						padding: 8px 10px;
						text-align: left'
					>
						{{.Name}}
					</th>
                    <td style='
						border: 1px solid rgb(160 160 160);
						padding: 8px 10px;
						text-align: left'
					>
						{{ printf "%.3f" .Value }}
					</td>
              </tr>
            {{end}}
            {{range .Counters}}
              <tr>
					<th style='
						border: 1px solid rgb(160 160 160);
						padding: 8px 10px;
						text-align: left'
					>
						{{.Name}}
					</th>
                    <td style='
						border: 1px solid rgb(160 160 160);
						padding: 8px 10px;
						text-align: left'
					>
						{{.Value}}
					</td>
              </tr>
            {{end}}
            </tbody>
       </table>
	`
	t := template.Must(template.New("metrics").Parse(templ))

	gauges, err := h.a.Gauges()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	counters, err := h.a.Counters()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	data := struct {
		Gauges   []viewing.Gauge
		Counters []viewing.Counter
	}{
		Gauges:   gauges,
		Counters: counters,
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	if err := t.Execute(w, data); err != nil {
		log.Println(err)
	}
}

func (h *Handler) value(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")

	if category != "gauge" && category != "counter" {
		http.Error(w, "Metric type is not supported", http.StatusNotImplemented)
		return
	}

	name := chi.URLParam(r, "name")

	switch category {
	case "gauge":
		value, err := h.a.Gauge(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		str := strconv.FormatFloat(value, 'f', -1, 64)

		w.Header().Set("Content-Type", "text/plain;charset=utf-8")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(str)); err != nil {
			log.Println(err)
		}
	case "counter":
		value, err := h.a.Counter(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		str := strconv.FormatInt(value, 10)

		w.Header().Set("Content-Type", "text/plain;charset=utf-8")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(str)); err != nil {
			log.Println(err)
		}
	}
}

func (h *Handler) valueJSON(w http.ResponseWriter, r *http.Request) {
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
		value, err := h.a.Gauge(req.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		req.Value = &value

	case api.CounterType:
		value, err := h.a.Counter(req.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		req.Delta = &value

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
