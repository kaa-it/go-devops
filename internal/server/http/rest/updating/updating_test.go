package updating

import (
	"bytes"
	gzipLib "compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/kaa-it/go-devops/internal/gzip"
	"github.com/stretchr/testify/require"

	"github.com/go-chi/chi/v5"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type fakeUpdateService struct {
	mock.Mock
}

func (s *fakeUpdateService) UpdateGauge(name string, value float64) error {
	args := s.Called(name, value)
	return args.Error(0)
}

func (s *fakeUpdateService) UpdateCounter(name string, value int64) error {
	args := s.Called(name, value)
	return args.Error(0)
}

func (s *fakeUpdateService) Gauge(name string) (float64, error) {
	args := s.Called(name)
	return args.Get(0).(float64), args.Error(1)
}

func (s *fakeUpdateService) Counter(name string) (int64, error) {
	args := s.Called(name)
	return args.Get(0).(int64), args.Error(1)
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

func TestUpdateHandler(t *testing.T) {
	type want struct {
		code     int
		response string
	}

	tests := []struct {
		name             string
		metricType       string
		metricName       string
		metricValue      string
		checkServiceCall bool
		want             want
	}{
		{
			name:             "success counter case",
			metricType:       "counter",
			metricName:       "test",
			metricValue:      "45",
			checkServiceCall: true,
			want: want{
				code:     http.StatusOK,
				response: "",
			},
		},
		{
			name:             "success gauge case",
			metricType:       "gauge",
			metricName:       "test",
			metricValue:      "4.5",
			checkServiceCall: true,
			want: want{
				code:     http.StatusOK,
				response: "",
			},
		},
		{
			name:        "metric type is not supported",
			metricType:  "test",
			metricName:  "test",
			metricValue: "4.5",
			want: want{
				code:     http.StatusNotImplemented,
				response: "Metric type is not supported\n",
			},
		},
		{
			name:        "invalid gauge metric value",
			metricType:  "gauge",
			metricName:  "test",
			metricValue: "ax",
			want: want{
				code:     http.StatusBadRequest,
				response: "Invalid metric value\n",
			},
		},
		{
			name:        "invalid counter metric value",
			metricType:  "counter",
			metricName:  "test",
			metricValue: "ax",
			want: want{
				code:     http.StatusBadRequest,
				response: "Invalid metric value\n",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &fakeUpdateService{}
			s.On("UpdateGauge", mock.Anything, mock.Anything).Return(nil)
			s.On("UpdateCounter", mock.Anything, mock.Anything).Return(nil)

			l := &fakeLogger{}
			l.On("RequestLogger", mock.Anything).Return(func(w http.ResponseWriter, r *http.Request) {
				l.h.ServeHTTP(w, r)
			})

			l.On("Error", mock.Anything).Return()

			h := NewHandler(s, l)

			l.h = h.update

			r := chi.NewRouter()
			r.Mount("/update", h.Route())

			srv := httptest.NewServer(r)

			defer srv.Close()

			url := fmt.Sprintf("%s/update/%s/%s/%s", srv.URL, test.metricType, test.metricName, test.metricValue)

			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL = url

			resp, err := req.Send()

			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, test.want.code, resp.StatusCode())

			if test.want.response != "" {
				assert.Equal(t, test.want.response, string(resp.Body()))
			}

			l.AssertNumberOfCalls(t, "RequestLogger", 2)

			if test.checkServiceCall {
				switch test.metricType {
				case "gauge":
					value, _ := strconv.ParseFloat(test.metricValue, 64)
					s.AssertCalled(t, "UpdateGauge", test.metricName, value)
					s.AssertNumberOfCalls(t, "UpdateGauge", 1)
				case "counter":
					value, _ := strconv.ParseInt(test.metricValue, 10, 64)
					s.AssertCalled(t, "UpdateCounter", test.metricName, value)
					s.AssertNumberOfCalls(t, "UpdateCounter", 1)
				}
			}
		})
	}
}

func TestJSONUpdateHandler(t *testing.T) {
	type want struct {
		code     int
		response string
	}

	tests := []struct {
		name             string
		body             string
		metricName       string
		metricType       string
		startDelta       int64
		delta            int64
		value            float64
		checkServiceCall bool
		want             want
	}{
		{
			name:             "success counter case",
			body:             `{"id": "test", "type": "counter", "delta": 45}`,
			metricName:       "test",
			metricType:       "counter",
			startDelta:       int64(45),
			delta:            int64(45),
			checkServiceCall: true,
			want: want{
				code:     http.StatusOK,
				response: `{"id": "test", "type": "counter", "delta": 90}`,
			},
		},
		{
			name:             "success gauge case",
			body:             `{"id": "test", "type": "gauge", "value": 45.2}`,
			metricName:       "test",
			metricType:       "gauge",
			value:            45.2,
			checkServiceCall: true,
			want: want{
				code:     http.StatusOK,
				response: `{"id": "test", "type": "gauge", "value": 45.2}`,
			},
		},
		{
			name: "metric type is not supported",
			body: `{"id": "test", "type": "test", "delta": 45}`,
			want: want{
				code:     http.StatusNotImplemented,
				response: "Metric type is not supported\n",
			},
		},
		{
			name: "no gauge metric value",
			body: `{"id": "test", "type": "gauge", "delta": 45}`,
			want: want{
				code:     http.StatusBadRequest,
				response: "Metric value not found\n",
			},
		},
		{
			name: "no counter metric value",
			body: `{"id": "test", "type": "counter", "value": 45.2}`,
			want: want{
				code:     http.StatusBadRequest,
				response: "Metric value not found\n",
			},
		},
		{
			name: "failed to parse request",
			body: `{"id": "test", "type": "counter", "delta": "45""}`,
			want: want{
				code:     http.StatusBadRequest,
				response: "invalid character '\"' after object key:value pair\n",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &fakeUpdateService{}
			s.On("UpdateGauge", mock.Anything, mock.Anything).Return(nil)
			s.On("UpdateCounter", mock.Anything, mock.Anything).Return(nil)
			s.On("Gauge", mock.Anything, mock.Anything).Return(test.value, nil)
			s.On("Counter", mock.Anything, mock.Anything).Return(test.delta+test.startDelta, nil)

			l := &fakeLogger{}
			l.On("RequestLogger", mock.Anything).Return(func(w http.ResponseWriter, r *http.Request) {
				l.h.ServeHTTP(w, r)
			})
			l.On("Error", mock.Anything).Return()

			h := NewHandler(s, l)

			l.h = h.updateJSON

			r := chi.NewRouter()
			r.Mount("/update", h.Route())

			srv := httptest.NewServer(r)

			defer srv.Close()

			url := fmt.Sprintf("%s/update", srv.URL)

			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL = url
			req.SetHeader("Content-Type", "application/json")
			req.SetBody(test.body)

			resp, err := req.Send()

			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, test.want.code, resp.StatusCode())

			if test.want.response != "" {
				if test.checkServiceCall {
					assert.JSONEq(t, test.want.response, string(resp.Body()))
				} else {
					assert.Equal(t, test.want.response, string(resp.Body()))
				}
			}

			l.AssertNumberOfCalls(t, "RequestLogger", 2)

			if test.checkServiceCall {
				switch test.metricType {
				case "gauge":
					s.AssertCalled(t, "UpdateGauge", test.metricName, test.value)
					s.AssertNumberOfCalls(t, "UpdateGauge", 1)
					s.AssertCalled(t, "Gauge", test.metricName)
					s.AssertNumberOfCalls(t, "Gauge", 1)
				case "counter":
					s.AssertCalled(t, "UpdateCounter", test.metricName, test.delta)
					s.AssertNumberOfCalls(t, "UpdateCounter", 1)
					s.AssertCalled(t, "Counter", test.metricName)
					s.AssertNumberOfCalls(t, "Counter", 1)
				}
			}
		})
	}
}

func TestUpdateGzip(t *testing.T) {
	t.Run("send gzip metric", func(t *testing.T) {
		response := `{"id": "test", "type": "gauge", "value": 45.2}`

		s := &fakeUpdateService{}
		s.On("UpdateGauge", mock.Anything, mock.Anything).Return(nil)
		s.On("Gauge", mock.Anything, mock.Anything).Return(45.2, nil)

		l := &fakeLogger{}
		l.On("RequestLogger", mock.Anything).Return(func(w http.ResponseWriter, r *http.Request) {
			l.h.ServeHTTP(w, r)
		})
		l.On("Error", mock.Anything).Return()

		h := NewHandler(s, l)

		l.h = gzip.Middleware(h.updateJSON)

		r := chi.NewRouter()
		r.Mount("/update", h.Route())

		srv := httptest.NewServer(r)

		defer srv.Close()

		url := fmt.Sprintf("%s/update", srv.URL)

		buf := bytes.NewBuffer(nil)
		zw := gzipLib.NewWriter(buf)

		_, err := zw.Write([]byte(response))
		require.NoError(t, err)

		err = zw.Close()
		require.NoError(t, err)

		req := resty.New().R()
		req.Method = http.MethodPost
		req.URL = url
		req.SetHeader("Content-Type", "application/json")
		req.SetHeader("Content-Encoding", "gzip")
		req.SetHeader("Accept-Encoding", "")
		req.SetBody(buf)

		resp, err := req.Send()

		assert.NoError(t, err, "error making HTTP request")
		assert.Equal(t, http.StatusOK, resp.StatusCode())

		assert.JSONEq(t, response, string(resp.Body()))

		l.AssertNumberOfCalls(t, "RequestLogger", 2)

		s.AssertCalled(t, "UpdateGauge", "test", 45.2)
		s.AssertNumberOfCalls(t, "UpdateGauge", 1)
		s.AssertCalled(t, "Gauge", "test")
		s.AssertNumberOfCalls(t, "Gauge", 1)
	})

	t.Run("accept gzip metric", func(t *testing.T) {
		response := `{"id": "test", "type": "gauge", "value": 45.2}`

		s := &fakeUpdateService{}
		s.On("UpdateGauge", mock.Anything, mock.Anything).Return(nil)
		s.On("Gauge", mock.Anything, mock.Anything).Return(45.2, nil)

		l := &fakeLogger{}
		l.On("RequestLogger", mock.Anything).Return(func(w http.ResponseWriter, r *http.Request) {
			l.h.ServeHTTP(w, r)
		})
		l.On("Error", mock.Anything).Return()

		h := NewHandler(s, l)

		l.h = gzip.Middleware(h.updateJSON)

		r := chi.NewRouter()
		r.Mount("/update", h.Route())

		srv := httptest.NewServer(r)

		defer srv.Close()

		url := fmt.Sprintf("%s/update", srv.URL)

		req := resty.New().R()
		req.Method = http.MethodPost
		req.URL = url
		req.SetHeader("Content-Type", "application/json")
		req.SetHeader("Accept-Encoding", "gzip")
		req.SetBody(response)
		req.SetDoNotParseResponse(true)

		resp, err := req.Send()

		assert.NoError(t, err, "error making HTTP request")
		assert.Equal(t, http.StatusOK, resp.StatusCode())

		defer resp.RawBody().Close()

		zr, err := gzipLib.NewReader(resp.RawBody())
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		assert.JSONEq(t, response, string(b))

		l.AssertNumberOfCalls(t, "RequestLogger", 2)

		s.AssertCalled(t, "UpdateGauge", "test", 45.2)
		s.AssertNumberOfCalls(t, "UpdateGauge", 1)
		s.AssertCalled(t, "Gauge", "test")
		s.AssertNumberOfCalls(t, "Gauge", 1)
	})
}
