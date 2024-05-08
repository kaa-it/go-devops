package main

import (
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/kaa-it/go-devops/internal/agent"
	"time"
)

func main() {
	_ = godotenv.Load()

	config := agent.NewConfig()
	client := resty.New()

	client.SetRetryCount(3)
	client.SetRetryWaitTime(time.Second)
	client.SetRetryMaxWaitTime(5 * time.Second)
	client.SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
		return 2 * time.Second, nil
	})

	metricsAgent := agent.New(client, config)

	metricsAgent.Run()
}
