package main

import (
	"github.com/joho/godotenv"
	"github.com/kaa-it/go-devops/internal/agent"
)

func main() {
	_ = godotenv.Load()

	config := agent.NewConfig()

	metricsAgent := agent.New(config)

	metricsAgent.Run()
}
