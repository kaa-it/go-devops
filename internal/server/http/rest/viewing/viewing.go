// Package viewing describes handlers for getting metrics from server
package viewing

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/kaa-it/go-devops/internal/server/trusted"

	"github.com/go-chi/chi/v5"

	"github.com/kaa-it/go-devops/internal/api"
	"github.com/kaa-it/go-devops/internal/gzip"
	"github.com/kaa-it/go-devops/internal/server/viewing"
)

type Logger interface {
	RequestLogger(h http.HandlerFunc) http.HandlerFunc
	Error(args ...interface{})
}

// Handler describes common state for all handlers in package
type Handler struct {
	a viewing.Service
	l Logger
}

// NewHandler creates new instance of Handler
func NewHandler(a viewing.Service, l Logger) *Handler {
	return &Handler{a, l}
}

// Route creates router for all routes controlled by the package
func (h *Handler) Route(trustedSubnet string) *chi.Mux {
	mux := chi.NewRouter()

	mux.Get("/", h.l.RequestLogger(trusted.Middleware(trustedSubnet, gzip.Middleware(h.home))))
	mux.Post("/value/", h.l.RequestLogger(trusted.Middleware(trustedSubnet, gzip.Middleware(h.valueJSON))))
	mux.Get("/value/{category}/{name}", h.l.RequestLogger(trusted.Middleware(trustedSubnet, h.value)))

	return mux
}

//		    @Tags	View
//			@Summary Request to get HTML page with all metrics
//			@Produce    html
//		    @Success	200
//		    @Failure	500
//	        @Router	    /	[get]
func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
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

	ctx := r.Context()

	gauges, err := h.a.Gauges(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	counters, err := h.a.Counters(ctx)
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

//			    @Tags	View
//				@Summary Request to get value of metric by its category and name
//				@Produce    plain
//		        @Param	    category   path       api.MetricsType  true "Metric type"
//	            @Param      name       path       string  true "Metric name"
//				@Success	200
//				@Failure    404        {string}   string
//				@Failure	501        {string}   string "Metric type is not supported"
//			    @Failure    500
//				@Router	    /value/{category}/{name}	[get]
func (h *Handler) value(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")

	if category != "gauge" && category != "counter" {
		h.l.Error(fmt.Sprintf("metric type %s is not supported", category))
		http.Error(w, "Metric type is not supported", http.StatusNotImplemented)
		return
	}

	name := chi.URLParam(r, "name")

	ctx := r.Context()

	switch category {
	case "gauge":
		value, err := h.a.Gauge(ctx, name)
		if err != nil {
			h.l.Error(fmt.Sprintf("failed to get gauge with name %s: %v", name, err))
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
		value, err := h.a.Counter(ctx, name)
		if err != nil {
			h.l.Error(fmt.Sprintf("failed to get counter with name %s: %v", name, err))
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

// MetricRequest describes body format for request metric value.
type MetricRequest struct {
	// ID - unique metric name.
	ID string `json:"id"`
	// MType - metric type.
	MType api.MetricsType `json:"type"`
}

//			@Tags	View
//			@Summary Request to get metric value in JSON format
//		    @Accept     json
//			@Produce    json
//			@Param	    request    body       MetricRequest  true "Metric value request"
//			@Success	200
//	        @Failure    400        {string}   string
//			@Failure    404        {string}   string
//			@Failure	501        {string}   string "Metric type is not supported"
//			@Failure    500
//			@Router	    /value/	[get]
func (h *Handler) valueJSON(w http.ResponseWriter, r *http.Request) {
	var req MetricRequest

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

	res := api.Metrics{
		ID:    req.ID,
		MType: req.MType,
	}

	switch req.MType {
	case api.GaugeType:
		value, err := h.a.Gauge(ctx, req.ID)
		if err != nil {
			h.l.Error(fmt.Sprintf("failed to get gauge with ID %s: %v", req.ID, err))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		res.Value = &value

	case api.CounterType:
		value, err := h.a.Counter(ctx, req.ID)
		if err != nil {
			h.l.Error(fmt.Sprintf("failed to get counter with ID %s: %v", req.ID, err))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		res.Delta = &value

	default:
		h.l.Error(fmt.Sprintf("metric type %s not supported", req.MType))
		http.Error(w, "Metric type is not supported", http.StatusNotImplemented)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(res); err != nil {
		h.l.Error(fmt.Sprintf("failed encoding body for update: %v", err))
		return
	}
}
