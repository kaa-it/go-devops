package main

import (
	"net/http"

	"github.com/joho/godotenv"
	"github.com/kaa-it/go-devops/internal/agent"
)

func main() {
	_ = godotenv.Load()

	config := agent.NewConfig()
	client := &http.Client{}

	metricsAgent := agent.New(client, config)

	metricsAgent.Run()
}
