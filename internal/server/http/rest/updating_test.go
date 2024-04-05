package rest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type fakeUpdateService struct {
	mock.Mock
}

func (s *fakeUpdateService) UpdateGauge(name string, value float64) {
	_ = s.Called(name, value)
}

func (s *fakeUpdateService) UpdateCounter(name string, value int64) {
	_ = s.Called(name, value)
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
			s.On("UpdateGauge", mock.Anything, mock.Anything).Return()
			s.On("UpdateCounter", mock.Anything, mock.Anything).Return()

			h := NewUpdatingHandler(s)

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
