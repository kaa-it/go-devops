package main

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"

	"github.com/kaa-it/go-devops/internal/agent"
)

const (
	_retryCount       = 3
	_retryWaitTime    = time.Second
	_retryMaxWaitTime = 5 * time.Second
	_retryDelay       = 2 * time.Second
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build data: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	_ = godotenv.Load()

	config := agent.NewConfig()
	client := resty.New()

	client.SetRetryCount(_retryCount)
	client.SetRetryWaitTime(_retryWaitTime)
	client.SetRetryMaxWaitTime(_retryMaxWaitTime)
	client.SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
		return _retryDelay, nil
	})

	metricsAgent := agent.New(client, config)

	metricsAgent.Run()
}
