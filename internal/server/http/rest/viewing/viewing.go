package viewing

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kaa-it/go-devops/internal/server/viewing"
)

type Logger interface {
	RequestLogger(h http.HandlerFunc) http.HandlerFunc
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

	mux.Get("/", h.l.RequestLogger(h.home))
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
		//str := fmt.Sprintf("%f", value)

		w.Header().Set("Content-Type", "text/plain;charset=utf-8")
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(str)), 10))
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
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(str)), 10))
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(str)); err != nil {
			log.Println(err)
		}
	}
}
