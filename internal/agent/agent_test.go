package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/stretchr/testify/assert"
)

func TestAgent(t *testing.T) {
	mux := http.NewServeMux()

	var metricCounter int

	mux.HandleFunc("/updates/", func(w http.ResponseWriter, r *http.Request) {
		metricCounter++
	})

	server := httptest.NewServer(mux)

	defer server.Close()

	config := NewConfig()
	config.Server.Address = strings.Split(server.URL, "//")[1]

	client := resty.NewWithClient(server.Client())

	agent := New(client, config)

	agent.poll()

	agent.report()

	assert.Equal(t, 1, metricCounter)
}
