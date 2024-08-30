package main

import (
	"log"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"

	"github.com/kaa-it/go-devops/internal/agent"
	"github.com/kaa-it/go-devops/internal/buildconfig"
)

const (
	_retryCount       = 3
	_retryWaitTime    = time.Second
	_retryMaxWaitTime = 5 * time.Second
	_retryDelay       = 2 * time.Second
)

func main() {
	buildconfig.PrintBuildInfo()

	_ = godotenv.Load()

	config, err := agent.NewConfig()
	if err != nil {
		log.Printf("failed to load config: %s", err)
		return
	}

	client := resty.New()

	client.SetRetryCount(_retryCount)
	client.SetRetryWaitTime(_retryWaitTime)
	client.SetRetryMaxWaitTime(_retryMaxWaitTime)
	client.SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
		return _retryDelay, nil
	})

	metricsAgent, err := agent.New(client, config)
	if err != nil {
		log.Printf("failed to create agent: %s", err)
		return
	}

	metricsAgent.Run()
}
