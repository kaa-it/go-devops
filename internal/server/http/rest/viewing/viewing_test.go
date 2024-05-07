package viewing

import (
	gzipLib "compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/kaa-it/go-devops/internal/gzip"
	"github.com/kaa-it/go-devops/internal/server/viewing"
	"github.com/stretchr/testify/require"

	"github.com/go-chi/chi/v5"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type fakeViewService struct {
	mock.Mock
}

func (s *fakeViewService) Gauge(name string) (float64, error) {
	args := s.Called(name)
	return args.Get(0).(float64), args.Error(1)
}

func (s *fakeViewService) Counter(name string) (int64, error) {
	args := s.Called(name)
	return args.Get(0).(int64), args.Error(1)
}

func (s *fakeViewService) Gauges() ([]viewing.Gauge, error) {
	args := s.Called()
	return args.Get(0).([]viewing.Gauge), args.Error(1)
}

func (s *fakeViewService) Counters() ([]viewing.Counter, error) {
	args := s.Called()
	return args.Get(0).([]viewing.Counter), args.Error(1)
}

type fakeLogger struct {
	mock.Mock
	h http.HandlerFunc
}

func (l *fakeLogger) RequestLogger(h http.HandlerFunc) http.HandlerFunc {
	args := l.Called(h)
	return args.Get(0).(func(w http.ResponseWriter, r *http.Request))
}

func (l *fakeLogger) Error(args ...interface{}) {
	_ = l.Called(args...)
}

func TestViewHandler(t *testing.T) {
	type want struct {
		code     int
		response string
	}

	tests := []struct {
		name              string
		metricType        string
		metricName        string
		metricValue       string
		checkServiceError bool
		want              want
	}{
		{
			name:        "success counter case",
			metricType:  "counter",
			metricName:  "test",
			metricValue: "45",
			want: want{
				code:     http.StatusOK,
				response: "45",
			},
		},
		{
			name:        "success gauge case",
			metricType:  "gauge",
			metricName:  "test",
			metricValue: "4.5",
			want: want{
				code:     http.StatusOK,
				response: "4.5",
			},
		},
		{
			name:       "metric type is not supported",
			metricType: "test",
			metricName: "test",
			want: want{
				code:     http.StatusNotImplemented,
				response: "Metric type is not supported\n",
			},
		},
		{
			name:              "gauge metric not found",
			metricType:        "gauge",
			metricName:        "test2",
			checkServiceError: true,
			want: want{
				code:     http.StatusNotFound,
				response: "gauge not found\n",
			},
		},
		{
			name:              "counter metric not found",
			metricType:        "counter",
			metricName:        "test2",
			checkServiceError: true,
			want: want{
				code:     http.StatusNotFound,
				response: "counter not found\n",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &fakeViewService{}

			switch test.metricType {
			case "gauge":
				if test.checkServiceError {
					s.On("Gauge", test.metricName).Return(
						float64(0),
						errors.New("gauge not found"),
					)
				} else {
					value, _ := strconv.ParseFloat(test.metricValue, 64)
					s.On("Gauge", test.metricName).Return(value, nil)
				}
			case "counter":
				if test.checkServiceError {
					s.On("Counter", test.metricName).Return(
						int64(0),
						errors.New("counter not found"),
					)
				} else {
					value, _ := strconv.ParseInt(test.metricValue, 10, 64)
					s.On("Counter", test.metricName).Return(value, nil)
				}
			}

			l := &fakeLogger{}
			l.On("RequestLogger", mock.Anything).Return(func(w http.ResponseWriter, r *http.Request) {
				l.h.ServeHTTP(w, r)
			})

			h := NewHandler(s, l)

			l.h = h.value

			r := chi.NewRouter()
			r.Mount("/", h.Route())

			srv := httptest.NewServer(r)

			defer srv.Close()

			url := fmt.Sprintf("%s/value/%s/%s", srv.URL, test.metricType, test.metricName)

			req := resty.New().R()
			req.Method = http.MethodGet
			req.URL = url

			resp, err := req.Send()

			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, test.want.response, string(resp.Body()))

			l.AssertNumberOfCalls(t, "RequestLogger", 3)

			switch test.metricType {
			case "gauge":
				s.AssertCalled(t, "Gauge", test.metricName)
				s.AssertNumberOfCalls(t, "Gauge", 1)
			case "counter":
				s.AssertCalled(t, "Counter", test.metricName)
				s.AssertNumberOfCalls(t, "Counter", 1)
			}
		})
	}
}

func TestViewJSONHandler(t *testing.T) {
	type want struct {
		code     int
		response string
	}

	tests := []struct {
		name              string
		body              string
		metricType        string
		metricName        string
		metricValue       string
		checkServiceError bool
		want              want
	}{
		{
			name:        "success counter case",
			body:        `{"id": "test", "type": "counter"}`,
			metricType:  "counter",
			metricName:  "test",
			metricValue: "45",
			want: want{
				code:     http.StatusOK,
				response: `{"id": "test", "type": "counter", "delta": 45 }`,
			},
		},
		{
			name:        "success gauge case",
			body:        `{"id": "test", "type": "gauge"}`,
			metricType:  "gauge",
			metricName:  "test",
			metricValue: "4.5",
			want: want{
				code:     http.StatusOK,
				response: `{"id": "test", "type": "gauge", "value": 4.5 }`,
			},
		},
		{
			name:              "metric type is not supported",
			body:              `{"id": "test", "type": "test"}`,
			metricType:        "test",
			metricName:        "test",
			checkServiceError: true,
			want: want{
				code:     http.StatusNotImplemented,
				response: "Metric type is not supported\n",
			},
		},
		{
			name:              "gauge metric not found",
			body:              `{"id": "test2", "type": "gauge"}`,
			metricType:        "gauge",
			metricName:        "test2",
			checkServiceError: true,
			want: want{
				code:     http.StatusNotFound,
				response: "gauge not found\n",
			},
		},
		{
			name:              "counter metric not found",
			body:              `{"id": "test2", "type": "counter"}`,
			metricType:        "counter",
			metricName:        "test2",
			checkServiceError: true,
			want: want{
				code:     http.StatusNotFound,
				response: "counter not found\n",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &fakeViewService{}

			switch test.metricType {
			case "gauge":
				if test.checkServiceError {
					s.On("Gauge", test.metricName).Return(
						float64(0),
						errors.New("gauge not found"),
					)
				} else {
					value, _ := strconv.ParseFloat(test.metricValue, 64)
					s.On("Gauge", test.metricName).Return(value, nil)
				}
			case "counter":
				if test.checkServiceError {
					s.On("Counter", test.metricName).Return(
						int64(0),
						errors.New("counter not found"),
					)
				} else {
					value, _ := strconv.ParseInt(test.metricValue, 10, 64)
					s.On("Counter", test.metricName).Return(value, nil)
				}
			}

			l := &fakeLogger{}
			l.On("RequestLogger", mock.Anything).Return(func(w http.ResponseWriter, r *http.Request) {
				l.h.ServeHTTP(w, r)
			})
			l.On("Error", mock.Anything).Return()

			h := NewHandler(s, l)

			l.h = h.valueJSON

			r := chi.NewRouter()
			r.Mount("/", h.Route())

			srv := httptest.NewServer(r)

			defer srv.Close()

			url := fmt.Sprintf("%s/value/", srv.URL)

			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL = url
			req.SetHeader("Content-Type", "application/json")

			req.SetBody(test.body)

			resp, err := req.Send()

			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, test.want.code, resp.StatusCode())

			if !test.checkServiceError {
				assert.JSONEq(t, test.want.response, string(resp.Body()))
			} else {
				assert.Equal(t, test.want.response, string(resp.Body()))
			}

			l.AssertNumberOfCalls(t, "RequestLogger", 3)

			switch test.metricType {
			case "gauge":
				s.AssertCalled(t, "Gauge", test.metricName)
				s.AssertNumberOfCalls(t, "Gauge", 1)
			case "counter":
				s.AssertCalled(t, "Counter", test.metricName)
				s.AssertNumberOfCalls(t, "Counter", 1)
			}
		})
	}
}

func TestViewJSONGzip(t *testing.T) {
	t.Run("test view gauge gzip", func(t *testing.T) {
		body := `{"id": "test", "type": "gauge"}`
		response := `{"id": "test", "type": "gauge", "value": 45.2}`

		s := &fakeViewService{}

		s.On("Gauge", "test").Return(45.2, nil)

		l := &fakeLogger{}
		l.On("RequestLogger", mock.Anything).Return(func(w http.ResponseWriter, r *http.Request) {
			l.h.ServeHTTP(w, r)
		})
		l.On("Error", mock.Anything).Return()

		h := NewHandler(s, l)

		l.h = gzip.Middleware(h.valueJSON)

		r := chi.NewRouter()
		r.Mount("/", h.Route())

		srv := httptest.NewServer(r)

		defer srv.Close()

		url := fmt.Sprintf("%s/value/", srv.URL)

		req := resty.New().R()
		req.Method = http.MethodPost
		req.URL = url
		req.SetHeader("Content-Type", "application/json")
		req.SetHeader("Accept-Encoding", "gzip")
		req.SetDoNotParseResponse(true)

		req.SetBody(body)

		resp, err := req.Send()

		assert.NoError(t, err, "error making HTTP request")
		assert.Equal(t, http.StatusOK, resp.StatusCode())

		defer resp.RawBody().Close()

		zr, err := gzipLib.NewReader(resp.RawBody())
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		assert.JSONEq(t, response, string(b))

		l.AssertNumberOfCalls(t, "RequestLogger", 3)

		s.AssertCalled(t, "Gauge", "test")
		s.AssertNumberOfCalls(t, "Gauge", 1)
	})
}
