package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgent(t *testing.T) {
	mux := http.NewServeMux()

	var metricCounter int

	mux.HandleFunc("/update/{category}/{name}/{value}", func(w http.ResponseWriter, r *http.Request) {
		metricCounter++
	})

	server := httptest.NewServer(mux)

	defer server.Close()

	config := NewConfig()
	config.Server.Address = strings.Split(server.URL, "//")[1]

	agent := New(server.Client(), config)

	agent.poll()

	agent.report()

	totalMetrics := agent.storage.TotalGauges() + agent.storage.TotalCounters()

	assert.Equal(t, totalMetrics, metricCounter)
}
