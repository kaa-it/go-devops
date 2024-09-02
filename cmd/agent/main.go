package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/kaa-it/go-devops/internal/agent"
	"github.com/kaa-it/go-devops/internal/buildconfig"
)

func main() {
	buildconfig.PrintBuildInfo()

	_ = godotenv.Load()

	config, err := agent.NewConfig()
	if err != nil {
		log.Printf("failed to load config: %s", err)
		return
	}

	metricsAgent, err := agent.New(config)
	if err != nil {
		log.Printf("failed to create agent: %s", err)
		return
	}

	metricsAgent.Run()
}
