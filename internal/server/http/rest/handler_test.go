package rest

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			name:             "success case",
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
			url := fmt.Sprintf("/update/%s/%s/%s", test.metricType, test.metricName, test.metricValue)
			request := httptest.NewRequest(http.MethodPost, url, nil)

			w := httptest.NewRecorder()

			s := &fakeUpdateService{}

			s.On("UpdateGauge", mock.Anything, mock.Anything).Return()

			h := NewHandler(s)

			h.Route().ServeHTTP(w, request)

			res := w.Result()

			assert.Equal(t, test.want.code, res.StatusCode)

			if test.want.response != "" {
				defer func() { _ = res.Body.Close() }()
				resBody, err := io.ReadAll(res.Body)

				require.NoError(t, err)
				assert.Equal(t, test.want.response, string(resBody))
			}

			if test.checkServiceCall {
				switch test.metricType {
				case "gauge":
					value, _ := strconv.ParseFloat(test.metricValue, 64)
					s.AssertCalled(t, "UpdateGauge", test.metricName, value)
				case "counter":
					value, _ := strconv.ParseInt(test.metricValue, 10, 64)
					s.AssertCalled(t, "UpdateCounter", test.metricName, value)
				}
			}
		})
	}
}
