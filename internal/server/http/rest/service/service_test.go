package service

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kaa-it/go-devops/internal/server/service"
)

func TestPing(t *testing.T) {
	type want struct {
		code     int
		response string
	}

	tests := []struct {
		name  string
		error bool
		want  want
	}{
		{
			name:  "success ping",
			error: false,
			want: want{
				code:     http.StatusOK,
				response: "",
			},
		},
		{
			name:  "ping failed",
			error: true,
			want: want{
				code:     http.StatusInternalServerError,
				response: "Ping failed\n",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := service.NewMockService(t)
			if test.error {
				s.On("Ping", mock.Anything).Return(
					errors.New("ping failed"),
				)
			} else {
				s.On("Ping", mock.Anything).Return(nil)
			}

			var h *Handler

			l := NewMockLogger(t)
			l.On("RequestLogger", mock.Anything).Return(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h.ping(w, r)
			}))

			if test.error {
				l.On("Error", mock.Anything).Return()
			}

			h = NewHandler(s, l)

			r := chi.NewRouter()
			r.Mount("/ping", h.Route())

			srv := httptest.NewServer(r)

			defer srv.Close()

			url := fmt.Sprintf("%s/ping", srv.URL)

			req := resty.New().R()
			req.Method = http.MethodGet
			req.URL = url

			resp, err := req.Send()

			assert.NoError(t, err, "error making HTTP request")
			s.AssertNumberOfCalls(t, "Ping", 1)
			l.AssertNumberOfCalls(t, "RequestLogger", 1)

			if test.error {
				assert.Equal(t, http.StatusInternalServerError, resp.StatusCode())
				assert.Equal(t, test.want.response, string(resp.Body()))
				l.AssertCalled(t, "Error", "ping failed: ping failed")
			} else {
				assert.Equal(t, http.StatusOK, resp.StatusCode())
			}
		})
	}
}
