package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgent(t *testing.T) {
	mux := http.NewServeMux()

	var metricCounter int

	mux.HandleFunc("/updates/", func(w http.ResponseWriter, r *http.Request) {
		metricCounter++
	})

	server := httptest.NewServer(mux)

	defer server.Close()

	config, err := NewConfig()
	require.NoError(t, err)

	config.Server.Address = strings.Split(server.URL, "//")[1]

	client := resty.NewWithClient(server.Client())

	agent, err := New(config)
	require.NoError(t, err)

	agent.SetRESTClient(client)

	agent.poll()

	agent.reportREST()

	assert.Equal(t, 1, metricCounter)
}
