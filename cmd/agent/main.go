package main

import (
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/kaa-it/go-devops/internal/agent"
)

func main() {
	_ = godotenv.Load()

	config := agent.NewConfig()
	client := resty.New()

	metricsAgent := agent.New(client, config)

	metricsAgent.Run()
}
